package hit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetScreenClass(t *testing.T) {
	assert.Equal(t, "", GetScreenClass(0))
	assert.Equal(t, "XS", GetScreenClass(42))
	assert.Equal(t, "XL", GetScreenClass(1024))
	assert.Equal(t, "XL", GetScreenClass(1025))
	assert.Equal(t, "XXL", GetScreenClass(1919))
}
