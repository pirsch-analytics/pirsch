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

func TestAddAverage(t *testing.T) {
	io := []struct {
		oldAverage int
		newAverage int
		newSize    int
		expected   int
	}{
		{0, 0, 0, 0},
		{42, 42, 5, 42},
		{42, 42, 10, 42},
		{42, 68, 10, 44},
		{68, 42, 10, 66},
	}

	for _, in := range io {
		assert.Equal(t, in.expected, addAverage(in.oldAverage, in.newAverage, in.newSize))
	}
}
