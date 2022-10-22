package referrer

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	input := []struct {
		referrer  string
		blacklist []string
	}{
		{"http://boring.old/domain", nil},
		{"http://boring.old/domain/", nil},
		{"https://with.subdomain.com/", nil},
		{"https://with.multiple.subdomains.com/and/a/path?plus=query&params=42#anchor", nil},
		{"http://boring.old/domain", []string{"boring.old"}},
		{"https://with.subdomain.com/", []string{"boring.old"}},
		{"https://sub.boring.old/domain", []string{"boring.old"}},
		{"https://sub.boring.old/domain", []string{"boring.old"}},
		{"https://example.com/", nil},
		{"https://example.com", nil},
		{"ReferrerName", nil},
		{"  ", nil},
		{"https://www.pirsch.io/", []string{"pirsch.io"}},
		{"https://www.pirsch.io/", []string{"pirsch.io", "www.pirsch.io"}},
		{"https://www.pirsch.io/", []string{"pirsch.io"}},
		{"pirsch.io", []string{"pirsch.io"}},
		{"49.12.18.161", nil},
		{"49.12.18.161/", nil},
		{"http://49.12.18.161/", nil},
		{"168.119.249.160:8080", nil},
		{"168.119.249.160:8080/signup", nil},
		{"https://168.119.249.160:8080", nil},
		{"https://168.119.249.160:8080/signup", nil},
		{"https://example.com", nil},
		{"https://example.com/", nil},
		{"sub.example.com/with/path/", nil},
		{"http://sub.example.com/with/path/", nil},
		{"https://www.google.com", nil},
		{"https://www.google.bf", nil},
		{"https://google.com/", nil},
		{"https://google.bf", nil},
		{"https://www.google.pl/products", nil},
		{"https://t.co/asdf", nil},
		{"https://t.co", nil},
		{"HTTPS://T.CO", nil},
	}
	expected := []struct {
		referrer string
		name     string
	}{
		{"http://boring.old/domain", "boring.old"},
		{"http://boring.old/domain/", "boring.old"}, // trailing slashes only matter for non-root domain URLs
		{"https://with.subdomain.com", "subdomain.com"},
		{"https://with.multiple.subdomains.com/and/a/path", "subdomains.com"},
		{"", ""},
		{"https://with.subdomain.com", "subdomain.com"},
		{"", ""},
		{"", ""},
		{"https://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"", "ReferrerName"},
		{"", ""},
		{"", ""},
		{"", ""},
		{"", ""},
		{"", ""},
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
		{"http://sub.example.com/with/path/", "example.com"},
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
		r.Header.Add("Referer", in.referrer)
		referrer, referrerName, _ := Get(r, "", in.blacklist)
		assert.Equal(t, expected[i].referrer, referrer)
		assert.Equal(t, expected[i].name, referrerName)
	}
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
	_, name, icon := Get(r, "", nil)
	assert.Equal(t, "Slack", name)
	assert.NotEmpty(t, icon)
	r.Header.Set("Referer", androidAppPrefix+"does-not-exist")
	ref, name, icon := Get(r, "", nil)
	assert.Equal(t, androidAppPrefix+"does-not-exist", ref)
	assert.Empty(t, name)
	assert.Empty(t, icon)
}

func TestContainsString(t *testing.T) {
	list := []string{"a", "b", "c", "d"}
	assert.False(t, containsString(list, "e"))
	assert.True(t, containsString(list, "c"))
}
