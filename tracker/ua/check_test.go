package ua

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsNonASCIICharacters(t *testing.T) {
	input := []string{
		"��!�<~2��T��Ė�B;",
		"��!�Hh��L~v�;",
		"��C�j�P��E8��x�O|��",
	}

	for _, in := range input {
		assert.True(t, ContainsNonASCIICharacters(in))
	}

	for _, in := range userAgentsAll {
		assert.False(t, ContainsNonASCIICharacters(in.ua))
	}
}
