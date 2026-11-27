package ip

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoLite2ASN(t *testing.T) {
	filter := NewGeoLite2ASN(GeoLite2ASNOptions{
		AllowASNs:            []uint{1221},
		DisableASNList:       true,
		OrganizationKeywords: []string{"google", "akamai", "telstra"},
	})
	assert.False(t, filter.Ignore("1.0.0.1"))
	assert.NoError(t, filter.UpdateFromFile("../../../test/GeoLite2-ASN-Test.mmdb"))
	assert.True(t, filter.Ignore("invalid"))
	assert.True(t, filter.Ignore("1.0.0.1"))
	assert.True(t, filter.Ignore("23.32.0.1"))
	assert.False(t, filter.Ignore("1.128.0.1"))
	assert.False(t, filter.Ignore("127.0.0.1"))
	assert.Error(t, filter.UpdateFromFile("../../../test/does-not-exist.mmdb"))
	assert.True(t, filter.Ignore("1.0.0.1"))
}

func TestGeoLite2ASNList(t *testing.T) {
	list := `AS15169 # Google
AS35994	# Akamai
-AS1221 # removed entry
# comment

invalid
7018,"AT&T Services"`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(list))
	}))
	defer server.Close()
	filter := NewGeoLite2ASN(GeoLite2ASNOptions{
		ASNListURL:                  server.URL,
		AllowASNs:                   []uint{7018},
		DisableOrganizationKeywords: true,
	})
	assert.NoError(t, filter.UpdateFromFile("../../../test/GeoLite2-ASN-Test.mmdb"))
	assert.NoError(t, filter.downloadASNList())
	assert.Len(t, filter.listASNs, 3)
	assert.True(t, filter.Ignore("1.0.0.1"))
	assert.True(t, filter.Ignore("23.32.0.1"))
	assert.False(t, filter.Ignore("1.128.0.1"))
	assert.False(t, filter.Ignore("12.81.92.1"))
	assert.Error(t, filter.updateASNList(strings.NewReader("# empty")))
	assert.Len(t, filter.listASNs, 3)
}

func TestGeoLite2ASNOrganizationKeywords(t *testing.T) {
	filter := NewGeoLite2ASN(GeoLite2ASNOptions{DisableASNList: true})
	assert.NoError(t, filter.UpdateFromFile("../../../test/GeoLite2-ASN-Test.mmdb"))

	// residential and mobile providers must not be filtered by the default keywords
	for _, ip := range []string{"1.128.0.1", "12.81.92.1", "31.224.0.1", "73.0.0.1", "79.192.0.1", "2003::1", "2a02:8100::1"} {
		assert.False(t, filter.Ignore(ip), ip)
	}

	filter = NewGeoLite2ASN(GeoLite2ASNOptions{
		DisableASNList:       true,
		OrganizationKeywords: []string{" AKAMAI "},
	})
	assert.NoError(t, filter.UpdateFromFile("../../../test/GeoLite2-ASN-Test.mmdb"))
	assert.True(t, filter.Ignore("23.32.0.1"))
	assert.False(t, filter.Ignore("1.0.0.1"))
}

func TestGeoLite2ASNDefaultAllowASNs(t *testing.T) {
	filter := NewGeoLite2ASN(GeoLite2ASNOptions{})
	assert.Len(t, filter.allowASNs, len(DefaultAllowASNs))
	assert.Contains(t, filter.allowASNs, uint(13335))
	filter = NewGeoLite2ASN(GeoLite2ASNOptions{AllowASNs: []uint{}})
	assert.Empty(t, filter.allowASNs)
}

func TestParseASNListLine(t *testing.T) {
	assert.Equal(t, uint(123), parseASNListLine("AS123 # comment"))
	assert.Equal(t, uint(123), parseASNListLine("  as123\t# comment"))
	assert.Equal(t, uint(123), parseASNListLine("123,\"Entity\""))
	assert.Equal(t, uint(0), parseASNListLine("-AS123"))
	assert.Equal(t, uint(0), parseASNListLine("ASN,Entity"))
	assert.Equal(t, uint(0), parseASNListLine("# AS123"))
	assert.Equal(t, uint(0), parseASNListLine("AS99999999999"))
}

func TestGeoLite2ASNDownloadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()
	filter := NewGeoLite2ASN(GeoLite2ASNOptions{
		LicenseKey:   "key",
		DownloadPath: t.TempDir(),
		DownloadURL:  server.URL,
		ASNListURL:   server.URL,
	})
	assert.Error(t, filter.DownloadAndUpdate())
	assert.Error(t, filter.downloadASNList())
	udger := NewUdger("key", t.TempDir(), server.URL+"/%s")
	assert.Error(t, udger.DownloadAndUpdate())
}
