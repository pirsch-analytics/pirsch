package query

import (
	"context"
	"encoding/json"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
	"github.com/stretchr/testify/assert"
)

/*
	TODO

	- Growth
	- Comparison Mode
	- Session Breakdown
	- Funnel
*/

func TestQueryFromSessions(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.PageViews{},
			metrics.Sessions{},
			metrics.Bounces{},
			metrics.BounceRate{},
			metrics.AvgSessionDuration{},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result dimensions
	assert.Len(t, r.Results, 2)
	assert.Equal(t, from, r.Results[0].DimensionValues[0])

	// result metrics row 0
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(3), r.Results[0].MetricValues[1])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[2])
	assert.Equal(t, int64(1), r.Results[0].MetricValues[3])
	assert.Equal(t, 0.5, r.Results[0].MetricValues[4])
	assert.InDelta(t, 150, r.Results[0].MetricValues[5], 0.001)

	// result metrics row 1
	assert.Equal(t, from.Add(time.Hour*24), r.Results[1].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[1].MetricValues[0])
	assert.Equal(t, uint64(5), r.Results[1].MetricValues[1])
	assert.Equal(t, uint64(3), r.Results[1].MetricValues[2])
	assert.Equal(t, int64(2), r.Results[1].MetricValues[3])
	assert.InDelta(t, 0.6666, r.Results[1].MetricValues[4], 0.001)
	assert.InDelta(t, 60, r.Results[1].MetricValues[5], 0.001)
}

func TestQueryFromPageViews(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Path{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.RelativeVisitors{},
			metrics.PageViews{},
			metrics.RelativeViews{},
			metrics.Sessions{},
			metrics.Bounces{},
			metrics.BounceRate{},
			metrics.AvgTimeOnPage{},
		},
		OrderBy: []request.OrderBy{
			{Metric: metrics.PageViews{}},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable) // overwritten for subquery
	assert.Equal(t, pkg.TableSessions, q.joinTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 12)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, uint64(1), args[3])
	assert.Equal(t, from, args[4])
	assert.Equal(t, to, args[5])
	assert.Equal(t, uint64(1), args[6])
	assert.Equal(t, from, args[7])
	assert.Equal(t, to, args[8])
	assert.Equal(t, uint64(1), args[9])
	assert.Equal(t, from, args[10])
	assert.Equal(t, to, args[11])

	// result
	assert.Len(t, r.Results, 3)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 8)
	assert.Len(t, r.Results[1].DimensionValues, 1)
	assert.Len(t, r.Results[1].MetricValues, 8)
	assert.Len(t, r.Results[2].DimensionValues, 1)
	assert.Len(t, r.Results[2].MetricValues, 8)

	// result dimensions
	assert.Equal(t, "/", r.Results[0].DimensionValues[0])
	assert.Equal(t, "/pricing", r.Results[1].DimensionValues[0])
	assert.Equal(t, "/landing", r.Results[2].DimensionValues[0])

	// result metrics row 0
	assert.Equal(t, uint64(4), r.Results[0].MetricValues[0])
	assert.InDelta(t, 1, r.Results[0].MetricValues[1], 0.001)
	assert.Equal(t, uint64(5), r.Results[0].MetricValues[2])
	assert.InDelta(t, 0.625, r.Results[0].MetricValues[3], 0.001)
	assert.Equal(t, uint64(5), r.Results[0].MetricValues[4])
	assert.Equal(t, int64(3), r.Results[0].MetricValues[5])
	assert.InDelta(t, 0.75, r.Results[0].MetricValues[6], 0.001)
	assert.InDelta(t, float64(300), r.Results[0].MetricValues[7], 0.001)

	// result metrics row 1
	assert.Equal(t, uint64(2), r.Results[1].MetricValues[0])
	assert.InDelta(t, 0.5, r.Results[1].MetricValues[1], 0.001)
	assert.Equal(t, uint64(2), r.Results[1].MetricValues[2])
	assert.InDelta(t, 0.25, r.Results[1].MetricValues[3], 0.001)
	assert.Equal(t, uint64(2), r.Results[1].MetricValues[4])
	assert.Equal(t, int64(0), r.Results[1].MetricValues[5])
	assert.InDelta(t, 0, r.Results[1].MetricValues[6], 0.001)
	assert.InDelta(t, float64(60), r.Results[1].MetricValues[7], 0.001)

	// result metrics row 2
	assert.Equal(t, uint64(1), r.Results[2].MetricValues[0])
	assert.InDelta(t, 0.25, r.Results[2].MetricValues[1], 0.001)
	assert.Equal(t, uint64(1), r.Results[2].MetricValues[2])
	assert.InDelta(t, 0.125, r.Results[2].MetricValues[3], 0.001)
	assert.Equal(t, uint64(1), r.Results[2].MetricValues[4])
	assert.Equal(t, int64(0), r.Results[2].MetricValues[5])
	assert.InDelta(t, 0, r.Results[2].MetricValues[6], 0.001)
	assert.InDelta(t, float64(120), r.Results[2].MetricValues[7], 0.001)
}

