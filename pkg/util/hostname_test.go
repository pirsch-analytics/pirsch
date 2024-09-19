package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
