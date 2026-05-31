package query

import (
	"context"
	"encoding/json"
	"os"
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
	loadTestData(t)
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
	assert.Len(t, r.Results, 2)
	assert.Equal(t, from, r.Results[0].DimensionValues[0])
	assert.Equal(t, uint64(2), r.Results[0].MetricValues[0])
	assert.Equal(t, 0.5, r.Results[0].MetricValues[1])
	assert.Equal(t, from.Add(time.Hour*24), r.Results[1].DimensionValues[0])
	assert.Equal(t, uint64(1), r.Results[1].MetricValues[0])
	assert.Equal(t, float64(0), r.Results[1].MetricValues[1])
}

func TestQueryFromPageViews(t *testing.T) {
	loadTestData(t)
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
	loadTestData(t)
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
	loadTestData(t)
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
	assert.Equal(t, "https://duckduckgo.com", args[8])
	// TODO
}

func TestQueryFromPageViewsFiltered(t *testing.T) {
	loadTestData(t)
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

func TestQueryFromEventsFiltered(t *testing.T) {
	loadTestData(t)
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

func newQuery() (*Query, time.Time, time.Time) {
	q := NewQuery(client)
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)
	return q, from, to
}

type pageViewData struct {
	model.PageView
	Tags string `csv:"tags"`
}

type eventData struct {
	model.Event
	MetaData string `csv:"meta_data"`
}

func loadTestData(t *testing.T) {
	db.CleanupDB(t, client)

	// load and store sessions
	sessionsFile, err := os.ReadFile("../../../test/sessions.csv")
	assert.NoError(t, err)
	var sessions []model.Session
	assert.NoError(t, gocsv.UnmarshalBytes(sessionsFile, &sessions))
	assert.NoError(t, client.SaveSessions(context.Background(), sessions))

	// load and store page views
	pageViewsFile, err := os.ReadFile("../../../test/page_views.csv")
	assert.NoError(t, err)
	var pageViewData []pageViewData
	assert.NoError(t, gocsv.UnmarshalBytes(pageViewsFile, &pageViewData))
	pageViews := make([]model.PageView, len(pageViewData))

	for i, pv := range pageViewData {
		pageViews[i] = pv.PageView

		if pv.Tags != "" {
			var tags map[string]string

			if err := json.Unmarshal([]byte(pv.Tags), &tags); err != nil {
				t.Fatal(err)
			}

			pageViews[i].Tags = tags
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
		events[i] = e.Event

		if e.MetaData != "" {
			var metaData map[string]any

			if err := json.Unmarshal([]byte(e.MetaData), &metaData); err != nil {
				t.Fatal(err)
			}

			events[i].MetaData = metaData
		}
	}

	assert.NoError(t, client.SaveEvents(context.Background(), events))
}
