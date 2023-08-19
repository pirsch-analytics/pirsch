package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsNonASCIICharacters(t *testing.T) {
	nonASCII := []string{
		"��!�<~2��T��Ė�B;",
		"��!�Hh��L~v�;",
		"��C�j�P��E8��x�O|��",
	}
	onlyASCII := []string{
		"ascii",
	}

	for _, in := range nonASCII {
		assert.True(t, ContainsNonASCIICharacters(in))
	}

	for _, in := range onlyASCII {
		assert.False(t, ContainsNonASCIICharacters(in))
	}
}

func TestShortenString(t *testing.T) {
	assert.Equal(t, "abcd", ShortenString("abcdefghi", 4))
	assert.Equal(t, "abcdefghi", ShortenString("abcdefghi", 100))
}
