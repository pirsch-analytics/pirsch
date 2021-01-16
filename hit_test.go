package pirsch

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHitFromRequest(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	req.Header.Set("Referer", "http://ref/")
	hit := HitFromRequest(req, "salt", &HitOptions{
		TenantID:     NewTenantID(42),
		sessionCache: newSessionCache(store, nil),
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})

	if hit.TenantID.Int64 != 42 ||
		!hit.TenantID.Valid ||
		hit.Fingerprint == "" ||
		!hit.Session.Valid || hit.Session.Time.IsZero() ||
		hit.Path != "/test/path" ||
		hit.URL.String != "/test/path?query=param&foo=bar#anchor" ||
		hit.Language.String != "de" ||
		hit.UserAgent.String != "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36" ||
		hit.Referrer.String != "http://ref/" ||
		hit.OS.String != OSWindows ||
		hit.OSVersion.String != "10" ||
		hit.Browser.String != BrowserChrome ||
		hit.BrowserVersion.String != "84.0" ||
		!hit.Desktop ||
		hit.Mobile ||
		hit.ScreenWidth != 640 ||
		hit.ScreenHeight != 1024 ||
		hit.Time.IsZero() {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestOverwrite(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	hit := HitFromRequest(req, "salt", &HitOptions{
		URL: "http://bar.foo/new/custom/path?query=param&foo=bar#anchor",
	})

	if hit.Path != "/new/custom/path" ||
		hit.URL.String != "http://bar.foo/new/custom/path?query=param&foo=bar#anchor" {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestOverwritePathAndReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	hit := HitFromRequest(req, "salt", &HitOptions{
		URL:      "http://bar.foo/overwrite/this?query=param&foo=bar#anchor",
		Path:     "/new/custom/path",
		Referrer: "http://custom.ref/",
	})

	if hit.Path != "/new/custom/path" ||
		hit.URL.String != "http://bar.foo/new/custom/path?query=param&foo=bar#anchor" ||
		hit.Referrer.String != "http://custom.ref/" {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestScreenSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	hit := HitFromRequest(req, "salt", &HitOptions{
		ScreenWidth:  -5,
		ScreenHeight: 400,
	})

	if hit.ScreenWidth != 0 || hit.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", hit.ScreenWidth, hit.ScreenHeight)
	}

	hit = HitFromRequest(req, "salt", &HitOptions{
		ScreenWidth:  400,
		ScreenHeight: 0,
	})

	if hit.ScreenWidth != 0 || hit.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", hit.ScreenWidth, hit.ScreenHeight)
	}

	hit = HitFromRequest(req, "salt", &HitOptions{
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

	defer geoDB.Close()
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	req.RemoteAddr = "81.2.69.142"
	hit := HitFromRequest(req, "salt", &HitOptions{
		geoDB: geoDB,
	})

	if hit.CountryCode.String != "gb" {
		t.Fatalf("Country code for hit must have been returned, but was: %v", hit.CountryCode.String)
	}

	req = httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	req.RemoteAddr = "127.0.0.1"
	hit = HitFromRequest(req, "salt", &HitOptions{
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

	if !IgnoreHit(req) {
		t.Fatal("Hit with X-Moz header must be ignored")
	}

	req.Header.Del("X-Moz")
	req.Header.Set("X-Purpose", "prefetch")

	if !IgnoreHit(req) {
		t.Fatal("Hit with X-Purpose header must be ignored")
	}

	req.Header.Set("X-Purpose", "preview")

	if !IgnoreHit(req) {
		t.Fatal("Hit with X-Purpose header must be ignored")
	}

	req.Header.Del("X-Purpose")
	req.Header.Set("Purpose", "prefetch")

	if !IgnoreHit(req) {
		t.Fatal("Hit with Purpose header must be ignored")
	}

	req.Header.Set("Purpose", "preview")

	if !IgnoreHit(req) {
		t.Fatal("Hit with Purpose header must be ignored")
	}

	req.Header.Del("Purpose")

	if IgnoreHit(req) {
		t.Fatal("Hit must not be ignored")
	}
}

func TestIgnoreHitUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "This is a bot request")

	if !IgnoreHit(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a crawler request")

	if !IgnoreHit(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a spider request")

	if !IgnoreHit(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Visit http://spam.com!")

	if !IgnoreHit(req) {
		t.Fatal("Hit with URL in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Mozilla/123.0")

	if IgnoreHit(req) {
		t.Fatal("Hit with regular User-Agent must not be ignored")
	}

	req.Header.Set("User-Agent", "")

	if !IgnoreHit(req) {
		t.Fatal("Hit with empty User-Agent must be ignored")
	}
}

func TestIgnoreHitBotUserAgent(t *testing.T) {
	for _, botUserAgent := range userAgentBlacklist {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", botUserAgent)

		if !IgnoreHit(req) {
			t.Fatalf("Hit with user agent '%v' must have been ignored", botUserAgent)
		}
	}
}

func TestIgnoreHitReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "ua")
	req.Header.Set("Referer", "2your.site")

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}

	req.Header.Set("Referer", "subdomain.2your.site")

	if !IgnoreHit(req) {
		t.Fatal("Request for subdomain must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/?ref=2your.site", nil)

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}
}

func TestIgnoreHitBrowserVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.4147.135 Safari/537.36")

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if IgnoreHit(req) {
		t.Fatal("Request must not have been ignored")
	}
}

func TestIgnoreHitDoNotTrack(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if IgnoreHit(req) {
		t.Fatal("Request must not have been ignored")
	}

	req.Header.Set("DNT", "1")

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}
}

func TestHitOptionsFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://test.com/my/path", nil)
	options := HitOptionsFromRequest(req)

	if options.TenantID.Int64 != 0 ||
		options.URL != "" ||
		options.Referrer != "" ||
		options.ScreenWidth != 0 ||
		options.ScreenHeight != 0 {
		t.Fatalf("Options not as expected: %v", options)
	}

	req = httptest.NewRequest(http.MethodGet, "http://test.com/my/path?tenantid=42&url=http://foo.bar/test&ref=http://ref/&w=640&h=1024", nil)
	options = HitOptionsFromRequest(req)

	if options.TenantID.Int64 != 42 ||
		options.URL != "http://foo.bar/test" ||
		options.Referrer != "http://ref/" ||
		options.ScreenWidth != 640 ||
		options.ScreenHeight != 1024 {
		t.Fatalf("Options not as expected: %v", options)
	}
}

func TestGetReferrer(t *testing.T) {
	input := []struct {
		referrer        string
		blacklist       []string
		ignoreSubdomain bool
	}{
		{"http://boring.old/domain", nil, false},
		{"https://with.subdomain.com/", nil, false},
		{"https://with.multiple.subdomains.com/and/a/path?plus=query&params=42#anchor", nil, false},
		{"http://boring.old/domain", []string{"boring.old"}, false},
		{"https://with.subdomain.com/", []string{"boring.old"}, false},
		{"https://sub.boring.old/domain", []string{"boring.old"}, false},
		{"https://sub.boring.old/domain", []string{"boring.old"}, true},
		{"https://example.com/", nil, false},
		{"https://example.com", nil, false},
	}
	expected := []string{
		"http://boring.old/domain",
		"https://with.subdomain.com/",
		"https://with.multiple.subdomains.com/and/a/path",
		"",
		"https://with.subdomain.com/",
		"https://sub.boring.old/domain",
		"",
		"https://example.com/",
		"https://example.com/",
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Add("Referer", in.referrer)

		if referrer := getReferrer(r, "", in.blacklist, in.ignoreSubdomain); referrer != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], referrer)
		}
	}
}

