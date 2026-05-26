package query

import (
	"testing"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
	"github.com/stretchr/testify/assert"
)

func TestQueryFromSessions(t *testing.T) {
	q := NewQuery(client)
	r := q.Run(request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			To:       time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC),
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.BounceRate{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
	})
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.joinTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)
	// TODO
}

func TestQueryFromPageViews(t *testing.T) {
	q := NewQuery(client)
	r := q.Run(request.Request{
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
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Empty(t, q.joinTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)
	// TODO
}

func TestQueryFromEvents(t *testing.T) {
	q := NewQuery(client)
	r := q.Run(request.Request{
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
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Empty(t, q.joinTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)
	// TODO
}

func TestQueryFromSessionsFiltered(t *testing.T) {
	q := NewQuery(client)
	r := q.Run(request.Request{
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
				Operator: request.OperatorOr,
				Filter: []request.Filter{
					{
						Operator:  request.OperatorIs,
						Dimension: dimensions.Path{},
						Values:    []any{"/pricing", "/about"},
					},
					{
						Operator:  request.OperatorIs,
						Dimension: dimensions.Referrer{},
						Values:    []any{"https://duckduckgo.com"},
					},
				},
			},
		},
	})
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.joinTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Empty(t, q.subqueryFilter)
	// TODO
}

func TestQueryFromPageViewsFiltered(t *testing.T) {
	q := NewQuery(client)
	r := q.Run(request.Request{
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
				Values:    []any{"/"},
			},
		},
	})
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Empty(t, q.joinTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Empty(t, q.subqueryFilter)
	// TODO
}

func TestQueryFromAllFiltered(t *testing.T) {
	q := NewQuery(client)
	r := q.Run(request.Request{
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
				Values:    []any{"/"},
			},
			{
				Dimension: dimensions.Event{},
				Values:    []any{"CTA Clicked"},
			},
		},
	})
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.joinTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Len(t, q.subqueryFilter, 1)
	// TODO
}

func TestQueryTimeOnPage(t *testing.T) {
	// TODO
}
