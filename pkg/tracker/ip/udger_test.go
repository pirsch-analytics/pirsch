package ip

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestUdger(t *testing.T) {
	udger := NewUdger("", "", "")
	udger.Update([]string{
		"90.154.29.38",
	}, []string{
		"2003:e1:7f03:a7b7:6328:b96a:4061:9999",
	}, []Range{
		{"123.0.0.0", "123.10.0.5"},
	}, []Range{
		{"2001:1ab0:f001::", "2001:1ab0:f001:ffff:ffff:ffff:ffff:ffff"},
	})
	assert.False(t, udger.Ignore("91.154.29.38"))
	assert.False(t, udger.Ignore("123.10.0.6"))
	assert.True(t, udger.Ignore("123.10.0.4"))
	assert.True(t, udger.Ignore("123.5.123.69"))
	assert.True(t, udger.Ignore("123.0.0.0"))
	assert.True(t, udger.Ignore("123.10.0.5"))
	assert.True(t, udger.Ignore("90.154.29.38"))
	assert.False(t, udger.Ignore("2003:e1:7f03:a7b7:6328:b96a:4061:8581"))
	assert.True(t, udger.Ignore("2003:e1:7f03:a7b7:6328:b96a:4061:9999"))
	assert.True(t, udger.Ignore("2001:1ab0:f001::"))
	assert.True(t, udger.Ignore("2001:1ab0:f001:ffff:ffff:ffff:ffff:ffff"))
	assert.True(t, udger.Ignore("2001:1ab0:f001:1000:0000:0000:0000:00ff"))
}

func BenchmarkUdger(b *testing.B) {
	accessKey := os.Getenv("UDGER_ACCESS_KEY")

	if accessKey != "" {
		udger := NewUdger(accessKey, "tmp", "")
		assert.NoError(b, udger.DownloadAndUpdate())

		b.Run("IPv4", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				assert.False(b, udger.Ignore("91.36.189.125"))
			}
		})

		b.Run("IPv6", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				assert.False(b, udger.Ignore("2003:e1:7f03:a7b7:6328:b96a:4061:8581"))
			}
		})
	}
}

func TestUdger_Ignore(t *testing.T) {
	accessKey := os.Getenv("UDGER_ACCESS_KEY")
	ips := []string{}

	if accessKey != "" {
		udger := NewUdger(accessKey, "tmp", "")
		assert.NoError(t, udger.DownloadAndUpdate())
		ignored := make([]string, 0)

		for _, ip := range ips {
			if udger.Ignore(ip) {
				ignored = append(ignored, ip)
			}
		}

		assert.Empty(t, ignored)
	}
}
