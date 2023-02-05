package referrer

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	input := []string{
		"http://boring.old/domain",
		"http://boring.old/domain/",
		"https://with.subdomain.com/",
		"https://with.multiple.subdomains.com/and/a/path?plus=query&params=42#anchor",
		"https://example.com/",
		"https://example.com",
		"ReferrerName",
		"  ",
		"pirsch.io",
		"49.12.18.161",
		"49.12.18.161/",
		"http://49.12.18.161/",
		"168.119.249.160:8080",
		"168.119.249.160:8080/signup",
		"https://168.119.249.160:8080",
		"https://168.119.249.160:8080/signup",
		"https://example.com",
		"https://example.com/",
		"sub.example.com/with/path/",
		"http://sub.example.com/with/path/",
		"https://www.google.com",
		"https://www.google.bf",
		"https://google.com/",
		"https://google.bf",
		"https://www.google.pl/products",
		"https://t.co/asdf",
		"https://t.co",
		"HTTPS://T.CO",
	}
	expected := []struct {
		referrer string
		name     string
	}{
		{"http://boring.old/domain", "boring.old"},
		{"http://boring.old/domain/", "boring.old"}, // trailing slashes only matter for non-root domain URLs
		{"https://with.subdomain.com", "with.subdomain.com"},
		{"https://with.multiple.subdomains.com/and/a/path", "with.multiple.subdomains.com"},
		{"https://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"", "ReferrerName"},
		{"", ""},
		{"", "pirsch.io"},
		{"", ""},
		{"", ""},
		{"", ""},
		{"", ""},
		{"", ""},
		{"", ""},
		{"", ""},
		{"https://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"", "sub.example.com/with/path/"},
		{"http://sub.example.com/with/path/", "sub.example.com"},
		{"https://www.google.com", "Google"},
		{"https://www.google.bf", "Google"},
		{"https://google.com", "Google"},
		{"https://google.bf", "Google"},
		{"https://www.google.pl/products", "Google Product Search"},
		{"https://t.co/asdf", "Twitter"},
		{"https://t.co", "Twitter"},
		{"https://t.co", "Twitter"},
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Add("Referer", in)
		referrer, referrerName, _ := Get(r, "", "")
		assert.Equal(t, expected[i].referrer, referrer)
		assert.Equal(t, expected[i].name, referrerName)
	}
}

func TestGetSameDomain(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	r.Header.Add("Referer", "https://example.com/foo/bar")
	referrer, referrerName, referrerIcon := Get(r, "", "example.com")
	assert.Empty(t, referrer)
	assert.Empty(t, referrerName)
	assert.Empty(t, referrerIcon)
	r = httptest.NewRequest(http.MethodGet, "https://example.com:8080/bar/foo", nil)
	referrer, referrerName, referrerIcon = Get(r, "https://example.com:8080/foo/bar", "example.com")
	assert.Empty(t, referrer)
	assert.Empty(t, referrerName)
	assert.Empty(t, referrerIcon)
	r = httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	r.Header.Add("Referer", "https://sub.example.com/foo/bar")
	referrer, referrerName, referrerIcon = Get(r, "", "example.com")
	assert.Equal(t, "https://sub.example.com/foo/bar", referrer)
	assert.Equal(t, "sub.example.com", referrerName)
	assert.Empty(t, referrerIcon)
}

func TestGetFromHeaderOrQuery(t *testing.T) {
	input := [][]string{
		{"", ""},
		{"ref", ""},
		{"ref", "domain"},
		{"ref", "domain+space"},
		{"referer", ""},
		{"referer", "domain"},
		{"referer", "domain+space"},
		{"referrer", ""},
		{"referrer", "domain"},
		{"referrer", "domain+space"},
		{"source", ""},
		{"source", "domain"},
		{"source", "domain+space"},
		{"utm_source", ""},
		{"utm_source", "domain"},
		{"utm_source", "domain+space"},
	}
	expected := []string{
		"",
		"",
		"domain",
		"domain space",
		"",
		"domain",
		"domain space",
		"",
		"domain",
		"domain space",
		"",
		"domain",
		"domain space",
		"",
		"domain",
		"domain space",
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/?"+in[0]+"="+in[1], nil)
		assert.Equal(t, expected[i], getFromHeaderOrQuery(r))
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
		assert.Equal(t, expected[i], stripSubdomain(in))
	}
}

func TestGetAndroidApp(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add("Referer", androidAppPrefix+"com.Slack")
	_, name, icon := Get(r, "", "")
	assert.Equal(t, "Slack", name)
	assert.NotEmpty(t, icon)
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add("Referer", androidAppPrefix+"com.pinterest/")
	_, name, icon = Get(r, "", "")
	assert.Equal(t, "Pinterest", name)
	assert.NotEmpty(t, icon)
	r.Header.Set("Referer", androidAppPrefix+"does-not-exist")
	ref, name, icon := Get(r, "", "")
	assert.Equal(t, androidAppPrefix+"does-not-exist", ref)
	assert.Empty(t, name)
	assert.Empty(t, icon)
}
