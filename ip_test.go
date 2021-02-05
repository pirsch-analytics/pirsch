package pirsch

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestParseForwardedHeader(t *testing.T) {
	header := []string{
		"for=12.34.56.78;host=example.com;proto=https, for=23.45.67.89",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=10.1.2.3",
		"for=192.0.2.60;proto=http;by=203.0.113.43",
		"proto=http;by=203.0.113.43;for=192.0.2.61",
		"   ",
		"",
	}
	expected := []string{
		"12.34.56.78",
		"12.34.56.78",
		"192.0.2.60",
		"192.0.2.61",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseForwardedHeader(head))
	}
}

func TestParseXForwardedForHeader(t *testing.T) {
	header := []string{
		"127.0.0.1",
		"127.0.0.1, 23.21.45.67",
		"127.0.0.1,23.21.45.67",
		"   ",
		"",
	}
	expected := []string{
		"127.0.0.1",
		"127.0.0.1",
		"127.0.0.1",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseXForwardedForHeader(head))
	}
}

func TestGetIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "123.456.789.012:29302"

	// no header, default
	assert.Equal(t, "123.456.789.012", getIP(r))

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	assert.Equal(t, "103.0.53.43", getIP(r))

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	assert.Equal(t, "192.0.2.60", getIP(r))

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67")
	assert.Equal(t, "127.0.0.1", getIP(r))

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67")
	assert.Equal(t, "127.0.0.1", getIP(r))
}
