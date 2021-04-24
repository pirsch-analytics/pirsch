package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFilter_Days(t *testing.T) {
	filter := NewFilter(NullTenant)
	assert.Equal(t, 6, filter.Days()) // the default filter covers the past week NOT including today
	filter.From = pastDay(20)
	filter.To = Today()
	filter.validate()
	assert.Equal(t, 20, filter.Days())
}

func TestFilter_Validate(t *testing.T) {
	filter := NewFilter(NullTenant)
	filter.validate()
	assert.NotNil(t, filter)
	assert.Equal(t, pastDay(6), filter.From)
	assert.Equal(t, pastDay(0), filter.To)
	filter = &Filter{From: pastDay(2), To: pastDay(5)}
	filter.validate()
	assert.Equal(t, pastDay(5), filter.From)
	assert.Equal(t, pastDay(2), filter.To)
	filter = &Filter{From: pastDay(2), To: Today().Add(time.Hour * 24 * 5)}
	filter.validate()
	assert.Equal(t, pastDay(2), filter.From)
	assert.Equal(t, Today(), filter.To)
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}