func TestQueryFromEvents(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Event{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.Events{},
			metrics.CR{},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 6)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, uint64(1), args[3])
	assert.Equal(t, from, args[4])
	assert.Equal(t, to, args[5])

	// result
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 3)
	assert.Equal(t, "Contact Button", r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(3), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(3), r.Results[0].MetricValues[1])
	assert.InDelta(t, 0.75, r.Results[0].MetricValues[2], 0.001)
}

func TestQueryFromSessionsFiltered(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
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
						Values:    []any{"/pricing", "/landing"},
					},
					{
						Operator:  request.OperatorIs,
						Dimension: dimensions.Referrer{},
						Values:    []any{"https://duckduckgo.com"},
					},
				},
			},
			{
				Dimension: dimensions.Platform{},
				Values:    []any{0, 1},
			},
		},
	}

	// tables and filter
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Len(t, q.primaryFilter, 2)
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
	assert.Equal(t, []any{"/pricing", "/landing"}, q.subqueryFilter[0].filter.Filter[0].Values)
	assert.Equal(t, []any{"https://duckduckgo.com"}, q.subqueryFilter[0].filter.Filter[1].Values)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 10)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, "/", args[3])
	assert.Equal(t, []any{0, 1}, args[4])
	assert.Equal(t, uint64(1), args[5])
	assert.Equal(t, from, args[6])
	assert.Equal(t, to, args[7])
	assert.Equal(t, []any{"/pricing", "/landing"}, args[8])
	assert.Equal(t, "https://duckduckgo.com", args[9])

	// result
	assert.Len(t, r.Results, 2)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)
	assert.Len(t, r.Results[1].DimensionValues, 1)
	assert.Len(t, r.Results[1].MetricValues, 1)
	assert.Equal(t, from, r.Results[0].DimensionValues[0])
	assert.Equal(t, from.Add(time.Hour*24), r.Results[1].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
}

func TestQueryFromPageViewsFiltered(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Path{},
		},
		Metrics: []metrics.Metric{
			metrics.PageViews{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.Path{},
				Values:    []any{"/"},
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 4)

	// result
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)
	assert.Equal(t, "/", r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(5), r.Results[0].MetricValues[0])
}

func TestQueryFromEventsFiltered(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.EntryPath{},
		},
		Metrics: []metrics.Metric{
			metrics.Entries{},
			metrics.Visitors{},
			metrics.EntryRate{},
			metrics.AvgTimeOnPage{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.EntryPath{},
				Values:    []any{"/"},
			},
			{
				Dimension: dimensions.Event{},
				Values:    []any{"Contact Button"},
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Len(t, q.subqueryFilter, 1)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 18)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, uint64(1), args[3])
	assert.Equal(t, from, args[4])
	assert.Equal(t, to, args[5])
	assert.Equal(t, "Contact Button", args[6])
	assert.Equal(t, uint64(1), args[7])
	assert.Equal(t, from, args[8])
	assert.Equal(t, to, args[9])
	assert.Equal(t, uint64(1), args[10])
	assert.Equal(t, from, args[11])
	assert.Equal(t, to, args[12])
	assert.Equal(t, "/", args[13])
	assert.Equal(t, uint64(1), args[14])
	assert.Equal(t, from, args[15])
	assert.Equal(t, to, args[16])
	assert.Equal(t, "Contact Button", args[17])

	// result
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 4)
	assert.Equal(t, "/", r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[1])
	assert.InDelta(t, 0.4, r.Results[0].MetricValues[2], 0.001)
	assert.InDelta(t, 0, r.Results[0].MetricValues[3], 0.001)
}

