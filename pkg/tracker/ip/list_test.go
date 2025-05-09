package ip

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
	}, nil, nil)
	assert.False(t, list.Ignore("91.154.29.38"))
	assert.False(t, list.Ignore("123.10.0.1"))
	assert.True(t, list.Ignore("90.154.29.38"))
	assert.False(t, list.Ignore("2001:1ab0:f001:a7b7:6328:b96a:4061:8888"))
	assert.False(t, list.Ignore("2011:1ab0:f001:a7b7:6328:b96a:4061:8888"))
	assert.True(t, list.Ignore("2003:e1:7f03:a7b7:6328:b96a:4061:9999"))
}
