package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	list := NewList()
	list.Update([]string{
		"90.154.29.38",
		"123.10.0.1",
	}, []string{
		"2003:e1:7f03:a7b7:6328:b96a:4061:9999",
		"2001:1ab0:f001:a7b7:6328:b96a:4061:8888",
	}, []string{
		"123.10.0.1",
	}, []string{
		"2001:1ab0:f001:a7b7:6328:b96a:4061:8888",
	}, []Range{
		{From: "199.16.0.0", To: "199.16.0.199"},
	}, []Range{
		{From: "2001:0db8:0000:0000:0000:0000:0000:0000", To: "2001:0db8:ffff:ffff:ffff:ffff:ffff:ffff"},
	})
	assert.False(t, list.Ignore("91.154.29.38"))
	assert.False(t, list.Ignore("123.10.0.1"))
	assert.True(t, list.Ignore("90.154.29.38"))
	assert.True(t, list.Ignore("199.16.0.0"))
	assert.True(t, list.Ignore("199.16.0.1"))
	assert.True(t, list.Ignore("199.16.0.199"))
	assert.False(t, list.Ignore("199.16.0.200"))
	assert.False(t, list.Ignore("2001:1ab0:f001:a7b7:6328:b96a:4061:8888"))
	assert.False(t, list.Ignore("2011:1ab0:f001:a7b7:6328:b96a:4061:8888"))
	assert.True(t, list.Ignore("2003:e1:7f03:a7b7:6328:b96a:4061:9999"))
	assert.True(t, list.Ignore("2001:db8:85a3::8a2e:370:7334"))
	assert.False(t, list.Ignore("2001:db9:85a3::8a2e:370:7334"))
}
