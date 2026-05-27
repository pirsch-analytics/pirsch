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
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.BounceRate{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	// TODO
}

func TestQueryFromPageViews(t *testing.T) {
	q := NewQuery(client)
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.PageViews{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Path{},
		},
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	// TODO
}

func TestQueryFromEvents(t *testing.T) {
	q := NewQuery(client)
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Event{},
		},
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	// TODO
}

func TestQueryFromSessionsFiltered(t *testing.T) {
	q := NewQuery(client)
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
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
				Dimension: dimensions.EntryPath{},
				Values:    []any{"/"},
			},
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
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Equal(t, pkg.TableSessions, q.primaryFilter[0].table)
	assert.Equal(t, dimensions.EntryPath{}, q.primaryFilter[0].filter.Dimension)
	assert.Equal(t, []any{"/"}, q.primaryFilter[0].filter.Values)
	assert.Len(t, q.subqueryFilter, 1)
	assert.Equal(t, pkg.TablePageViews, q.subqueryFilter[0].table)
	assert.Equal(t, request.OperatorOr, q.subqueryFilter[0].filter.Operator)
	assert.Len(t, q.subqueryFilter[0].filter.Filter, 2)
	assert.Equal(t, request.OperatorIs, q.subqueryFilter[0].filter.Filter[0].Operator)
	assert.Equal(t, request.OperatorIs, q.subqueryFilter[0].filter.Filter[1].Operator)
	assert.Equal(t, dimensions.Path{}, q.subqueryFilter[0].filter.Filter[0].Dimension)
	assert.Equal(t, dimensions.Referrer{}, q.subqueryFilter[0].filter.Filter[1].Dimension)
	assert.Equal(t, []any{"/pricing", "/about"}, q.subqueryFilter[0].filter.Filter[0].Values)
	assert.Equal(t, []any{"https://duckduckgo.com"}, q.subqueryFilter[0].filter.Filter[1].Values)
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 9)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, "/", args[3])
	assert.Equal(t, uint64(1), args[4])
	assert.Equal(t, from, args[5])
	assert.Equal(t, to, args[6])
	assert.Equal(t, []any{"/pricing", "/about"}, args[7])
	assert.Equal(t, []any{"https://duckduckgo.com"}, args[8])
	// TODO
}

func TestQueryFromPageViewsFiltered(t *testing.T) {
	q := NewQuery(client)
	req := request.Request{
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
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Empty(t, q.subqueryFilter)
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 4)
	// TODO
}

func TestQueryFromAllFiltered(t *testing.T) {
	q := NewQuery(client)
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
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
	}
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Len(t, q.subqueryFilter, 1)
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 8)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, "/", args[3])
	assert.Equal(t, uint64(1), args[4])
	assert.Equal(t, from, args[5])
	assert.Equal(t, to, args[6])
	assert.Equal(t, "CTA Clicked", args[7])
	// TODO
}

func TestQueryTimeOnPage(t *testing.T) {
	// TODO
}