func TestQueryDimensionOnly(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Referrer{},
		},
		Filter: []request.Filter{
			{
				Operator:  request.OperatorContains,
				Dimension: dimensions.Referrer{},
				Values:    []any{"go"},
			},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.Referrer{},
				Direction: request.DirectionASC,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Len(t, q.subqueryFilter, 0)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 4)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, "%go%", args[3])

	// result
	assert.Len(t, r.Results, 2)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 0)
	assert.Equal(t, "https://duckduckgo.com", r.Results[0].DimensionValues[0])
	assert.Equal(t, "https://google.com", r.Results[1].DimensionValues[0])
}

func TestQueryTime(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, _, _ := newQuery()
	from := time.Date(2026, time.January, 2, 9, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 2, 9, 30, 0, 0, time.UTC)
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:        from,
			To:          to,
			Timezone:    time.UTC,
			IncludeTime: true,
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result
	assert.Len(t, r.Results, 1)
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
}

func TestQueryLimit(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.Bounces{},
			metrics.BounceRate{},
		},
		Pagination: &request.Pagination{
			Limit: 1,
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result
	assert.Len(t, r.Results, 1)
	assert.Equal(t, from, r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, int64(1), r.Results[0].MetricValues[1])
	assert.Equal(t, 0.5, r.Results[0].MetricValues[2])
}

func TestQueryOffsetLimit(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.Bounces{},
			metrics.BounceRate{},
		},
		Pagination: &request.Pagination{
			Offset: 1,
			Limit:  1,
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result
	assert.Len(t, r.Results, 1)
	assert.Equal(t, from.Add(time.Hour*24), r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, int64(2), r.Results[0].MetricValues[1])
	assert.InDelta(t, 0.6666, r.Results[0].MetricValues[2], 0.001)
}

func TestQueryEventList(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Event{},
			dimensions.EventMeta{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.Events{},
		},
		OrderBy: []request.OrderBy{
			{
				Metric: metrics.Events{},
			},
			{
				Dimension: dimensions.Event{},
				Direction: request.DirectionASC,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result dimensions and metrics
	assert.Len(t, r.Results, 3)
	assert.Len(t, r.Results[0].DimensionValues, 2)
	assert.Len(t, r.Results[0].MetricValues, 2)
	assert.Len(t, r.Results[1].DimensionValues, 2)
	assert.Len(t, r.Results[1].MetricValues, 2)

	// result row 0
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[1])
	assert.Equal(t, "Contact Button", r.Results[0].DimensionValues[0])
	assert.Equal(t, `{"label":"Get in touch","position":"text","price":67.9}`, r.Results[0].DimensionValues[1])

	// result row 1
	assert.Equal(t, uint64(1), r.Results[1].MetricValues[0])
	assert.Equal(t, uint64(1), r.Results[1].MetricValues[1])
	assert.Equal(t, "Contact Button", r.Results[1].DimensionValues[0])
	assert.Equal(t, `{"ab-test":[2,5],"position":"hero","price":99.54}`, r.Results[1].DimensionValues[1])

	// result row 2
	assert.Equal(t, uint64(1), r.Results[2].MetricValues[0])
	assert.Equal(t, uint64(1), r.Results[2].MetricValues[1])
	assert.Equal(t, "Contact Button", r.Results[2].DimensionValues[0])
	assert.Equal(t, `{"label":"Get in touch","position":"text","price":24.99}`, r.Results[2].DimensionValues[1])
}

func TestQueryEventMetaDataFilterKey(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Event{},
			dimensions.EventMeta{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			//metrics.PageViews{}, // TODO does this still make sense?
			metrics.CR{},
			metrics.Events{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.Event{},
				Values:    []any{"Contact Button"},
			},
			{
				Dimension: dimensions.EventMetaKey{},
				Values:    []any{"ab-test.0"},
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Len(t, q.primaryFilter, 2)
	assert.Equal(t, []any{"Contact Button"}, q.primaryFilter[0].filter.Values)
	assert.Equal(t, []any{"ab-test.0"}, q.primaryFilter[1].filter.Values)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 7)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, uint64(1), args[3])
	assert.Equal(t, from, args[4])
	assert.Equal(t, to, args[5])
	assert.Equal(t, "Contact Button", args[6])

	// result dimensions and metrics
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 2)
	assert.Len(t, r.Results[0].MetricValues, 3)

	// result row
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
	assert.InDelta(t, 0.3333, r.Results[0].MetricValues[1], 0.001)
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[2])
	assert.Equal(t, "Contact Button", r.Results[0].DimensionValues[0])
	assert.Equal(t, `{"ab-test":[2,5],"position":"hero","price":99.54}`, r.Results[0].DimensionValues[1])
}

