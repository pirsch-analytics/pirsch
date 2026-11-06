package query

import (
	"testing"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
	"github.com/stretchr/testify/assert"
)

func TestQueryTimezoneDay(t *testing.T) {
	loadTestData(t, []string{
		"hourly timezone",
	})
	berlin, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	losAngeles, err := time.LoadLocation("America/Los_Angeles")
	assert.NoError(t, err)

	for _, testCase := range []struct {
		tz       *time.Location
		from, to time.Time
		expected []uint64
	}{
		{
			tz:       time.UTC,
			from:     time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2026, time.January, 17, 0, 0, 0, 0, time.UTC),
			expected: []uint64{1, 3, 1},
		},
		{
			// east of UTC: 23:00 UTC is the next day
			tz:       berlin,
			from:     time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2026, time.January, 17, 0, 0, 0, 0, time.UTC),
			expected: []uint64{1, 2, 2},
		},
		{
			// west of UTC: early UTC hours are the previous day
			tz:       losAngeles,
			from:     time.Date(2026, time.January, 14, 0, 0, 0, 0, time.UTC),
			to:       time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC),
			expected: []uint64{1, 1, 3},
		},
		{
			// the location of From and To is ignored, only the wall clock date is used
			tz:       berlin,
			from:     time.Date(2026, time.January, 15, 0, 0, 0, 0, berlin),
			to:       time.Date(2026, time.January, 17, 0, 0, 0, 0, berlin),
			expected: []uint64{1, 2, 2},
		},
	} {
		q := NewQuery(client)
		req := request.Request{
			SiteID: 1,
			Period: request.Period{
				From:     testCase.from,
				To:       testCase.to,
				Timezone: testCase.tz,
			},
			Dimensions: []dimensions.Dimension{
				dimensions.Day{},
			},
			Metrics: []metrics.Metric{
				metrics.Visitors{},
			},
			OrderBy: []request.OrderBy{
				{
					Dimension: dimensions.Day{},
					Direction: request.DirectionASC,
				},
			},
		}
		r := q.Run(req)
		assert.Empty(t, r.Meta.Errors)
		assert.Len(t, r.Results, len(testCase.expected), testCase.tz.String())

		for i, result := range r.Results {
			day := testCase.from.AddDate(0, 0, i)
			assert.Equal(t, time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC), result.DimensionValues[0], testCase.tz.String())
			assert.Equal(t, testCase.expected[i], result.MetricValues[0], testCase.tz.String())
		}
	}
}

func TestQueryTimezoneHourMultipleDays(t *testing.T) {
	loadTestData(t, []string{
		"hourly timezone",
	})
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	q := NewQuery(client)
	from := time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       from.AddDate(0, 0, 1),
			Timezone: tz,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Hour{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.Hour{},
				Direction: request.DirectionASC,
			},
		},
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)

	// both days must be filled completely, including the hours of the last day
	assert.Len(t, r.Results, 48)
	start := time.Date(2026, time.January, 16, 0, 0, 0, 0, tz)

	for i, result := range r.Results {
		hour := start.Add(time.Hour * time.Duration(i))
		assert.True(t, hour.Equal(result.DimensionValues[0].(time.Time)), i)

		// 01:00, 12:00, 23:00 UTC on the 16th and 00:30 UTC on the 17th
		if i == 2 || i == 13 || i == 24 || i == 25 {
			assert.Equal(t, uint64(1), result.MetricValues[0], i)
		} else {
			assert.Equal(t, uint64(0), result.MetricValues[0], i)
		}
	}
}

func TestQueryTimezoneMinuteSingleDay(t *testing.T) {
	loadTestData(t, []string{
		"hourly timezone",
	})
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	q := NewQuery(client)
	from := time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       from,
			Timezone: tz,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Minute{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.Minute{},
				Direction: request.DirectionASC,
			},
		},
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)

	// the whole day must be filled, not just the first hour
	assert.Len(t, r.Results, 24*60)
	assert.Equal(t, uint64(1), r.Results[2*60].MetricValues[0])
	assert.Equal(t, uint64(1), r.Results[13*60].MetricValues[0])
}

func TestQueryTimezoneDaylightSavingTime(t *testing.T) {
	// no data, only the filled rows are checked
	loadTestData(t, []string{"none"})
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)

	for _, testCase := range []struct {
		day   time.Time
		hours int
	}{
		{time.Date(2026, time.March, 29, 0, 0, 0, 0, time.UTC), 23},
		{time.Date(2026, time.October, 25, 0, 0, 0, 0, time.UTC), 25},
		{time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), 24},
	} {
		q := NewQuery(client)
		req := request.Request{
			SiteID: 1,
			Period: request.Period{
				From:     testCase.day,
				To:       testCase.day,
				Timezone: tz,
			},
			Dimensions: []dimensions.Dimension{
				dimensions.Hour{},
			},
			Metrics: []metrics.Metric{
				metrics.Visitors{},
			},
			OrderBy: []request.OrderBy{
				{
					Dimension: dimensions.Hour{},
					Direction: request.DirectionASC,
				},
			},
		}
		r := q.Run(req)
		assert.Empty(t, r.Meta.Errors)
		assert.Len(t, r.Results, testCase.hours, testCase.day.String())
	}
}

