package request

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPeriodValidate(t *testing.T) {
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)

	// defaults and swapping
	p := Period{
		From:        to,
		To:          from,
		WeekdayMode: 5,
		Compare: &ComparePeriod{
			From: to,
			To:   from,
		},
	}
	assert.NoError(t, p.validate())
	assert.Equal(t, time.UTC, p.Timezone)
	assert.Equal(t, WeekdayMonday, p.WeekdayMode)
	assert.Equal(t, from, p.From)
	assert.Equal(t, to, p.To)
	assert.Equal(t, from, p.Compare.From)
	assert.Equal(t, to, p.Compare.To)

	// open-ended periods are not swapped
	p = Period{From: from}
	assert.NoError(t, p.validate())
	assert.Equal(t, from, p.From)
	assert.True(t, p.To.IsZero())
	p = Period{To: to}
	assert.NoError(t, p.validate())
	assert.True(t, p.From.IsZero())
	assert.Equal(t, to, p.To)
}

func TestPeriodValidateTimezone(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)

	for _, tz := range []*time.Location{time.UTC, berlin} {
		p := Period{Timezone: tz}
		assert.NoError(t, p.validate(), tz.String())
	}

	for _, tz := range []*time.Location{
		time.Local,
		time.FixedZone("UTC+2", 7200),
		time.FixedZone("", 0),
		time.FixedZone("UTC', 'x", 0),
	} {
		p := Period{Timezone: tz}
		assert.Error(t, p.validate(), tz.String())
	}
}

func TestFunnelRequestValidatePeriod(t *testing.T) {
	r := FunnelRequest{
		SiteID: 1,
		Period: Period{
			WeekdayMode: 5,
			Timezone:    time.Local,
		},
		Filter: [][]Filter{{{}}, {{}}},
	}
	errs := r.Validate()
	assert.Len(t, errs, 1)
	assert.Equal(t, WeekdayMonday, r.Period.WeekdayMode)
}
