package ip

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestIP(t *testing.T) {
	i := NewIP(DefaultHeaderParser, nil)
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "123.456.789.012:29302"
	req := &ingest.Request{
		Request: r,
	}

	// no header, default
	cancel, err := i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "123.456.789.012", req.IP)

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "103.0.53.43", req.IP)

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "192.0.2.60", req.IP)

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "65.182.89.102", req.IP)

	// True-Client-IP
	r.Header.Set("True-Client-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "65.182.89.102", req.IP)

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "65.182.89.102", req.IP)

	// no parser
	i.parser = nil
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "123.456.789.012", req.IP)
}

func TestIPWithProxy(t *testing.T) {
	allowedProxySubnetList := []string{"10.0.0.0/8"}
	allowedProxySubnets := make([]net.IPNet, 0)

	for _, v := range allowedProxySubnetList {
		_, cidr, err := net.ParseCIDR(v)

		if err != nil {
			continue
		}

		allowedProxySubnets = append(allowedProxySubnets, *cidr)
	}

	i := NewIP(DefaultHeaderParser, allowedProxySubnets)
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.8:29302"
	req := &ingest.Request{
		Request: r,
	}

	// no header, default
	cancel, err := i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.8", req.IP)

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "103.0.53.43", req.IP)

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "192.0.2.60", req.IP)

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "65.182.89.102", req.IP)

	// True-Client-IP
	r.Header.Set("True-Client-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "65.182.89.102", req.IP)

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "65.182.89.102", req.IP)

	// no parser
	i.parser = nil
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.8", req.IP)

	// invalid remote IP
	r.RemoteAddr = "1.1.1.1"
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67, 65.182.89.102")
	cancel, err = i.Step(req)
	assert.True(t, cancel)
	assert.NoError(t, err)
}
