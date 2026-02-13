package tracker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestOptions_validate(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	options := Options{Title: util.RandString(600)}
	options.validate(req)
	assert.Equal(t, "https://example.com", options.URL)
	assert.Equal(t, "example.com", options.Hostname)
	assert.Len(t, options.Title, 512)

	options = Options{URL: "https://example.com/foo/bar?query=parameter#anchor"}
	options.validate(req)
	assert.Equal(t, "https://example.com/foo/bar?query=parameter#anchor", options.URL)
	assert.Equal(t, "example.com", options.Hostname)

	options = Options{
		URL:  "https://example.com/foo/bar?query=parameter#anchor",
		Path: "/new/path",
		Tags: map[string]string{
			"key0":   "value0",
			" key1 ": " value1 ",
			"ignore": "",
			"":       "ignore",
		},
	}
	options.validate(req)
	assert.Equal(t, "https://example.com/new/path?query=parameter#anchor", options.URL)
	assert.Equal(t, "example.com", options.Hostname)
	k, v := options.getTags()
	assert.Len(t, k, 2)
	assert.Len(t, v, 2)
	assert.Contains(t, k, "key0")
	assert.Contains(t, k, "key1")
	assert.Contains(t, v, "value0")
	assert.Contains(t, v, "value1")
}
