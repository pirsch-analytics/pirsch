package referrer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferrerFromHeaderOrQuery(t *testing.T) {
	input := [][]string{
		{"", "", ""},
		{"ref", "", ""},
		{"ref", "domain", ""},
		{"ref", "domain+space", ""},
		{"ref", "domain+space", "https://overwrite-this.com"},
		{"source", "domain+space", "https://overwrite-this.com"},
		{"utm_source", "domain+space", "https://overwrite-this.com"},
		{"referer", "", ""},
		{"referer", "domain", ""},
		{"referer", "domain+space", ""},
		{"referrer", "", ""},
		{"referrer", "domain", ""},
		{"referrer", "domain+space", ""},
		{"source", "", ""},
		{"source", "domain", ""},
		{"source", "domain+space", ""},
		{"utm_source", "", ""},
		{"utm_source", "domain", ""},
		{"utm_source", "domain+space", ""},
	}
	expected := []string{
		"",
		"",
		"domain",
		"domain space",
		"domain space",
		"https://overwrite-this.com",
		"https://overwrite-this.com",
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

		if in[2] != "" {
			r.Header.Set("Referer", in[2])
		}

		assert.Equal(t, expected[i], fromHeaderOrQuery(r))
	}
}
