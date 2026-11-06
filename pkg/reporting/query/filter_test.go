package query

import (
	"testing"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/report"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
	"github.com/stretchr/testify/assert"
)

func TestQueryBuildQueryFilterGroups(t *testing.T) {
	q := NewQuery(client)

	// nested groups keep their precedence
	where, args := q.buildQueryFilter(pkg.TableSessions, request.Filter{
		Operator: request.OperatorGroupAnd,
		Filter: []request.Filter{
			{Dimension: dimensions.Country{}, Values: []any{"de"}},
			{
				Operator: request.OperatorGroupOr,
				Filter: []request.Filter{
					{Dimension: dimensions.City{}, Values: []any{"Berlin"}},
					{Dimension: dimensions.City{}, Values: []any{"Hamburg"}},
				},
			},
		},
	})
	assert.Equal(t, "(country_code = ?) AND ((city = ?) OR (city = ?))", where)
	assert.Equal(t, []any{"de", "Berlin", "Hamburg"}, args)

	// NOT negates the whole group
	where, args = q.buildQueryFilter(pkg.TableSessions, request.Filter{
		Operator: request.OperatorGroupNot,
		Filter: []request.Filter{
			{Dimension: dimensions.Country{}, Values: []any{"de"}},
			{Dimension: dimensions.City{}, Values: []any{"Berlin"}},
		},
	})
	assert.Equal(t, "NOT ((country_code = ?) AND (city = ?))", where)
	assert.Equal(t, []any{"de", "Berlin"}, args)

	// empty filters are skipped
	where, args = q.buildQueryFilter(pkg.TableSessions, request.Filter{
		Operator: request.OperatorGroupOr,
		Filter: []request.Filter{
			{Dimension: dimensions.Country{}},
			{Dimension: dimensions.City{}, Values: []any{"Berlin"}},
		},
	})
	assert.Equal(t, "(city = ?)", where)
	assert.Equal(t, []any{"Berlin"}, args)

	where, args = q.buildQueryFilter(pkg.TableSessions, request.Filter{
		Operator: request.OperatorGroupOr,
		Filter: []request.Filter{
			{Dimension: dimensions.Country{}},
		},
	})
	assert.Empty(t, where)
	assert.Empty(t, args)
}

func TestQueryBuildQueryWhereEmptyFilter(t *testing.T) {
	q := NewQuery(client)
	req := request.Request{
		SiteID: 1,
		Filter: []request.Filter{
			{Dimension: dimensions.Country{}},
			{
				Operator: request.OperatorGroupOr,
				Filter:   []request.Filter{{Dimension: dimensions.City{}}},
			},
		},
	}
	assert.Nil(t, q.prepare(&req))
	where, args := q.buildQueryWhere(req)
	assert.Equal(t, "WHERE site_id = ? ", where)
	assert.Equal(t, []any{uint64(1)}, args)
}

func TestQueryFilterContainsEscaping(t *testing.T) {
	q := NewQuery(client)
	assert.Equal(t, []any{`100\%`, `a\_b`, `back\\slash`}, q.buildQueryFilterILikeValues([]any{"100%", "a_b", `back\slash`}))

	loadTestData(t, []string{
		"simple",
	})
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	for _, testCase := range []struct {
		value    string
		expected int
	}{
		{"pricing", 1},
		{"pr_cing", 0}, // _ must not act as a wildcard
		{"%", 0},       // % must not act as a wildcard
	} {
		r := NewQuery(client).Run(request.Request{
			SiteID: 1,
			Period: request.Period{
				From: from,
				To:   from.AddDate(0, 0, 30),
			},
			Dimensions: []dimensions.Dimension{
				dimensions.Path{},
			},
			Filter: []request.Filter{
				{
					Operator:  request.OperatorContains,
					Dimension: dimensions.Path{},
					Values:    []any{testCase.value},
				},
			},
		})
		assert.Empty(t, r.Meta.Errors)
		assert.Len(t, r.Results, testCase.expected, testCase.value)
	}
}

