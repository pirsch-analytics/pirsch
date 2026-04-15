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

func TestContainsNonASCIICharacters(t *testing.T) {
	input := []string{
		"��!�<~2��T��Ė�B;",
		"��!�Hh��L~v�;",
		"��C�j�P��E8��x�O|��",
	}

	for _, in := range input {
		assert.True(t, ContainsNonASCIICharacters(in))
	}

	input = []string{
		"this is fine",
	}

	for _, in := range input {
		assert.False(t, ContainsNonASCIICharacters(in))
	}
}
