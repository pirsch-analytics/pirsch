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
		referrer, _, _ := getReferrer(r, "", in.blacklist, in.ignoreSubdomain)
		assert.Equal(t, expected[i], referrer)
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
