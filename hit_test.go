package pirsch

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHitFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "user-agent")
	req.Header.Set("Referer", "ref")
	hit := HitFromRequest(req, "salt", &HitOptions{
		TenantID: NewTenantID(42),
	})

	if hit.TenantID.Int64 != 42 ||
		!hit.TenantID.Valid ||
		hit.Fingerprint == "" ||
		hit.Path != "/test/path" ||
		hit.URL != "/test/path?query=param&foo=bar#anchor" ||
		hit.Language != "de-de" ||
		hit.UserAgent != "user-agent" ||
		hit.Ref != "ref" ||
		hit.Time.IsZero() {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	hit := HitFromRequest(req, "salt", &HitOptions{
		Path: "/new/custom/path",
	})

	if hit.Path != "/new/custom/path" ||
		hit.URL != "/new/custom/path?query=param&foo=bar#anchor" {
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

func TestGetReferer(t *testing.T) {
	input := []struct {
		referer         string
		blacklist       []string
		ignoreSubdomain bool
	}{
		{"http://boring.old/domain", nil, false},
		{"https://with.subdomain.com/", nil, false},
		{"https://with.multiple.subdomains.com/and/a/path?plus=query&params=42", nil, false},
		{"http://boring.old/domain", []string{"boring.old"}, false},
		{"https://with.subdomain.com/", []string{"boring.old"}, false},
		{"https://sub.boring.old/domain", []string{"boring.old"}, false},
		{"https://sub.boring.old/domain", []string{"boring.old"}, true},
	}
	expected := []string{
		"http://boring.old/domain",
		"https://with.subdomain.com/",
		"https://with.multiple.subdomains.com/and/a/path?plus=query&params=42",
		"",
		"https://with.subdomain.com/",
		"https://sub.boring.old/domain",
		"",
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Add("Referer", in.referer)

		if referer := getReferer(r, in.blacklist, in.ignoreSubdomain); referer != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], referer)
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
