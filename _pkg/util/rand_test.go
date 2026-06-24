package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	assert.Len(t, RandString(20), 20)
	assert.NotEqual(t, RandString(10), RandString(10))
}
