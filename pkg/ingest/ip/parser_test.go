package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseForwardedHeader(t *testing.T) {
	header := []string{
		"for=12.34.56.78;host=example.com;proto=https, for=23.45.67.89",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=65.182.89.102",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=10.1.2.3",
		"for=192.0.2.60;proto=http;by=203.0.113.43",
		"for=10.1.2.3;proto=http;by=203.0.113.43",
		"proto=http;by=203.0.113.43;for=192.0.2.61",
		"proto=http;by=203.0.113.43;for=10.1.2.3",
		"   ",
		"",
	}
	expected := []string{
		"23.45.67.89",
		"65.182.89.102",
		"",
		"192.0.2.60",
		"",
		"192.0.2.61",
		"",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], ParseForwardedHeader(head))
	}
}

func TestParseXForwardedForHeader(t *testing.T) {
	header := []string{
		"65.182.89.102",
		"127.0.0.1, 23.21.45.67, 65.182.89.102",
		"127.0.0.1,23.21.45.67,65.182.89.102",
		"65.182.89.102,23.21.45.67,127.0.0.1",
		"   ",
		"",
	}
	expected := []string{
		"65.182.89.102",
		"65.182.89.102",
		"65.182.89.102",
		"",
		"",
		"",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], ParseXForwardedForHeader(head))
	}
}

func TestParseXRealIPHeader(t *testing.T) {
	header := []string{
		"",
		"  ",
		"invalid",
		"127.0.0.1",
		"65.182.89.102",
	}
	expected := []string{
		"",
		"",
		"",
		"",
		"65.182.89.102",
	}

	for i, head := range header {
		assert.Equal(t, expected[i], ParseXRealIPHeader(head))
	}
}

func TestIsValidIP(t *testing.T) {
	assert.False(t, isValidIP("invalid"))
	assert.False(t, isValidIP(""))
	assert.False(t, isValidIP("  "))
	assert.False(t, isValidIP("127.0.0.1"))
	assert.False(t, isValidIP("0.0.0.0"))
	assert.True(t, isValidIP("1.2.3.4"))
}
