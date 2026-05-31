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
			metrics.BounceRate{},
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
	assert.Len(t, r.Results, 2)
	assert.Equal(t, from, r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, 0.5, r.Results[0].MetricValues[1])
	assert.Equal(t, from.Add(time.Hour*24), r.Results[1].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[1].MetricValues[0])
	assert.InDelta(t, 0.6666, r.Results[1].MetricValues[1], 0.001)
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
			metrics.PageViews{},
		},
		OrderBy: []request.OrderBy{
			{Metric: metrics.PageViews{}},
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

	// result
	assert.Len(t, r.Results, 3)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)
	assert.Len(t, r.Results[1].DimensionValues, 1)
	assert.Len(t, r.Results[1].MetricValues, 1)
	assert.Len(t, r.Results[2].DimensionValues, 1)
	assert.Len(t, r.Results[2].MetricValues, 1)
	assert.Equal(t, "/", r.Results[0].DimensionValues[0])
	assert.Equal(t, "/pricing", r.Results[1].DimensionValues[0])
	assert.Equal(t, "/landing", r.Results[2].DimensionValues[0])
	assert.Equal(t, uint64(5), r.Results[0].MetricValues[0])
	assert.Equal(t, uint64(2), r.Results[1].MetricValues[0])
	assert.Equal(t, uint64(1), r.Results[2].MetricValues[0])
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

	// result
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)
	assert.Equal(t, "Contact Button", r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
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
		},
	}

	// tables and filter
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
	assert.Equal(t, []any{"/pricing", "/landing"}, q.subqueryFilter[0].filter.Filter[0].Values)
	assert.Equal(t, []any{"https://duckduckgo.com"}, q.subqueryFilter[0].filter.Filter[1].Values)

	// query
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
	assert.Equal(t, []any{"/pricing", "/landing"}, args[7])
	assert.Equal(t, "https://duckduckgo.com", args[8])

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
	assert.Len(t, args, 8)
	assert.Equal(t, uint64(1), args[0])
	assert.Equal(t, from, args[1])
	assert.Equal(t, to, args[2])
	assert.Equal(t, "/", args[3])
	assert.Equal(t, uint64(1), args[4])
	assert.Equal(t, from, args[5])
	assert.Equal(t, to, args[6])
	assert.Equal(t, "Contact Button", args[7])

	// result
	assert.Len(t, r.Results, 1)
	assert.Len(t, r.Results[0].DimensionValues, 1)
	assert.Len(t, r.Results[0].MetricValues, 1)
	assert.Equal(t, "/", r.Results[0].DimensionValues[0])
	assert.Equal(t, int64(1), r.Results[0].MetricValues[0])
}

func TestQueryTimeOnPage(t *testing.T) {
	// TODO
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
	assert.Equal(t, 0.5, r.Results[0].MetricValues[1])
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
	assert.InDelta(t, 0.6666, r.Results[0].MetricValues[1], 0.001)
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
	sessions := make([]model.Session, len(sessionData))

	for i, s := range sessionData {
		if len(scenarios) == 0 || slices.Contains(scenarios, s.Scenario) {
			sessions[i] = s.Session
		}
	}

	assert.NoError(t, client.SaveSessions(context.Background(), sessions))

	// load and store page views
	pageViewsFile, err := os.ReadFile("../../../test/page_views.csv")
	assert.NoError(t, err)
	var pageViewData []pageViewData
	assert.NoError(t, gocsv.UnmarshalBytes(pageViewsFile, &pageViewData))
	pageViews := make([]model.PageView, len(pageViewData))

	for i, pv := range pageViewData {
		if len(scenarios) == 0 || slices.Contains(scenarios, pv.Scenario) {
			pageViews[i] = pv.PageView

			if pv.Tags != "" {
				var tags map[string]string

				if err := json.Unmarshal([]byte(pv.Tags), &tags); err != nil {
					t.Fatal(err)
				}

				pageViews[i].Tags = tags
			}
		}
	}

	assert.NoError(t, client.SavePageViews(context.Background(), pageViews))

	// load and store events
	eventsFile, err := os.ReadFile("../../../test/events.csv")
	assert.NoError(t, err)
	var eventData []eventData
	assert.NoError(t, gocsv.UnmarshalBytes(eventsFile, &eventData))
	events := make([]model.Event, len(eventData))

	for i, e := range eventData {
		if len(scenarios) == 0 || slices.Contains(scenarios, e.Scenario) {
			events[i] = e.Event

			if e.MetaData != "" {
				var metaData map[string]any

				if err := json.Unmarshal([]byte(e.MetaData), &metaData); err != nil {
					t.Fatal(err)
				}

				events[i].MetaData = metaData
			}
		}
	}

	assert.NoError(t, client.SaveEvents(context.Background(), events))
}
