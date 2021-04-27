package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFilter_Validate(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.validate()
	assert.NotNil(t, filter)
	assert.Zero(t, filter.From)
	assert.Zero(t, filter.To)
	filter = &Filter{From: pastDay(2), To: pastDay(5)}
	filter.validate()
	assert.Equal(t, pastDay(5), filter.From)
	assert.Equal(t, pastDay(2), filter.To)
	filter = &Filter{From: pastDay(2), To: Today().Add(time.Hour * 24 * 5)}
	filter.validate()
	assert.Equal(t, pastDay(2), filter.From)
	assert.Equal(t, Today(), filter.To)
	filter = &Filter{Day: time.Now()}
	filter.validate()
	assert.Zero(t, filter.Day.Hour())
	assert.Zero(t, filter.Day.Minute())
	assert.Zero(t, filter.Day.Second())
	assert.Zero(t, filter.Day.Nanosecond())
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}