func TestQueryBuildQueryWhereSiteAndPeriod(t *testing.T) {
	q := NewQuery(client)
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	day := time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC)

	for _, testCase := range []struct {
		period   request.Period
		imported bool
		query    string
		args     []any
	}{
		{
			period: request.Period{},
			query:  "WHERE site_id = ? ",
			args:   []any{uint64(1)},
		},
		{
			period: request.Period{From: day, To: day, Timezone: tz},
			query:  `WHERE site_id = ? AND toDate("time", 'Europe/Berlin') = toDate(?, 'Europe/Berlin') `,
			args:   []any{uint64(1), "2026-01-16"},
		},
		{
			period: request.Period{From: day, To: day.AddDate(0, 0, 1), Timezone: tz},
			query:  `WHERE site_id = ? AND toDate("time", 'Europe/Berlin') BETWEEN toDate(?, 'Europe/Berlin') AND toDate(?, 'Europe/Berlin') `,
			args:   []any{uint64(1), "2026-01-16", "2026-01-17"},
		},
		{
			period: request.Period{From: day, To: day.Add(time.Hour), Timezone: tz, IncludeTime: true},
			query:  `WHERE site_id = ? AND toDateTime("start", 'Europe/Berlin') BETWEEN toDateTime(?, 'Europe/Berlin') AND toDateTime(?, 'Europe/Berlin') `,
			args:   []any{uint64(1), "2026-01-16 00:00:00", "2026-01-16 01:00:00"},
		},
		{
			// open-ended
			period: request.Period{From: day, Timezone: tz},
			query:  `WHERE site_id = ? AND toDate("time", 'Europe/Berlin') >= toDate(?, 'Europe/Berlin') `,
			args:   []any{uint64(1), "2026-01-16"},
		},
		{
			period: request.Period{To: day, Timezone: tz},
			query:  `WHERE site_id = ? AND toDate("time", 'Europe/Berlin') <= toDate(?, 'Europe/Berlin') `,
			args:   []any{uint64(1), "2026-01-16"},
		},
		{
			// imported statistics are stored by date without timezone
			period:   request.Period{From: day, To: day.AddDate(0, 0, 1), Timezone: tz},
			imported: true,
			query:    `WHERE site_id = ? AND "date" BETWEEN toDate(?) AND toDate(?) `,
			args:     []any{uint64(1), "2026-01-16", "2026-01-17"},
		},
	} {
		query, args := q.buildQueryWhereSiteAndPeriod(1, testCase.period, pkg.TableSessions, testCase.imported)
		assert.Equal(t, testCase.query, query)
		assert.Equal(t, testCase.args, args)
	}
}

func TestQueryBuildQueryWithFill(t *testing.T) {
	q := NewQuery(client)
	day := time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC)
	req := request.Request{
		Period: request.Period{
			From:     day,
			To:       day.AddDate(0, 0, 1),
			Timezone: time.UTC,
		},
	}

	// date only, the last day is filled completely
	query, _ := q.buildQueryWithFill(req, dimensions.Hour{})
	assert.Equal(t, "WITH FILL FROM toStartOfDay(toDateTime(?, 'UTC')) TO toStartOfDay(toDateTime(?, 'UTC')) + INTERVAL 1 DAY STEP INTERVAL 1 HOUR", query)
	query, _ = q.buildQueryWithFill(req, dimensions.Minute{})
	assert.Equal(t, "WITH FILL FROM toStartOfDay(toDateTime(?, 'UTC')) TO toStartOfDay(toDateTime(?, 'UTC')) + INTERVAL 1 DAY STEP INTERVAL 1 MINUTE", query)

	// open-ended periods cannot be filled
	req.Period.From = time.Time{}
	query, args := q.buildQueryWithFill(req, dimensions.Day{})
	assert.Empty(t, query)
	assert.Empty(t, args)
}

