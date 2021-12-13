package pirsch

import (
	"github.com/stretchr/testify/assert"
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
		{"http://boring.old/domain/", nil, false},
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
		{"49.12.18.161", nil, false},
		{"49.12.18.161/", nil, false},
		{"http://49.12.18.161/", nil, false},
		{"168.119.249.160:8080", nil, false},
		{"168.119.249.160:8080/signup", nil, false},
		{"https://168.119.249.160:8080", nil, false},
		{"https://168.119.249.160:8080/signup", nil, false},
		{"https://example.com", nil, false},
		{"https://example.com/", nil, false},
		{"sub.example.com/with/path/", nil, false},
		{"http://sub.example.com/with/path/", nil, false},
		{"https://www.google.com", nil, false},
		{"https://www.google.bf", nil, false},
		{"https://google.com/", nil, false},
		{"https://google.bf", nil, false},
		{"https://www.google.pl/products", nil, false},
	}
	expected := []struct {
		referrer string
		name     string
	}{
		{"http://boring.old/domain", "boring.old"},
		{"http://boring.old/domain/", "boring.old"}, // trailing slashes only matter for non-root domain URLs
		{"https://with.subdomain.com", "with.subdomain.com"},
		{"https://with.multiple.subdomains.com/and/a/path", "with.multiple.subdomains.com"},
		{"", ""},
		{"https://with.subdomain.com", "with.subdomain.com"},
		{"https://sub.boring.old/domain", "sub.boring.old"},
		{"", ""},
		{"https://example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"", "ReferrerName"},
		{"", ""},
		{"https://www.pirsch.io", "www.pirsch.io"},
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
		{"http://sub.example.com/with/path/", "sub.example.com"},
		{"https://www.google.com", "Google"},
		{"https://www.google.bf", "Google"},
		{"https://google.com", "Google"},
		{"https://google.bf", "Google"},
		{"https://www.google.pl/products", "Google Product Search"},
	}

	for i, in := range input {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Add("Referer", in.referrer)
		referrer, referrerName, _ := getReferrer(r, "", in.blacklist, in.ignoreSubdomain)
		assert.Equal(t, expected[i].referrer, referrer)
		assert.Equal(t, expected[i].name, referrerName)
	}
}

func TestGetReferrerFromHeaderOrQuery(t *testing.T) {
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
		assert.Equal(t, expected[i], getReferrerFromHeaderOrQuery(r))
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

func TestGetReferrerAndroidApp(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add("Referer", androidAppPrefix+"com.Slack")
	_, name, icon := getReferrer(r, "", nil, false)
	assert.Equal(t, "Slack", name)
	assert.NotEmpty(t, icon)
	r.Header.Set("Referer", androidAppPrefix+"does-not-exist")
	ref, name, icon := getReferrer(r, "", nil, false)
	assert.Equal(t, androidAppPrefix+"does-not-exist", ref)
	assert.Empty(t, name)
	assert.Empty(t, icon)
}

func TestContainsString(t *testing.T) {
	list := []string{"a", "b", "c", "d"}
	assert.False(t, containsString(list, "e"))
	assert.True(t, containsString(list, "c"))
}
