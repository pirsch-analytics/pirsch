package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBuildQuery(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: Today(), Path: "/"},
		{VisitorID: 1, Time: Today().Add(time.Minute * 2), Path: "/foo"},
		{VisitorID: 1, Time: Today().Add(time.Minute*2 + time.Second*2), Path: "/foo"},
		{VisitorID: 1, Time: Today().Add(time.Minute*2 + time.Second*23), Path: "/bar"},

		{VisitorID: 2, Time: Today(), Path: "/bar"},
		{VisitorID: 2, Time: Today().Add(time.Second * 16), Path: "/foo"},
		{VisitorID: 2, Time: Today().Add(time.Second*16 + time.Second*8), Path: "/"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: Today(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: Today(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 1, Time: Today().Add(time.Minute * 2), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: -1, VisitorID: 2, Time: Today(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: Today().Add(time.Second * 16), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today().Add(time.Minute * 2), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 1, Time: Today().Add(time.Minute*2 + time.Second*23), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: -1, VisitorID: 2, Time: Today().Add(time.Second * 16), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 2, Time: Today().Add(time.Second*16 + time.Second*8), EntryPath: "/bar", ExitPath: "/", PageViews: 3},
		},
	})

	// no filter (from page views)
	analyzer := NewAnalyzer(dbClient)
	args, query := buildQuery(analyzer.getFilter(nil), []field{fieldPath, fieldVisitors}, []field{fieldPath}, []field{fieldVisitors, fieldPath})
	var stats []PageStats
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 3)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 2, stats[1].Visitors)
	assert.Equal(t, 2, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join (from page views)
	args, query = buildQuery(analyzer.getFilter(&Filter{EntryPath: "/"}), []field{fieldPath, fieldVisitors}, []field{fieldPath}, []field{fieldPath})
	stats = stats[:0]
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 3)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join and filter (from page views)
	args, query = buildQuery(analyzer.getFilter(&Filter{EntryPath: "/", Path: "/foo"}), []field{fieldPath, fieldVisitors}, []field{fieldPath}, []field{fieldPath})
	stats = stats[:0]
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 1, stats[0].Visitors)

	// filter (from page views)
	args, query = buildQuery(analyzer.getFilter(&Filter{Path: "/foo"}), []field{fieldPath, fieldVisitors}, []field{fieldPath}, []field{fieldPath})
	stats = stats[:0]
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 2, stats[0].Visitors)

	// no filter (from sessions)
	args, query = buildQuery(analyzer.getFilter(nil), []field{fieldVisitors, fieldSessions, fieldViews, fieldBounces, fieldBounceRate}, nil, nil)
	var vstats PageStats
	assert.NoError(t, dbClient.Get(&vstats, query, args...))
	assert.Equal(t, 2, vstats.Visitors)
	assert.Equal(t, 2, vstats.Sessions)
	assert.Equal(t, 6, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)

	// filter (from page views)
	args, query = buildQuery(analyzer.getFilter(&Filter{Path: "/foo", EntryPath: "/"}), []field{fieldVisitors, fieldRelativeVisitors, fieldSessions, fieldViews, fieldRelativeViews, fieldBounces, fieldBounceRate}, nil, nil)
	assert.NoError(t, dbClient.Get(&vstats, query, args...))
	assert.Equal(t, 1, vstats.Visitors)
	assert.Equal(t, 1, vstats.Sessions)
	assert.Equal(t, 2, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)
	assert.InDelta(t, 0.5, vstats.RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, vstats.RelativeViews, 0.01)
}