func TestGetReferrerFromHeaderOrQuery(t *testing.T) {
	input := [][]string{
		{"", ""},
		{"ref", ""},
		{"ref", "domain"},
		{"referer", ""},
		{"referer", "domain"},
		{"referrer", ""},
		{"referrer", "domain"},
		{"source", "domain"},
		{"source", ""},
		{"utm_source", "domain"},
		{"utm_source", ""},
	}
	expected := []string{
		"",
		"",
		"domain",
		"",
		"domain",
		"",
		"domain",
		"domain",
		"",
		"domain",
		"",
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/?"+in[0]+"="+in[1], nil)

		if out := getReferrerFromHeaderOrQuery(r); out != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], out)
		}
	}
}

func TestStripSubdomain(t *testing.T) {
	input := []string{
		"",
		".",
		"..",
		"...",
		" ",
		"\t",
		"boring.old",
		"with.subdomain.com",
		"with.multiple.subdomains.com",
	}
	expected := []string{
		"",
		".",
		"..",
		".",
		" ",
		"\t",
		"boring.old",
		"subdomain.com",
		"subdomains.com",
	}

	for i, in := range input {
		if hostname := stripSubdomain(in); hostname != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], hostname)
		}
	}
}

func TestGetLanguage(t *testing.T) {
	input := []string{
		"",
		"  \t ",
		"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5",
		"en-us, en",
		"en-gb, en",
		"invalid",
	}
	expected := []string{
		"",
		"",
		"fr",
		"en",
		"en",
		"",
	}

	for i, in := range input {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", in)

		if lang := getLanguage(req); lang != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], lang)
		}
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
