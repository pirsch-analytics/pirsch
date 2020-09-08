package pirsch

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHitFromRequest(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	req.Header.Set("Referer", "ref")
	hit := HitFromRequest(req, "salt", &HitOptions{
		TenantID:     NewTenantID(42),
		sessionCache: newSessionCache(store, 0, 0),
	})

	if hit.TenantID.Int64 != 42 ||
		!hit.TenantID.Valid ||
		hit.Fingerprint == "" ||
		!hit.Session.Valid || hit.Session.Time.IsZero() ||
		hit.Path.String != "/test/path" ||
		hit.URL.String != "/test/path?query=param&foo=bar#anchor" ||
		hit.Language.String != "de-de" ||
		hit.UserAgent.String != "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36" ||
		hit.Referrer.String != "ref" ||
		hit.OS.String != OSWindows ||
		hit.OSVersion.String != "10" ||
		hit.Browser.String != BrowserChrome ||
		hit.BrowserVersion.String != "84.0" ||
		!hit.Desktop ||
		hit.Mobile ||
		hit.Time.IsZero() {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	hit := HitFromRequest(req, "salt", &HitOptions{
		Path: "/new/custom/path",
	})

	if hit.Path.String != "/new/custom/path" ||
		hit.URL.String != "/new/custom/path?query=param&foo=bar#anchor" {
		t.Fatalf("Hit not as expected: %v", hit)
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
	req.Header.Add("Referer", "2your.site")

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/?ref=2your.site", nil)

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
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
	}
	expected := []string{
		"http://boring.old/domain",
		"https://with.subdomain.com/",
		"https://with.multiple.subdomains.com/and/a/path",
		"",
		"https://with.subdomain.com/",
		"https://sub.boring.old/domain",
		"",
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Add("Referer", in.referrer)

		if referrer := getReferrer(r, in.blacklist, in.ignoreSubdomain); referrer != expected[i] {
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
	}
	expected := []string{
		"",
		"",
		"domain",
		"",
		"domain",
		"",
		"domain",
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
