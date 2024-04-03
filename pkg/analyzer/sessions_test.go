package analyzer

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSessions_List(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Minute * 10), Start: util.Today(), EntryPath: "/", ExitPath: "/pricing", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 30), Start: util.Today(), EntryPath: "/blog", ExitPath: "/", PageViews: 8, IsBounce: false},
			{Sign: 1, VisitorID: 3, SessionID: 3, Time: util.Today().Add(time.Minute * 45), Start: util.Today(), EntryPath: "/", ExitPath: "/about", PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 4, SessionID: 4, Time: util.Today().Add(time.Minute * 47), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, SessionID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Minute * 10), Path: "/pricing"},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 30), Path: "/blog"},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 31), Path: "/blog/1"},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 32), Path: "/blog/2"},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 33), Path: "/blog/3", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 34), Path: "/blog/4"},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 35), Path: "/blog/5"},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 36), Path: "/blog/6"},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Minute * 45), Path: "/"},
		{VisitorID: 3, SessionID: 3, Time: util.Today().Add(time.Minute * 42), Path: "/"},
		{VisitorID: 3, SessionID: 3, Time: util.Today().Add(time.Minute * 43), Path: "/"},
		{VisitorID: 3, SessionID: 3, Time: util.Today().Add(time.Minute * 45), Path: "/about"},
		{VisitorID: 4, SessionID: 4, Time: util.Today().Add(time.Minute * 47), Path: "/"},
	}))
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Minute * 5), Name: "event", MetaKeys: []string{"key"}, MetaValues: []string{"value"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.Sessions.List(nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 4)
	assert.Equal(t, uint64(1), stats[0].VisitorID)
	assert.Equal(t, uint64(2), stats[1].VisitorID)
	assert.Equal(t, uint64(3), stats[2].VisitorID)
	assert.Equal(t, uint64(4), stats[3].VisitorID)
	assert.Equal(t, uint32(1), stats[0].SessionID)
	assert.Equal(t, uint32(2), stats[1].SessionID)
	assert.Equal(t, uint32(3), stats[2].SessionID)
	assert.Equal(t, uint32(4), stats[3].SessionID)
	assert.Equal(t, uint16(2), stats[0].PageViews)
	assert.Equal(t, uint16(8), stats[1].PageViews)
	assert.Equal(t, uint16(3), stats[2].PageViews)
	assert.Equal(t, uint16(1), stats[3].PageViews)
	assert.Equal(t, "/", stats[0].EntryPath)
	assert.Equal(t, "/blog", stats[1].EntryPath)
	assert.Equal(t, "/", stats[2].EntryPath)
	assert.Equal(t, "/", stats[3].EntryPath)
	assert.Equal(t, "/pricing", stats[0].ExitPath)
	assert.Equal(t, "/", stats[1].ExitPath)
	assert.Equal(t, "/about", stats[2].ExitPath)
	assert.Equal(t, "/", stats[3].ExitPath)
	assert.False(t, stats[0].IsBounce)
	assert.False(t, stats[1].IsBounce)
	assert.False(t, stats[2].IsBounce)
	assert.True(t, stats[3].IsBounce)
	stats, err = analyzer.Sessions.List(&Filter{
		From:        util.Today().Add(time.Minute * 11),
		To:          util.Today().Add(time.Minute * 46),
		IncludeTime: true,
		EntryPath:   []string{"/blog"},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, uint64(2), stats[0].VisitorID)
	stats, err = analyzer.Sessions.List(&Filter{
		EventName: []string{"event"},
		EventMeta: map[string]string{
			"key": "value",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, uint64(1), stats[0].VisitorID)
	stats, err = analyzer.Sessions.List(&Filter{
		Tags: map[string]string{
			"author": "John",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, uint64(2), stats[0].VisitorID)
	stats, err = analyzer.Sessions.List(&Filter{
		Path: []string{"/about"},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, uint64(3), stats[0].VisitorID)
	_, err = analyzer.Sessions.List(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Sessions.List(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Sessions.List(&Filter{
		From:                 util.Today(),
		To:                   util.Today(),
		Limit:                200,
		IncludeTime:          true,
		MaxTimeOnPageSeconds: 3600,
		Sample:               10_000_000,
	})
	assert.NoError(t, err)
}

func TestSessions_Breakdown(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 10), Start: util.Today(), EntryPath: "/", ExitPath: "/pricing", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Second * 20), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 10), Path: "/pricing", DurationSeconds: 10},
		{VisitorID: 1, SessionID: 1, Time: util.Today(), Path: "/", DurationSeconds: 0},
		{VisitorID: 2, SessionID: 2, Time: util.Today(), Path: "/", DurationSeconds: 20},
	}))
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 5), Name: "event", MetaKeys: []string{"key"}, MetaValues: []string{"value"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	steps, err := analyzer.Sessions.Breakdown(nil)
	assert.NoError(t, err)
	assert.Nil(t, steps)
	steps, err = analyzer.Sessions.Breakdown(&Filter{
		VisitorID: 1,
		SessionID: 1,
		From:      util.Today(),
		To:        util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, steps, 3)
	assert.Equal(t, util.Today(), steps[0].PageView.Time)
	assert.Equal(t, util.Today().Add(time.Second*5), steps[1].Event.Time)
	assert.Equal(t, util.Today().Add(time.Second*10), steps[2].PageView.Time)
	assert.Equal(t, "/", steps[0].PageView.Path)
	assert.Equal(t, "/pricing", steps[2].PageView.Path)
	assert.Equal(t, "event", steps[1].Event.Name)
	assert.Equal(t, "key", steps[1].Event.MetaKeys[0])
	assert.Equal(t, "value", steps[1].Event.MetaValues[0])
	assert.Equal(t, uint32(10), steps[0].PageView.DurationSeconds)
	assert.Equal(t, uint32(0), steps[2].PageView.DurationSeconds)
	steps, err = analyzer.Sessions.Breakdown(&Filter{
		VisitorID:            1,
		SessionID:            1,
		From:                 util.Today(),
		To:                   util.Today(),
		Limit:                200,
		IncludeTime:          true,
		MaxTimeOnPageSeconds: 3600,
		Sample:               10_000_000,
	})
	assert.NoError(t, err)
	assert.Len(t, steps, 3)
}
