package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripWWW(t *testing.T) {
	list := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"example.com", "example.com"},
		{"sub.example.com", "sub.example.com"},
		{"www.sub.example.com", "www.sub.example.com"},
		{"www.example.com", "example.com"},
	}

	for _, item := range list {
		assert.Equal(t, item.out, StripWWW(item.in))
	}
}

func TestShorten(t *testing.T) {
	assert.Equal(t, "abcd", Shorten("abcdefghi", 4))
	assert.Equal(t, "abcdefghi", Shorten("abcdefghi", 100))
}