func TestQueryEventMetaDataFilterValue(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Event{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.EventMeta{
					Path: "position",
				},
				Values: []any{"hero"},
			},
			{
				Operator: request.OperatorIsNot,
				Dimension: dimensions.EventMeta{
					Path: "ab-test.0",
				},
				Values: []any{1},
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Len(t, q.primaryFilter, 2)
	assert.Equal(t, []any{"hero"}, q.primaryFilter[0].filter.Values)
	assert.Equal(t, []any{1}, q.primaryFilter[1].filter.Values)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 5)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, "hero", args[3])
	assert.Equal(t, 1, args[4])

	// result dimensions and metrics
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)

	// result row
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
	assert.Equal(t, "Contact Button", r.Results[0].DimensionValues[0])
}

func TestQueryEventMetaDataFunction(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.EventMeta{
				Path:     "price",
				Type:     dimensions.EventMetaTypeFloat,
				Function: dimensions.EventMetaFunctionSum,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result dimensions and metrics
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Empty(t, r.Results[0].MetricValues)

	// result row
	assert.Equal(t, 192.43, r.Results[0].DimensionValues[0])
}

func TestQueryEventMetaDataCastType(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.EventMeta{
				Path: "price",
				Type: dimensions.EventMetaTypeFloat,
			},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.EventMeta{
					Type: dimensions.EventMetaTypeFloat,
				},
				Direction: request.DirectionDESC,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableEvents, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result dimensions and metrics
	assert.Len(t, r.Results, 3)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Empty(t, r.Results[0].MetricValues)

	// result rows
	assert.Equal(t, 99.54, r.Results[0].DimensionValues[0])
	assert.Equal(t, 67.9, r.Results[1].DimensionValues[0])
	assert.Equal(t, 24.99, r.Results[2].DimensionValues[0])
}

func TestQueryTagKeysList(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"simple",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.TagKey{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.TagKey{},
				Direction: request.DirectionASC,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result dimensions and metrics
	assert.Len(t, r.Results, 2)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)
	assert.Len(t, r.Results[1].DimensionValues, 1)
	assert.Len(t, r.Results[1].MetricValues, 1)

	// result row 0
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
	assert.Equal(t, "ab-test", r.Results[0].DimensionValues[0])

	// result row 1
	assert.Equal(t, uint64(2), r.Results[1].MetricValues[0])
	assert.Equal(t, "author", r.Results[1].DimensionValues[0])
}

func TestQueryTagBreakdown(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"simple",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.TagValue{
				Key: "author",
			},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.PageViews{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.TagKey{},
				Values:    []any{"author"},
			},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.TagValue{
					Key: "author", // TODO validate in request that this matches the dimension (or set it)
				},
				Direction: request.DirectionASC,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Len(t, q.primaryFilter, 1)
	assert.Empty(t, q.subqueryFilter)
	assert.Equal(t, "author", q.primaryFilter[0].filter.Values[0])

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 5)
	assert.Equal(t, "author", args[0])
	assert.Equal(t, uint64(1), args[1])
	assert.Equal(t, from, args[2])
	assert.Equal(t, to, args[3])
	assert.Equal(t, "author", args[4])

	// result dimensions and metrics
	assert.Len(t, r.Results, 2)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 2)
	assert.Len(t, r.Results[1].DimensionValues, 1)
	assert.Len(t, r.Results[1].MetricValues, 2)

	// result row 0
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[1])
	assert.Equal(t, "John Doe", r.Results[0].DimensionValues[0])

	// result row 1
	assert.Equal(t, uint64(1), r.Results[1].MetricValues[0])
	assert.Equal(t, uint64(1), r.Results[1].MetricValues[1])
	assert.Equal(t, "Marvin Blum", r.Results[1].DimensionValues[0])
}

