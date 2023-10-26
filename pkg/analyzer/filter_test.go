package analyzer

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFilter_Validate(t *testing.T) {
	filter := NewFilter(pkg.NullClient)
	filter.validate()
	assert.NotNil(t, filter)
	assert.NotNil(t, filter.Timezone)
	assert.Equal(t, time.UTC, filter.Timezone)
	assert.Zero(t, filter.From)
	assert.Zero(t, filter.To)
	filter = &Filter{From: util.PastDay(2), To: util.PastDay(5), Limit: 42}
	filter.validate()
	assert.Equal(t, util.PastDay(5), filter.From)
	assert.Equal(t, util.PastDay(2), filter.To)
	assert.Equal(t, 42, filter.Limit)
	filter = &Filter{From: util.PastDay(2), To: util.Today().Add(time.Hour * 24 * 5)}
	filter.validate()
	assert.Equal(t, util.PastDay(2), filter.From)
	assert.Equal(t, util.Today().Add(time.Hour*24), filter.To)
	filter = &Filter{Limit: -42, Path: []string{"/path"}, PathPattern: []string{"pattern"}}
	filter.validate()
	assert.Zero(t, filter.Limit)
	assert.Len(t, filter.Path, 1)
	assert.Equal(t, "/path", filter.Path[0])
	assert.Empty(t, filter.PathPattern)
	filter = &Filter{Limit: -42, PathPattern: []string{"pattern", "pattern"}}
	filter.validate()
	assert.Empty(t, filter.Path)
	assert.Len(t, filter.PathPattern, 1)
	assert.Equal(t, "pattern", filter.PathPattern[0])
}

func TestFilter_RemoveDuplicates(t *testing.T) {
	filter := NewFilter(pkg.NullClient)
	filter.Path = []string{
		"/",
		"/",
		"/foo",
		"/Foo",
		"/bar",
		"/foo",
	}
	filter.validate()
	assert.Len(t, filter.Path, 4)
	assert.Equal(t, "/", filter.Path[0])
	assert.Equal(t, "/foo", filter.Path[1])
	assert.Equal(t, "/Foo", filter.Path[2])
	assert.Equal(t, "/bar", filter.Path[3])
}

func TestFilter_BuildQuery(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 1, Time: util.Today().Add(time.Minute * 2), Path: "/foo"},
		{VisitorID: 1, Time: util.Today().Add(time.Minute*2 + time.Second*2), Path: "/foo"},
		{VisitorID: 1, Time: util.Today().Add(time.Minute*2 + time.Second*23), Path: "/bar"},

		{VisitorID: 2, Time: util.Today(), Path: "/bar"},
		{VisitorID: 2, Time: util.Today().Add(time.Second * 16), Path: "/foo"},
		{VisitorID: 2, Time: util.Today().Add(time.Second*16 + time.Second*8), Path: "/"},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: util.Today(), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Minute * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: -1, VisitorID: 2, Time: util.Today(), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Second * 16), Start: time.Now(), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today().Add(time.Minute * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Minute*2 + time.Second*23), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: -1, VisitorID: 2, Time: util.Today().Add(time.Second * 16), Start: time.Now(), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Second*16 + time.Second*8), Start: time.Now(), EntryPath: "/bar", ExitPath: "/", PageViews: 3},
		},
	})

	// no filter (from page views)
	analyzer := NewAnalyzer(dbClient)
	q, args := analyzer.getFilter(nil).buildQuery([]Field{FieldPath, FieldVisitors},
		[]Field{FieldPath}, []Field{FieldVisitors, FieldPath})
	var stats []model.PageStats
	rows, err := dbClient.Query(q, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 3)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 2, stats[1].Visitors)
	assert.Equal(t, 2, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join (from page views)
	q, args = analyzer.getFilter(&Filter{EntryPath: []string{"/"}}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	rows, err = dbClient.Query(q, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 3)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join and filter (from page views)
	q, args = analyzer.getFilter(&Filter{EntryPath: []string{"/"}, Path: []string{"/foo"}}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	rows, err = dbClient.Query(q, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 1, stats[0].Visitors)

	// filter (from page views)
	q, args = analyzer.getFilter(&Filter{Path: []string{"/foo"}}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	rows, err = dbClient.Query(q, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 2, stats[0].Visitors)

	// no filter (from sessions)
	q, args = analyzer.getFilter(nil).buildQuery([]Field{FieldVisitors, FieldSessions, FieldViews, FieldBounces, FieldBounceRate}, nil, nil)
	var vstats model.PageStats
	assert.NoError(t, dbClient.QueryRow(q, args...).Scan(&vstats.Visitors, &vstats.Sessions, &vstats.Views, &vstats.Bounces, &vstats.BounceRate))
	assert.Equal(t, 2, vstats.Visitors)
	assert.Equal(t, 2, vstats.Sessions)
	assert.Equal(t, 6, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)

	// filter (from page views)
	q, args = analyzer.getFilter(&Filter{Path: []string{"/foo"}, EntryPath: []string{"/"}}).buildQuery([]Field{FieldVisitors, FieldRelativeVisitors, FieldSessions, FieldViews, FieldRelativeViews, FieldBounces, FieldBounceRate}, nil, nil)
	assert.NoError(t, dbClient.QueryRow(q, args...).Scan(&vstats.Visitors, &vstats.RelativeVisitors, &vstats.Sessions, &vstats.Views, &vstats.RelativeViews, &vstats.Bounces, &vstats.BounceRate))
	assert.Equal(t, 1, vstats.Visitors)
	assert.Equal(t, 1, vstats.Sessions)
	assert.Equal(t, 2, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)
	assert.InDelta(t, 0.5, vstats.RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, vstats.RelativeViews, 0.01)

	// filter period
	q, args = analyzer.getFilter(&Filter{Period: pkg.PeriodWeek}).buildQuery([]Field{FieldDay, FieldVisitors}, []Field{FieldDay}, []Field{FieldDay})
	var visitors []model.VisitorStats
	rows, err = dbClient.Query(q, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.VisitorStats
		assert.NoError(t, rows.Scan(&stat.Day, &stat.Visitors))
		visitors = append(visitors, stat)
	}

	assert.Len(t, visitors, 1)
}
