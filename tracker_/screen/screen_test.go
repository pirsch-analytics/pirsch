package screen

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetClass(t *testing.T) {
	assert.Equal(t, "", GetClass(0))
	assert.Equal(t, "XS", GetClass(42))
	assert.Equal(t, "XL", GetClass(1024))
	assert.Equal(t, "XL", GetClass(1025))
	assert.Equal(t, "HD", GetClass(1919))
	assert.Equal(t, "Full HD", GetClass(2559))
	assert.Equal(t, "WQHD", GetClass(3839))
	assert.Equal(t, "UHD 4K", GetClass(5119))
	assert.Equal(t, "UHD 5K", GetClass(5120))
}
