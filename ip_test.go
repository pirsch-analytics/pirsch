package pirsch

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestParseForwardedHeader(t *testing.T) {
	header := []string{
		"for=12.34.56.78;host=example.com;proto=https, for=23.45.67.89",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=65.182.89.102",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=10.1.2.3",
		"for=192.0.2.60;proto=http;by=203.0.113.43",
		"for=10.1.2.3;proto=http;by=203.0.113.43",
		"proto=http;by=203.0.113.43;for=192.0.2.61",
		"proto=http;by=203.0.113.43;for=10.1.2.3",
		"   ",
		"",
	}
	expected := []string{
		"23.45.67.89",
		"65.182.89.102",
		"",
		"192.0.2.60",
		"",
		"192.0.2.61",
		"",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseForwardedHeader(head))
	}
}

func TestParseXForwardedForHeader(t *testing.T) {
	header := []string{
		"65.182.89.102",
		"127.0.0.1, 23.21.45.67, 65.182.89.102",
		"127.0.0.1,23.21.45.67,65.182.89.102",
		"65.182.89.102,23.21.45.67,127.0.0.1",
		"   ",
		"",
	}
	expected := []string{
		"65.182.89.102",
		"65.182.89.102",
		"65.182.89.102",
		"",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseXForwardedForHeader(head))
	}
}

func TestParseXRealIPHeader(t *testing.T) {
	header := []string{
		"",
		"  ",
		"invalid",
		"127.0.0.1",
		"65.182.89.102",
	}
	expected := []string{
		"",
		"",
		"",
		"",
		"65.182.89.102",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], parseXRealIPHeader(head))
	}
}

func TestGetIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "123.456.789.012:29302"

	// no header, default
	assert.Equal(t, "123.456.789.012", getIP(r, DefaultHeaderParser))

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	assert.Equal(t, "103.0.53.43", getIP(r, DefaultHeaderParser))

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	assert.Equal(t, "192.0.2.60", getIP(r, DefaultHeaderParser))

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser))

	// True-Client-IP
	r.Header.Set("True-Client-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser))

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser))

	// no parser
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "123.456.789.012", getIP(r, nil))
}
