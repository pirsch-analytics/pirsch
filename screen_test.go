package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetScreenClass(t *testing.T) {
	assert.Equal(t, "", GetScreenClass(0))
	assert.Equal(t, "XS", GetScreenClass(42))
	assert.Equal(t, "XL", GetScreenClass(1024))
	assert.Equal(t, "XL", GetScreenClass(1025))
	assert.Equal(t, "HD", GetScreenClass(1919))
	assert.Equal(t, "Full HD", GetScreenClass(2559))
	assert.Equal(t, "WQHD", GetScreenClass(3839))
	assert.Equal(t, "UHD 4K", GetScreenClass(5119))
	assert.Equal(t, "UHD 5K", GetScreenClass(5120))
}