func TestQueryTagFilter(t *testing.T) {
	loadTestData(t, []string{
		"simple bounced + event (non-interactive)",
		"simple",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Channel{},
		},
		Metrics: []metrics.Metric{
			metrics.Visitors{},
			metrics.PageViews{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.TagValue{
					Key: "author",
				},
				Values: []any{"Marvin Blum"},
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Len(t, q.subqueryFilter, 1)
	assert.Equal(t, "Marvin Blum", q.subqueryFilter[0].filter.Values[0])

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 8)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, uint64(1), args[3])
	assert.Equal(t, from, args[4])
	assert.Equal(t, to, args[5])
	assert.Equal(t, "author", args[6])
	assert.Equal(t, "Marvin Blum", args[7])

	// result dimensions and metrics
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 2)

	// result row
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(1), r.Results[0].MetricValues[1])
	assert.Equal(t, "Organic Search", r.Results[0].DimensionValues[0])
}

func TestQueryTimeOnPage(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Path{},
		},
		Metrics: []metrics.Metric{
			metrics.AvgTimeOnPage{},
		},
		Filter: []request.Filter{
			{
				Dimension: dimensions.EntryPath{},
				Values:    []any{"/"},
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Len(t, q.subqueryFilter, 1)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 14)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, uint64(1), args[3])
	assert.Equal(t, from, args[4])
	assert.Equal(t, to, args[5])
	assert.Equal(t, uint64(1), args[7])
	assert.Equal(t, from, args[8])
	assert.Equal(t, to, args[9])
	assert.Equal(t, uint64(1), args[10])
	assert.Equal(t, from, args[11])
	assert.Equal(t, to, args[12])
	assert.Equal(t, "/", args[13])

	// result
	assert.Len(t, r.Results, 2)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)

	// result row 0
	assert.Equal(t, "/pricing", r.Results[0].DimensionValues[0])
	assert.InDelta(t, 0, r.Results[0].MetricValues[0], 0.001)

	// result row 1
	assert.Equal(t, "/", r.Results[1].DimensionValues[0])
	assert.InDelta(t, 300, r.Results[1].MetricValues[0], 0.001)
}

func TestQueryTimeOnPagePerDay(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.Day{},
			dimensions.Path{},
		},
		Metrics: []metrics.Metric{
			metrics.AvgTimeOnPage{},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.Day{},
				Direction: request.DirectionASC,
			},
			{
				Dimension: dimensions.Path{},
				Direction: request.DirectionASC,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TablePageViews, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 6)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, uint64(1), args[3])
	assert.Equal(t, from, args[4])
	assert.Equal(t, to, args[5])

	// result
	assert.Len(t, r.Results, 5)
	assert.Len(t, r.Results[0].DimensionValues, 2)
	assert.Len(t, r.Results[0].MetricValues, 1)

	// result row 0
	assert.Equal(t, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), r.Results[0].DimensionValues[0])
	assert.Equal(t, "/", r.Results[0].DimensionValues[1])
	assert.InDelta(t, 300, r.Results[0].MetricValues[0], 0.001)

	// result row 1
	assert.Equal(t, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), r.Results[1].DimensionValues[0])
	assert.Equal(t, "/pricing", r.Results[1].DimensionValues[1])
	assert.InDelta(t, 0, r.Results[1].MetricValues[0], 0.001)

	// result row 2
	assert.Equal(t, time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC), r.Results[2].DimensionValues[0])
	assert.Equal(t, "/", r.Results[2].DimensionValues[1])
	assert.InDelta(t, 0, r.Results[2].MetricValues[0], 0.001)

	// result row 3
	assert.Equal(t, time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC), r.Results[3].DimensionValues[0])
	assert.Equal(t, "/landing", r.Results[3].DimensionValues[1])
	assert.InDelta(t, 120, r.Results[3].MetricValues[0], 0.001)

	// result row 4
	assert.Equal(t, time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC), r.Results[4].DimensionValues[0])
	assert.Equal(t, "/pricing", r.Results[4].DimensionValues[1])
	assert.InDelta(t, 60, r.Results[4].MetricValues[0], 0.001)
}

