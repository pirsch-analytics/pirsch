package ip

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUdger(t *testing.T) {
	udger := NewUdger()
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
