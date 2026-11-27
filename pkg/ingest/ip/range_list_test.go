package ip

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeList(t *testing.T) {
	var fail atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() && r.URL.Path == "/a" {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		switch r.URL.Path {
		case "/a":
			_, _ = w.Write([]byte("10.0.0.0/24\n10.0.1.0/24\n# comment\ninvalid\n192.168.1.1\n2001:db8::/32\n"))
		case "/b":
			_, _ = w.Write([]byte("\"cidr\",\"hostmin\",\"hostmax\",\"vendor\"\n" +
				"\"172.16.0.0/16\",\"172.16.0.0\",\"172.16.255.255\",\"AWS\"\n" +
				"\"104.16.0.0/13\",\"104.16.0.0\",\"104.23.255.255\",\"CloudFlare\"\n"))
		}
	}))
	defer server.Close()
	list := NewRangeList([]RangeSource{
		{URL: server.URL + "/a"},
		{URL: server.URL + "/b", Exclude: []string{"cloudflare"}},
	})
	assert.False(t, list.Ignore("10.0.0.1"))
	assert.NoError(t, list.DownloadAndUpdate())
	assert.Len(t, list.rangesV4, 3)
	assert.Len(t, list.rangesV6, 1)
	assert.True(t, list.Ignore("invalid"))
	assert.True(t, list.Ignore("10.0.0.0"))
	assert.True(t, list.Ignore("10.0.1.255"))
	assert.True(t, list.Ignore("::ffff:10.0.1.1"))
	assert.False(t, list.Ignore("10.0.2.0"))
	assert.False(t, list.Ignore("9.255.255.255"))
	assert.True(t, list.Ignore("192.168.1.1"))
	assert.False(t, list.Ignore("192.168.1.2"))
	assert.True(t, list.Ignore("172.16.200.3"))
	assert.False(t, list.Ignore("104.16.0.1"))
	assert.True(t, list.Ignore("2001:db8:ffff::1"))
	assert.False(t, list.Ignore("2001:db9::1"))

	// keep previous ranges if a list fails
	fail.Store(true)
	assert.Error(t, list.DownloadAndUpdate())
	assert.True(t, list.Ignore("10.0.0.1"))
	assert.True(t, list.Ignore("172.16.0.1"))
}

func TestParseRangeLine(t *testing.T) {
	r, ok := parseRangeLine("10.1.2.3/8")
	assert.True(t, ok)
	assert.Equal(t, "10.0.0.0", r.from.String())
	assert.Equal(t, "10.255.255.255", r.to.String())
	r, ok = parseRangeLine("  1.2.3.0/22 # comment")
	assert.True(t, ok)
	assert.Equal(t, "1.2.0.0", r.from.String())
	assert.Equal(t, "1.2.3.255", r.to.String())
	r, ok = parseRangeLine("2001:db8::/33")
	assert.True(t, ok)
	assert.Equal(t, "2001:db8:7fff:ffff:ffff:ffff:ffff:ffff", r.to.String())
	r, ok = parseRangeLine("0.0.0.0/0")
	assert.True(t, ok)
	assert.Equal(t, "255.255.255.255", r.to.String())
	_, ok = parseRangeLine("\"cidr\",\"hostmin\"")
	assert.False(t, ok)
	_, ok = parseRangeLine("1.2.3.0/33")
	assert.False(t, ok)
}

func TestDefaultRangeSources(t *testing.T) {
	assert.Len(t, DefaultRangeSources, 1+9*2+7+23)
}
