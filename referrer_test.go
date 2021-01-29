package pirsch

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
		{"ReferrerName", nil, false},
		{"  ", nil, false},
		{"https://www.pirsch.io/", []string{"pirsch.io"}, false},
		{"https://www.pirsch.io/", []string{"pirsch.io", "www.pirsch.io"}, false},
		{"https://www.pirsch.io/", []string{"pirsch.io"}, true},
		{"pirsch.io", []string{"pirsch.io"}, false},
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
		"ReferrerName",
		"",
		"https://www.pirsch.io/",
		"",
		"",
		"",
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Add("Referer", in.referrer)

		if referrer, _, _ := getReferrer(r, "", in.blacklist, in.ignoreSubdomain); referrer != expected[i] {
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

func TestGetReferrerAndroidApp(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add("Referer", androidAppPrefix+"com.Slack")
	_, name, icon := getReferrer(r, "", nil, false)

	if name != "Slack" || icon == "" {
		t.Fatalf("Android app name and icon must have been returned, but was: %v %v", name, icon)
	}

	r.Header.Set("Referer", androidAppPrefix+"does-not-exist")
	ref, name, icon := getReferrer(r, "", nil, false)

	if ref != androidAppPrefix+"does-not-exist" || name != "" || icon != "" {
		t.Fatalf("Android app name and icon must not have been returned, but was: %v %v", name, icon)
	}
}
