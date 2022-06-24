package pirsch

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "123.456.789.012", getIP(r, DefaultHeaderParser, nil))

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	assert.Equal(t, "103.0.53.43", getIP(r, DefaultHeaderParser, nil))

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	assert.Equal(t, "192.0.2.60", getIP(r, DefaultHeaderParser, nil))

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser, nil))

	// True-Client-IP
	r.Header.Set("True-Client-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser, nil))

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser, nil))

	// no parser
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "123.456.789.012", getIP(r, nil, nil))
}

func TestGetIPWithProxy(t *testing.T) {
	allowedProxySubnetList := []string{"10.0.0.0/8"}
	allowedProxySubnets := make([]net.IPNet, 0)

	for _, v := range allowedProxySubnetList {
		_, cidr, err := net.ParseCIDR(v)

		if err != nil {
			continue
		}

		allowedProxySubnets = append(allowedProxySubnets, *cidr)
	}

	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.8:29302"

	// no header, default
	assert.Equal(t, "10.0.0.8", getIP(r, DefaultHeaderParser, allowedProxySubnets))

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	assert.Equal(t, "103.0.53.43", getIP(r, DefaultHeaderParser, allowedProxySubnets))

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	assert.Equal(t, "192.0.2.60", getIP(r, DefaultHeaderParser, allowedProxySubnets))

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser, allowedProxySubnets))

	// True-Client-IP
	r.Header.Set("True-Client-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser, allowedProxySubnets))

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "65.182.89.102", getIP(r, DefaultHeaderParser, allowedProxySubnets))

	// no parser
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "10.0.0.8", getIP(r, nil, allowedProxySubnets))

	// invalid remote IP
	r.RemoteAddr = "1.1.1.1"
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	assert.Equal(t, "1.1.1.1", getIP(r, DefaultHeaderParser, allowedProxySubnets))
}
