package tracker

import (
	"github.com/pirsch-analytics/pirsch/v6/internal/util"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
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
	}
	options.validate(req)
	assert.Equal(t, "https://example.com/new/path?query=parameter#anchor", options.URL)
	assert.Equal(t, "example.com", options.Hostname)
}

func TestOptionsFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://foo.bar?url=https://example.com/test&t=Title&ref=Referrer&w=1920&h=1080", nil)
	options := OptionsFromRequest(req)
	options.validate(req)
	assert.Equal(t, "https://example.com/test", options.URL)
	assert.Equal(t, "example.com", options.Hostname)
	assert.Equal(t, "/test", options.Path)
	assert.Equal(t, "Title", options.Title)
	assert.Equal(t, "Referrer", options.Referrer)
	assert.Equal(t, uint16(1920), options.ScreenWidth)
	assert.Equal(t, uint16(1080), options.ScreenHeight)
}