func TestQueryListSessions(t *testing.T) {
	loadTestData(t, []string{
		"scenario",
		"simple bounced + event (non-interactive)",
		"simple",
		"three page views + event",
		"referrer reset",
	})
	q, from, to := newQuery()
	req := request.Request{
		SiteID: 1,
		Period: request.Period{
			From:     from,
			To:       to,
			Timezone: time.UTC,
		},
		Dimensions: []dimensions.Dimension{
			dimensions.VisitorID{},
			dimensions.SessionID{},
			dimensions.Start{},
			dimensions.Duration{},
			dimensions.PageViews{},
			dimensions.Bounced{},
			dimensions.EntryPath{},
			dimensions.ExitPath{},
			dimensions.EntryTitle{},
			dimensions.ExitTitle{},
			dimensions.Time{},
			dimensions.Hostname{},
			dimensions.Language{},
			dimensions.Country{},
			dimensions.Region{},
			dimensions.City{},
			dimensions.Referrer{},
			dimensions.ReferrerName{},
			dimensions.ReferrerIcon{},
			dimensions.OS{},
			dimensions.OSVersion{},
			dimensions.Browser{},
			dimensions.BrowserVersion{},
			dimensions.Platform{},
			dimensions.ScreenClass{},
			dimensions.UTMSource{},
			dimensions.UTMContent{},
			dimensions.UTMMedium{},
			dimensions.UTMCampaign{},
			dimensions.UTMTerm{},
			dimensions.Channel{},
		},
		OrderBy: []request.OrderBy{
			{
				Dimension: dimensions.VisitorID{},
				Direction: request.DirectionASC,
			},
		},
	}

	// tables
	r := q.Run(req)
	assert.Empty(t, r.Meta.Errors)
	assert.Equal(t, pkg.TableSessions, q.primaryTable)
	assert.Empty(t, q.primaryFilter)
	assert.Empty(t, q.subqueryFilter)

	// query
	query, args := q.buildQuery(req)
	assert.NotEmpty(t, query)
	assert.Len(t, args, 3)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])

	// result
	assert.Len(t, r.Results, 5)
	assert.Len(t, r.Results[0].DimensionValues, 31)
	assert.Empty(t, r.Results[0].MetricValues)

	// result row 0 (only fully compare this, as this test would get really long otherwise)
	assert.Equal(t, uint64(1), r.Results[0].DimensionValues[0])
	assert.Equal(t, uint32(1), r.Results[0].DimensionValues[1])
	assert.Equal(t, time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC), r.Results[0].DimensionValues[2])
	assert.Equal(t, uint32(0), r.Results[0].DimensionValues[3])
	assert.Equal(t, uint16(1), r.Results[0].DimensionValues[4])
	assert.Equal(t, true, r.Results[0].DimensionValues[5])
	assert.Equal(t, "/", r.Results[0].DimensionValues[6])
	assert.Equal(t, "/", r.Results[0].DimensionValues[7])
	assert.Equal(t, "Home", r.Results[0].DimensionValues[8])
	assert.Equal(t, "Home", r.Results[0].DimensionValues[9])
	assert.Equal(t, time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC), r.Results[0].DimensionValues[10])
	assert.Equal(t, "example.com", r.Results[0].DimensionValues[11])
	assert.Equal(t, "en", r.Results[0].DimensionValues[12])
	assert.Equal(t, "us", r.Results[0].DimensionValues[13])
	assert.Equal(t, "Virginia", r.Results[0].DimensionValues[14])
	assert.Equal(t, "Ashburn", r.Results[0].DimensionValues[15])
	assert.Equal(t, "https://duckduckgo.com", r.Results[0].DimensionValues[16])
	assert.Equal(t, "DuckDuckGo", r.Results[0].DimensionValues[17])
	assert.Empty(t, r.Results[0].DimensionValues[18])
	assert.Equal(t, pkg.OSWindows, r.Results[0].DimensionValues[19])
	assert.Equal(t, "10", r.Results[0].DimensionValues[20])
	assert.Equal(t, pkg.BrowserChrome, r.Results[0].DimensionValues[21])
	assert.Equal(t, "142", r.Results[0].DimensionValues[22])
	assert.Equal(t, pkg.PlatformDesktop, r.Results[0].DimensionValues[23])
	assert.Equal(t, "Full HD", r.Results[0].DimensionValues[24])
	assert.Equal(t, "DuckDuckGo", r.Results[0].DimensionValues[25])
	assert.Equal(t, "Main", r.Results[0].DimensionValues[26])
	assert.Equal(t, "Search", r.Results[0].DimensionValues[27])
	assert.Equal(t, "Paid", r.Results[0].DimensionValues[28])
	assert.Equal(t, "privacy+analytics", r.Results[0].DimensionValues[29])
	assert.Equal(t, "Organic Search", r.Results[0].DimensionValues[30])

	// result row 1
	assert.Equal(t, uint64(2), r.Results[1].DimensionValues[0])
	assert.Equal(t, uint32(2), r.Results[1].DimensionValues[1])

	// result row 2
	assert.Equal(t, uint64(3), r.Results[2].DimensionValues[0])
	assert.Equal(t, uint32(3), r.Results[2].DimensionValues[1])

	// result row 3
	assert.Equal(t, uint64(4), r.Results[3].DimensionValues[0])
	assert.Equal(t, uint32(4), r.Results[3].DimensionValues[1])

	// result row 4
	assert.Equal(t, uint64(4), r.Results[4].DimensionValues[0])
	assert.Equal(t, uint32(5), r.Results[4].DimensionValues[1])
}

