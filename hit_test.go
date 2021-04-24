package pirsch

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHitFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar&utm_source=test&utm_medium=email&utm_campaign=newsletter&utm_content=signup&utm_term=keywords", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	req.Header.Set("Referer", "http://ref/")
	hit := FromRequest(req, "salt", &Options{
		Client:       dbClient,
		TenantID:     NewTenantID(42),
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})
	assert.True(t, hit.TenantID.Valid)
	assert.Equal(t, 42, int(hit.TenantID.Int64))
	assert.NotEmpty(t, hit.Fingerprint)
	assert.NoError(t, dbClient.SaveHits([]Hit{hit}))

	if hit.Time.IsZero() ||
		!hit.Session.Valid || hit.Session.Time.IsZero() ||
		hit.UserAgent != "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36" ||
		hit.Path != "/test/path" ||
		hit.URL != "/test/path?query=param&foo=bar&utm_source=test&utm_medium=email&utm_campaign=newsletter&utm_content=signup&utm_term=keywords" ||
		hit.Language.String != "de" ||
		hit.Referrer.String != "http://ref/" ||
		hit.OS.String != OSWindows ||
		hit.OSVersion.String != "10" ||
		hit.Browser.String != BrowserChrome ||
		hit.BrowserVersion.String != "84.0" ||
		!hit.Desktop ||
		hit.Mobile ||
		hit.ScreenWidth != 640 ||
		hit.ScreenHeight != 1024 ||
		hit.UTMSource.String != "test" ||
		hit.UTMMedium.String != "email" ||
		hit.UTMCampaign.String != "newsletter" ||
		hit.UTMContent.String != "signup" ||
		hit.UTMTerm.String != "keywords" {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestSession(t *testing.T) {
	cleanupDB()
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	req.Header.Set("Referer", "http://ref/")
	hit1 := FromRequest(req, "salt", &Options{
		Client: dbClient,
	})
	assert.False(t, hit1.TenantID.Valid)
	assert.NotEmpty(t, hit1.Fingerprint)
	assert.NoError(t, dbClient.SaveHits([]Hit{hit1}))
	hit2 := FromRequest(req, "salt", &Options{
		Client: dbClient,
	})
	assert.False(t, hit2.TenantID.Valid)
	assert.NotEmpty(t, hit2.Fingerprint)
	assert.Equal(t, hit1.Session.Time.Unix(), hit2.Session.Time.Unix())
}

func TestHitFromRequestOverwrite(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	hit := FromRequest(req, "salt", &Options{
		URL: "http://bar.foo/new/custom/path?query=param&foo=bar#anchor",
	})

	if hit.Path != "/new/custom/path" ||
		hit.URL != "http://bar.foo/new/custom/path?query=param&foo=bar#anchor" {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestOverwritePathAndReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	hit := FromRequest(req, "salt", &Options{
		URL:      "http://bar.foo/overwrite/this?query=param&foo=bar#anchor",
		Path:     "/new/custom/path",
		Referrer: "http://custom.ref/",
	})

	if hit.Path != "/new/custom/path" ||
		hit.URL != "http://bar.foo/new/custom/path?query=param&foo=bar#anchor" ||
		hit.Referrer.String != "http://custom.ref/" {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestScreenSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	hit := FromRequest(req, "salt", &Options{
		ScreenWidth:  -5,
		ScreenHeight: 400,
	})

	if hit.ScreenWidth != 0 || hit.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", hit.ScreenWidth, hit.ScreenHeight)
	}

	hit = FromRequest(req, "salt", &Options{
		ScreenWidth:  400,
		ScreenHeight: 0,
	})

	if hit.ScreenWidth != 0 || hit.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", hit.ScreenWidth, hit.ScreenHeight)
	}

	hit = FromRequest(req, "salt", &Options{
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})

	if hit.ScreenWidth != 640 || hit.ScreenHeight != 1024 {
		t.Fatalf("Screen size must be set, but was: %v %v", hit.ScreenWidth, hit.ScreenHeight)
	}
}

func TestHitFromRequestCountryCode(t *testing.T) {
	geoDB, err := NewGeoDB(GeoDBConfig{
		File: filepath.Join("geodb/GeoIP2-Country-Test.mmdb"),
	})

	if err != nil {
		t.Fatalf("Geo DB must have been loaded, but was: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	req.RemoteAddr = "81.2.69.142"
	hit := FromRequest(req, "salt", &Options{
		geoDB: geoDB,
	})

	if hit.CountryCode.String != "gb" {
		t.Fatalf("Country code for hit must have been returned, but was: %v", hit.CountryCode.String)
	}

	req = httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	req.RemoteAddr = "127.0.0.1"
	hit = FromRequest(req, "salt", &Options{
		geoDB: geoDB,
	})

	if hit.CountryCode.String != "" {
		t.Fatalf("Country code for hit must be empty, but was: %v", hit.CountryCode.String)
	}
}

func TestIgnoreHitPrefetch(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "valid")
	req.Header.Set("X-Moz", "prefetch")

	if !Ignore(req) {
		t.Fatal("Hit with X-Moz header must be ignored")
	}

	req.Header.Del("X-Moz")
	req.Header.Set("X-Purpose", "prefetch")

	if !Ignore(req) {
		t.Fatal("Hit with X-Purpose header must be ignored")
	}

	req.Header.Set("X-Purpose", "preview")

	if !Ignore(req) {
		t.Fatal("Hit with X-Purpose header must be ignored")
	}

	req.Header.Del("X-Purpose")
	req.Header.Set("Purpose", "prefetch")

	if !Ignore(req) {
		t.Fatal("Hit with Purpose header must be ignored")
	}

	req.Header.Set("Purpose", "preview")

	if !Ignore(req) {
		t.Fatal("Hit with Purpose header must be ignored")
	}

	req.Header.Del("Purpose")

	if Ignore(req) {
		t.Fatal("Hit must not be ignored")
	}
}

func TestIgnoreHitUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "This is a bot request")

	if !Ignore(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a crawler request")

	if !Ignore(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a spider request")

	if !Ignore(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Visit http://spam.com!")

	if !Ignore(req) {
		t.Fatal("Hit with URL in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Mozilla/123.0")

	if Ignore(req) {
		t.Fatal("Hit with regular User-Agent must not be ignored")
	}

	req.Header.Set("User-Agent", "")

	if !Ignore(req) {
		t.Fatal("Hit with empty User-Agent must be ignored")
	}
}

func TestIgnoreHitBotUserAgent(t *testing.T) {
	for _, botUserAgent := range userAgentBlacklist {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", botUserAgent)

		if !Ignore(req) {
			t.Fatalf("Hit with user agent '%v' must have been ignored", botUserAgent)
		}
	}
}

func TestIgnoreHitReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "ua")
	req.Header.Set("Referer", "2your.site")

	if !Ignore(req) {
		t.Fatal("Request must have been ignored")
	}

	req.Header.Set("Referer", "subdomain.2your.site")

	if !Ignore(req) {
		t.Fatal("Request for subdomain must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/?ref=2your.site", nil)

	if !Ignore(req) {
		t.Fatal("Request must have been ignored")
	}
}

func TestIgnoreHitBrowserVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.4147.135 Safari/537.36")

	if !Ignore(req) {
		t.Fatal("Request must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if Ignore(req) {
		t.Fatal("Request must not have been ignored")
	}
}

func TestIgnoreHitDoNotTrack(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if Ignore(req) {
		t.Fatal("Request must not have been ignored")
	}

	req.Header.Set("DNT", "1")

	if !Ignore(req) {
		t.Fatal("Request must have been ignored")
	}
}

func TestHitOptionsFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://test.com/my/path", nil)
	options := OptionsFromRequest(req)

	if options.TenantID.Int64 != 0 ||
		options.URL != "" ||
		options.Referrer != "" ||
		options.ScreenWidth != 0 ||
		options.ScreenHeight != 0 {
		t.Fatalf("Options not as expected: %v", options)
	}

	req = httptest.NewRequest(http.MethodGet, "http://test.com/my/path?tenantid=42&url=http://foo.bar/test&ref=http://ref/&w=640&h=1024", nil)
	options = OptionsFromRequest(req)

	if options.TenantID.Int64 != 42 ||
		options.URL != "http://foo.bar/test" ||
		options.Referrer != "http://ref/" ||
		options.ScreenWidth != 640 ||
		options.ScreenHeight != 1024 {
		t.Fatalf("Options not as expected: %v", options)
	}
}

func TestShortenString(t *testing.T) {
	out := shortenString("Hello World", 5)

	if out != "Hello" {
		t.Fatalf("String must have been shortened to 'Hello', but was: %v", out)
	}

	out = shortenString("Hello World", 50)

	if out != "Hello World" {
		t.Fatalf("String must not have been shortened, but was: %v", out)
	}
}

func TestGetNullInt64QueryParam(t *testing.T) {
	input := []string{
		"",
		"   ",
		"asdf",
		"32asdf",
		"42",
	}
	expected := []int64{
		0,
		0,
		0,
		0,
		42,
	}

	for i, in := range input {
		if out := getNullInt64QueryParam(in); out.Int64 != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], out)
		}
	}
}

func TestGetIntQueryParam(t *testing.T) {
	input := []string{
		"",
		"   ",
		"asdf",
		"32asdf",
		"42",
	}
	expected := []int{
		0,
		0,
		0,
		0,
		42,
	}

	for i, in := range input {
		if out := getIntQueryParam(in); out != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], out)
		}
	}
}
