package query

import (
	"testing"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
)

func TestQueryFromSessions(t *testing.T) {
	q := NewQuery(client)
	q.Run(request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
	})
	// TODO
}

func TestQueryFromPageViews(t *testing.T) {
	q := NewQuery(client)
	q.Run(request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.PageViews{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Path{},
		},
	})
	// TODO
}

func TestQueryFromEvents(t *testing.T) {
	q := NewQuery(client)
	q.Run(request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Event{},
		},
	})
	// TODO
}

func TestQueryFromSessionsFiltered(t *testing.T) {
	q := NewQuery(client)
	q.Run(request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.Path{},
				Values:    []string{"/"},
			},
		},
	})
	// TODO
}

func TestQueryFromPageViewsFiltered(t *testing.T) {
	q := NewQuery(client)
	q.Run(request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.PageViews{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Path{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.Path{},
				Values:    []string{"/"},
			},
		},
	})
	// TODO
}

func TestQueryFromAllFiltered(t *testing.T) {
	q := NewQuery(client)
	q.Run(request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.Entries{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.EntryPath{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.EntryPath{},
				Values:    []string{"/"},
			},
			{
				Dimension: dimensions.Event{},
				Values:    []string{"CTA Clicked"},
			},
		},
	})
	// TODO
}