func TestQueryListSessionBreakdown(t *testing.T) {
	// TODO
}

func TestBuildQueryFilterJSONPath(t *testing.T) {
	input := []string{
		"field",
		"foo.bar",
		"foo.0.bar.9",
	}
	result := []string{
		`."field"`,
		`."foo"."bar"`,
		`."foo"[1]."bar"[10]`,
	}
	q, _, _ := newQuery()

	for i, in := range input {
		assert.Equal(t, result[i], q.buildQueryFilterJSONPath(in))
	}
}

func newQuery() (*Query, time.Time, time.Time) {
	q := NewQuery(client)
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	return q, from, to
}

type sessionData struct {
	model.Session
	Scenario string `csv:"scenario"`
}

type pageViewData struct {
	model.PageView
	Scenario string `csv:"scenario"`
	Tags     string `csv:"tags"`
}

type eventData struct {
	model.Event
	Scenario string `csv:"scenario"`
	MetaData string `csv:"meta_data"`
}

func loadTestData(t *testing.T, scenarios []string) {
	db.CleanupDB(t, client)

	// load and store sessions
	sessionsFile, err := os.ReadFile("../../../test/sessions.csv")
	assert.NoError(t, err)
	var sessionData []sessionData
	assert.NoError(t, gocsv.UnmarshalBytes(sessionsFile, &sessionData))
	sessions := make([]model.Session, 0, len(sessionData))

	for _, s := range sessionData {
		if len(scenarios) == 0 || slices.Contains(scenarios, s.Scenario) {
			sessions = append(sessions, s.Session)
		}
	}

	assert.NoError(t, client.SaveSessions(context.Background(), sessions))

	// load and store page views
	pageViewsFile, err := os.ReadFile("../../../test/page_views.csv")
	assert.NoError(t, err)
	var pageViewData []pageViewData
	assert.NoError(t, gocsv.UnmarshalBytes(pageViewsFile, &pageViewData))
	pageViews := make([]model.PageView, 0, len(pageViewData))

	for _, pv := range pageViewData {
		if len(scenarios) == 0 || slices.Contains(scenarios, pv.Scenario) {
			pageViews = append(pageViews, pv.PageView)

			if pv.Tags != "" {
				var tags map[string]string

				if err := json.Unmarshal([]byte(pv.Tags), &tags); err != nil {
					t.Fatal(err)
				}

				pageViews[len(pageViews)-1].Tags = tags
			}
		}
	}

	assert.NoError(t, client.SavePageViews(context.Background(), pageViews))

	// load and store events
	eventsFile, err := os.ReadFile("../../../test/events.csv")
	assert.NoError(t, err)
	var eventData []eventData
	assert.NoError(t, gocsv.UnmarshalBytes(eventsFile, &eventData))
	events := make([]model.Event, 0, len(eventData))

	for _, e := range eventData {
		if len(scenarios) == 0 || slices.Contains(scenarios, e.Scenario) {
			events = append(events, e.Event)

			if e.MetaData != "" {
				var metaData map[string]any

				if err := json.Unmarshal([]byte(e.MetaData), &metaData); err != nil {
					t.Fatal(err)
				}

				events[len(events)-1].MetaData = metaData
			}
		}
	}

	assert.NoError(t, client.SaveEvents(context.Background(), events))
}
