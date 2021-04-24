package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTenantID(t *testing.T) {
	assert.False(t, NewTenantID(-1).Valid)
	assert.False(t, NewTenantID(0).Valid)
	assert.True(t, NewTenantID(42).Valid)
}