func TestQuerySortResults(t *testing.T) {
	q := NewQuery(client)
	q.primaryTable = pkg.TableSessions
	day := time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC)
	results := []report.Result{
		{DimensionValues: []any{"/b", day}, MetricValues: []any{uint64(1), 0.5}},
		{DimensionValues: []any{"/a", day.AddDate(0, 0, 1)}, MetricValues: []any{uint64(3), 0.1}},
		{DimensionValues: []any{"/c", day.AddDate(0, 0, -1)}, MetricValues: []any{uint64(2), 0.9}},
	}
	requestDimensions := []dimensions.Dimension{dimensions.Path{}, dimensions.Day{}}
	requestMetrics := []metrics.Metric{metrics.Visitors{}, metrics.BounceRate{}}

	// sort by dimension
	q.sortResults(results, requestDimensions, requestMetrics, []request.OrderBy{{Dimension: dimensions.Path{}, Direction: request.DirectionASC}})
	assert.Equal(t, "/a", results[0].DimensionValues[0])
	assert.Equal(t, "/b", results[1].DimensionValues[0])
	assert.Equal(t, "/c", results[2].DimensionValues[0])

	// sort by time dimension
	q.sortResults(results, requestDimensions, requestMetrics, []request.OrderBy{{Dimension: dimensions.Day{}, Direction: request.DirectionDESC}})
	assert.Equal(t, "/a", results[0].DimensionValues[0])
	assert.Equal(t, "/b", results[1].DimensionValues[0])
	assert.Equal(t, "/c", results[2].DimensionValues[0])

	// sort by metric
	q.sortResults(results, requestDimensions, requestMetrics, []request.OrderBy{{Metric: metrics.BounceRate{}, Direction: request.DirectionDESC}})
	assert.Equal(t, "/c", results[0].DimensionValues[0])
	assert.Equal(t, "/b", results[1].DimensionValues[0])
	assert.Equal(t, "/a", results[2].DimensionValues[0])
}

func TestQueryCompareAny(t *testing.T) {
	q := NewQuery(client)
	day := time.Date(2026, time.January, 16, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, -1, q.compareAny(uint64(1), uint64(2)))
	assert.Equal(t, 1, q.compareAny(2.5, 1.5))
	assert.Equal(t, 0, q.compareAny("a", "a"))
	assert.Equal(t, -1, q.compareAny(uint8(1), uint8(2)))
	assert.Equal(t, 1, q.compareAny(int64(2), int64(-1)))
	assert.Equal(t, -1, q.compareAny(day, day.Add(time.Hour)))
	assert.Equal(t, -1, q.compareAny(nil, uint64(1)))
	assert.Equal(t, 1, q.compareAny(uint64(1), nil))
	assert.Equal(t, 0, q.compareAny(nil, nil))

	// different types must not panic
	assert.Equal(t, 0, q.compareAny(uint64(1), "1"))
	assert.Equal(t, 0, q.compareAny(struct{}{}, struct{}{}))
}

func TestQueryMergeResultsJoinTableEqualsPrimary(t *testing.T) {
	q := NewQuery(client)
	q.primaryTable = pkg.TableEvents
	q.joinTable = pkg.TableSessions
	requestMetrics := []metrics.Metric{metrics.Events{}, metrics.Bounces{}, metrics.Visitors{}}
	primaryMetrics, secondaryMetrics := q.splitMetrics(requestMetrics)
	assert.Equal(t, []metrics.Metric{metrics.Events{}, metrics.Visitors{}}, primaryMetrics)
	assert.Equal(t, []metrics.Metric{metrics.Bounces{}}, secondaryMetrics)
	primary := []report.Result{{DimensionValues: []any{"/"}, MetricValues: []any{uint64(5), uint64(2)}}}
	secondary := []report.Result{{DimensionValues: []any{"/"}, MetricValues: []any{uint64(1)}}}
	q.mergeResults(primary, secondary, requestMetrics, primaryMetrics)
	assert.Equal(t, []any{uint64(5), uint64(1), uint64(2)}, primary[0].MetricValues)
}

func TestQueryFunnelStepError(t *testing.T) {
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := request.FunnelRequest{
		SiteID: 1,
		Period: request.Period{
			From: from,
			To:   from.AddDate(0, 0, 30),
		},
		Filter: [][]request.Filter{
			{{Dimension: dimensions.Path{}, Values: []any{"/"}}},
			{
				{
					// mixing tables within an OR group is not supported
					Operator: request.OperatorGroupOr,
					Filter: []request.Filter{
						{Dimension: dimensions.Event{}, Values: []any{"Contact Button"}},
						{Dimension: dimensions.EntryPath{}, Values: []any{"/"}},
					},
				},
			},
		},
	}
	query, args, errs := NewQuery(client).DryRunFunnel(req)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "funnel step 2")
	assert.Empty(t, query)
	assert.Nil(t, args)
	r := NewQuery(client).Funnel(req)
	assert.Len(t, r.Meta.Errors, 1)
}
