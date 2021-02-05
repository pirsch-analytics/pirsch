package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRunAtMidnight(t *testing.T) {
	cancel := RunAtMidnight(func() {
		panic("Function must not be called")
	})
	cancel()
}

func TestNewTenantID(t *testing.T) {
	assert.False(t, NewTenantID(-1).Valid)
	assert.False(t, NewTenantID(0).Valid)
	assert.True(t, NewTenantID(42).Valid)
}

func TestContainsString(t *testing.T) {
	list := []string{"a", "b", "c", "d"}
	assert.False(t, containsString(list, "e"))
	assert.True(t, containsString(list, "c"))
}

func TestHourInTimezone(t *testing.T) {
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	assert.Equal(t, 6, hourInTimezone(5, tz))
}