func TestQueryTimezoneWeekAndWeekday(t *testing.T) {
	loadTestData(t, []string{
		"hourly timezone",
	})
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)

	// 2026-01-17 is a Saturday, 2026-01-18 a Sunday
	for _, testCase := range []struct {
		mode      request.WeekdayMode
		weekStart time.Time
		weekdays  map[uint8]uint64
	}{
		{
			mode:      request.WeekdayMonday,
			weekStart: time.Date(2026, time.January, 12, 0, 0, 0, 0, time.UTC),
			weekdays:  map[uint8]uint64{3: 1, 4: 2, 5: 2}, // Thursday, Friday, Saturday (0 = Monday)
		},
		{
			mode:      request.WeekdaySunday,
			weekStart: time.Date(2026, time.January, 11, 0, 0, 0, 0, time.UTC),
			weekdays:  map[uint8]uint64{4: 1, 5: 2, 6: 2}, // Thursday, Friday, Saturday (0 = Sunday)
		},
	} {
		req := request.Request{
			SiteID: 1,
			Period: request.Period{
				From:        time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC),
				To:          time.Date(2026, time.January, 17, 0, 0, 0, 0, time.UTC),
				Timezone:    tz,
				WeekdayMode: testCase.mode,
			},
			Dimensions: []dimensions.Dimension{
				dimensions.Week{},
			},
			Metrics: []metrics.Metric{
				metrics.Visitors{},
			},
		}
		r := NewQuery(client).Run(req)
		assert.Empty(t, r.Meta.Errors)
		assert.Len(t, r.Results, 1)
		assert.Equal(t, testCase.weekStart, r.Results[0].DimensionValues[0])
		assert.Equal(t, uint64(5), r.Results[0].MetricValues[0])

		req.Dimensions = []dimensions.Dimension{dimensions.Weekday{}}
		r = NewQuery(client).Run(req)
		assert.Empty(t, r.Meta.Errors)
		assert.Len(t, r.Results, len(testCase.weekdays))

		for _, result := range r.Results {
			weekday := result.DimensionValues[0].(uint8)
			assert.Equal(t, testCase.weekdays[weekday], result.MetricValues[0], weekday)
		}
	}
}

func TestQueryTimezoneIncludeTime(t *testing.T) {
	loadTestData(t, []string{
		"hourly timezone",
	})
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)

	// 23:30 to 01:00 Berlin time includes 23:00 UTC on the 16th (00:00 Berlin) only,
	// as 00:30 UTC on the 17th is 01:30 Berlin time
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:        time.Date(2026, time.January, 16, 23, 30, 0, 0, time.UTC),
			To:          time.Date(2026, time.January, 17, 1, 0, 0, 0, time.UTC),
			Timezone:    tz,
			IncludeTime: true,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
	}
	r := NewQuery(client).Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Len(t, r.Results, 1)
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
}

func TestQueryTimezoneCompare(t *testing.T) {
	loadTestData(t, []string{
		"hourly timezone",
	})
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 17, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 17, 0, 0, 0, 0, time.UTC),
			Timezone: tz,
			Compare: &request.ComparePeriod{
				From: time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC),
			},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
	}
	r := NewQuery(client).Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Len(t, r.Results, 1)
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(2), r.Results[0].CompareMetricValues[0])
}

func TestQueryTimezoneFunnel(t *testing.T) {
	loadTestData(t, []string{
		"hourly timezone",
	})
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	day := time.Date(2026, time.January, 17, 0, 0, 0, 0, time.UTC)
	r := NewQuery(client).Funnel(request.FunnelRequest{
		SiteID: 1,
		Period: request.Period{
			From:     day,
			To:       day,
			Timezone: tz,
		},
		Filter: [][]request.Filter{
			{{Dimension: dimensions.Path{}, Values: []any{"/"}}},
			{{Dimension: dimensions.Path{}, Values: []any{"/"}}},
		},
	})
	assert.Empty(t, r.Meta.Errors)
	assert.Len(t, r.Steps, 2)
	assert.Equal(t, uint64(2), r.Steps[0].Visitors)
}

func TestQueryInvalidTimezone(t *testing.T) {
	for _, tz := range []*time.Location{
		time.Local,
		time.FixedZone("UTC+2", 7200),
		time.FixedZone("", 0),
	} {
		r := NewQuery(client).Run(request.Request{
			SiteID: 1,
			Period: request.Period{
				Timezone: tz,
			},
		})
		assert.Len(t, r.Meta.Errors, 1, tz.String())
	}
}

func TestQueryImportedNoOverlap(t *testing.T) {
	loadTestData(t, []string{
		"imported visitors",
	})

	// imported statistics exist for Dec 29-31, but must only be used until the day before ImportedUntil
	options := &request.Options{
		IncludeImportedStatistics: true,
	}
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:          time.Date(2025, time.December, 29, 0, 0, 0, 0, time.UTC),
			To:            time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			ImportedUntil: time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC),
		},
		Metrics: []metrics.Metric{
			metrics.PageViews{},
			metrics.RelativeViews{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.Day{},
				Direction: request.DirectionASC,
			},
		},
		Options: options,
	}
	r := NewQuery(client).Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Len(t, r.Results, 4)
	assert.Equal(t, uint64(9), r.Results[0].MetricValues[0])
	assert.InDelta(t, 0.6, r.Results[0].MetricValues[1], 0.001)
	assert.Equal(t, uint64(6), r.Results[1].MetricValues[0])
	assert.InDelta(t, 0.4, r.Results[1].MetricValues[1], 0.001)
	assert.Equal(t, uint64(0), r.Results[2].MetricValues[0])
	assert.Equal(t, uint64(0), r.Results[3].MetricValues[0])

	// the options passed in must not be modified
	assert.True(t, options.IncludeImportedStatistics)
}
