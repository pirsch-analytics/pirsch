package analyzer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzer_ActiveVisitors(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: time.Now().Add(-time.Minute * 30), Path: "/", Title: "Home"},
		{VisitorID: 1, Time: time.Now().Add(-time.Minute * 20), Path: "/", Title: "Home"},
		{VisitorID: 1, Time: time.Now().Add(-time.Minute * 15), Path: "/bar", Title: "Bar"},
		{VisitorID: 2, Time: time.Now().Add(-time.Minute * 4), Path: "/bar", Title: "Bar"},
		{VisitorID: 2, Time: time.Now().Add(-time.Minute * 3), Path: "/foo", Title: "Foo"},
		{VisitorID: 3, Time: time.Now().Add(-time.Minute * 3), Path: "/", Title: "Home"},
		{VisitorID: 4, Time: time.Now().Add(-time.Minute), Path: "/", Title: "Home"},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(-time.Minute * 25), Start: time.Now()},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(-time.Minute * 25), Start: time.Now()},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(-time.Minute * 15), Start: time.Now()},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(-time.Minute * 3), Start: time.Now()},
			{Sign: 1, VisitorID: 3, Time: time.Now().Add(-time.Minute * 5), Start: time.Now()},
		},
		{
			{Sign: -1, VisitorID: 3, Time: time.Now().Add(-time.Minute * 5), Start: time.Now()},
			{Sign: 1, VisitorID: 3, Time: time.Now().Add(-time.Minute * 3), Start: time.Now()},
			{Sign: 1, VisitorID: 4, Time: time.Now().Add(-time.Minute), Start: time.Now()},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, count, err := analyzer.Visitors.Active(nil, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path)
	assert.Equal(t, "/bar", visitors[1].Path)
	assert.Equal(t, "/foo", visitors[2].Path)
	assert.Empty(t, visitors[0].Title)
	assert.Empty(t, visitors[1].Title)
	assert.Empty(t, visitors[2].Title)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	visitors, count, err = analyzer.Visitors.Active(&Filter{Path: []string{"/bar"}}, time.Minute*30)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "/bar", visitors[0].Path)
	assert.Equal(t, 2, visitors[0].Visitors)
	_, _, err = analyzer.Visitors.Active(getMaxFilter(""), time.Minute*10)
	assert.NoError(t, err)
	visitors, count, err = analyzer.Visitors.Active(&Filter{IncludeTitle: true}, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Home", visitors[0].Title)
	assert.Equal(t, "Bar", visitors[1].Title)
	assert.Equal(t, "Foo", visitors[2].Title)
	_, _, err = analyzer.Visitors.Active(getMaxFilter(""), time.Minute*10)
	assert.NoError(t, err)
	_, _, err = analyzer.Visitors.Active(getMaxFilter("event"), time.Minute*10)
	assert.NoError(t, err)
}

func TestAnalyzer_TotalVisitors(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 3, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: false, DurationSeconds: 600},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 4, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, Time: util.PastDay(4), SessionID: 4, Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 3, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 4, Time: util.PastDay(4), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 7, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 8, Time: util.PastDay(2), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}, Path: "/"},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Total(&Filter{From: util.PastDay(4), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 9, visitors.Visitors)
	assert.Equal(t, 11, visitors.Sessions)
	assert.Equal(t, 13, visitors.Views)
	assert.Equal(t, 8, visitors.Bounces)
	assert.InDelta(t, 0.7272, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{From: util.PastDay(2), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 5, visitors.Visitors)
	assert.Equal(t, 5, visitors.Sessions)
	assert.Equal(t, 6, visitors.Views)
	assert.Equal(t, 3, visitors.Bounces)
	assert.InDelta(t, 0.6, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{From: time.Now().UTC().Add(-time.Minute * 15), To: util.Today(), IncludeTime: true, IncludeCR: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	assert.InDelta(t, 1, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		EventName: []string{"event"},
		EventMeta: map[string]string{"foo": "val0"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 2, visitors.Views)
	assert.Equal(t, 0, visitors.Bounces)
	assert.InDelta(t, 0, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.11, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		Tags:      map[string]string{"author": "John"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, visitors.Visitors)
	assert.Equal(t, 6, visitors.Sessions)
	assert.Equal(t, 7, visitors.Views)
	assert.Equal(t, 4, visitors.Bounces)
	assert.InDelta(t, 0.6666, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.5555, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		Tags:      map[string]string{"author": "!John"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, visitors.Visitors)
	assert.Equal(t, 6, visitors.Sessions)
	assert.Equal(t, 6, visitors.Views)
	assert.Equal(t, 4, visitors.Bounces)
	assert.InDelta(t, 0.6666, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.5555, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		Tags:      map[string]string{"author": "Alice"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, visitors.Visitors)
	assert.Equal(t, 2, visitors.Sessions)
	assert.Equal(t, 2, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 0.5, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.2222, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:        util.PastDay(4),
		To:          util.Today(),
		EventName:   []string{"event"},
		PathPattern: []string{`\/.*`},
		IncludeCR:   true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 2, visitors.Views)
	assert.Equal(t, 0, visitors.Bounces)
	assert.InDelta(t, 0, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.1111, visitors.CR, 0.01)

	// ignore metadata when event name is not set
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		EventMeta: map[string]string{"foo": "val0"},
	})
	assert.NoError(t, err)

	// imported statistics
	past5Days := util.PastDay(5).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past5Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:          util.PastDay(5),
		To:            util.Today(),
		ImportedUntil: util.PastDay(4),
	})
	assert.NoError(t, err)
	assert.Equal(t, 11, visitors.Visitors)
	assert.Equal(t, 14, visitors.Sessions)
	assert.Equal(t, 17, visitors.Views)
	assert.Equal(t, 9, visitors.Bounces)
	assert.InDelta(t, 0.6428, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:          util.PastDay(5),
		To:            util.PastDay(5),
		ImportedUntil: util.PastDay(4),
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, visitors.Visitors)
	assert.Equal(t, 3, visitors.Sessions)
	assert.Equal(t, 4, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 0.3333, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:          util.PastDay(5),
		To:            util.PastDay(5),
		ImportedUntil: util.PastDay(4),
		Platform:      pkg.PlatformMobile,
	})
	assert.NoError(t, err)
}

func TestAnalyzer_TotalUniqueVisitors(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 3, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: false, DurationSeconds: 600},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 4, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4), SessionID: 4, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 3, Path: "/bar"},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 4, Time: util.PastDay(4), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 7, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 8, Time: util.PastDay(2), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.TotalVisitors(&Filter{From: util.PastDay(4), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 9, visitors)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{From: util.PastDay(2), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 5, visitors)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{
		From:        time.Now().UTC().Add(-time.Minute * 15),
		To:          util.Today(),
		IncludeTime: true,
		IncludeCR:   true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		EventName: []string{"event"},
		EventMeta: map[string]string{"foo": "val0"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 9, visitors)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{
		From: util.PastDay(4),
		To:   util.Today(),
		Sort: []Sort{
			{Field: FieldRelativeVisitors, Direction: pkg.DirectionASC},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 9, visitors)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{
		From: util.PastDay(4),
		To:   util.Today(),
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 9, visitors)

	// imported statistics
	past5Days := util.PastDay(5).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past5Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.TotalVisitors(&Filter{
		From:          util.PastDay(5),
		To:            util.Today(),
		ImportedUntil: util.PastDay(4),
	})
	assert.NoError(t, err)
	assert.Equal(t, 11, visitors)
}

func TestAnalyzer_TotalPageViews(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 3, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: false, DurationSeconds: 600},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 4, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4), SessionID: 4, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 3, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 4, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 7, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 8, Time: util.PastDay(2), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	pageViews, err := analyzer.Visitors.TotalPageViews(&Filter{From: util.PastDay(4), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 13, pageViews)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{From: util.PastDay(2), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 6, pageViews)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, pageViews)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, pageViews)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{
		From:        time.Now().UTC().Add(-time.Minute * 15),
		To:          util.Today(),
		IncludeTime: true,
		IncludeCR:   true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, pageViews)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		EventName: []string{"event"},
		EventMeta: map[string]string{"foo": "val0"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 13, pageViews)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{
		From: util.PastDay(4),
		To:   util.Today(),
		Sort: []Sort{
			{Field: FieldRelativeVisitors, Direction: pkg.DirectionASC},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 13, pageViews)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{
		From: util.PastDay(4),
		To:   util.Today(),
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 13, pageViews)

	// imported statistics
	past5Days := util.PastDay(5).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past5Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	pageViews, err = analyzer.Visitors.TotalPageViews(&Filter{
		From:          util.PastDay(5),
		To:            util.Today(),
		ImportedUntil: util.PastDay(4),
	})
	assert.NoError(t, err)
	assert.Equal(t, 17, pageViews)
}

func TestAnalyzer_TotalSessions(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4), Start: time.Now(), SessionID: 4, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 3, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: false, DurationSeconds: 600},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 4, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4), SessionID: 4, Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 3, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 4, Time: util.PastDay(4), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 7, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 8, Time: util.PastDay(2), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	sessions, err := analyzer.Visitors.TotalSessions(&Filter{From: util.PastDay(4), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 11, sessions)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{From: util.PastDay(2), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 5, sessions)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, sessions)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, sessions)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{
		From:        time.Now().UTC().Add(-time.Minute * 15),
		To:          util.Today(),
		IncludeTime: true,
		IncludeCR:   true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, sessions)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		EventName: []string{"event"},
		EventMeta: map[string]string{"foo": "val0"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 11, sessions)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{
		From: util.PastDay(4),
		To:   util.Today(),
		Sort: []Sort{
			{Field: FieldRelativeVisitors, Direction: pkg.DirectionASC},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 11, sessions)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{
		From: util.PastDay(4),
		To:   util.Today(),
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 11, sessions)

	// imported statistics
	past5Days := util.PastDay(5).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past5Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	sessions, err = analyzer.Visitors.TotalSessions(&Filter{
		From:          util.PastDay(5),
		To:            util.Today(),
		ImportedUntil: util.PastDay(4),
	})
	assert.NoError(t, err)
	assert.Equal(t, 14, sessions)
}

func TestAnalyzer_TotalVisitorsEvent(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 2, Time: util.Today(), Path: "/foo"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: util.Today(), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"1", "2"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Total(&Filter{From: util.Today(), To: util.Today(), EventName: []string{"!event"}})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{From: util.Today(), To: util.Today(), EventName: []string{"event"}, EventMetaKey: []string{"bar"}, IncludeCR: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors.CR, 0.01)
	visitors, err = analyzer.Visitors.Total(&Filter{From: util.Today(), To: util.Today(), EventName: []string{"event"}, EventMeta: map[string]string{"bar": "2"}, IncludeCR: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors.CR, 0.01)
}

func TestAnalyzer_TotalVisitorsCustomMetric(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 5, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 2, Time: util.Today(), Path: "/foo"},
		{VisitorID: 3, Time: util.Today(), Path: "/bar"},
		{VisitorID: 4, Time: util.Today(), Path: "/"},
		{VisitorID: 5, Time: util.Today(), Path: "/"},
		{VisitorID: 6, Time: util.Today(), Path: "/foo"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: util.Today(), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.89", "EUR"}, Path: "/"},
		{VisitorID: 3, Time: util.Today(), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"3.12", "EUR"}, Path: "/bar"},
		{VisitorID: 4, Time: util.Today(), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.77", "USD"}, Path: "/"},
		{VisitorID: 6, Time: util.Today(), Name: "Sale", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "2.98"}, Path: "/foo"},
		{VisitorID: 6, Time: util.Today(), Name: "Unrelated", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "99"}, Path: "/foo"},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Total(&Filter{
		From:             util.Today(),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 4, visitors.Visitors)
	assert.Equal(t, 4, visitors.Sessions)
	assert.Equal(t, 4, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 0.25, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0.6666, visitors.CR, 0.01)
	assert.InDelta(t, 2.44, visitors.CustomMetricAvg, 0.001)
	assert.InDelta(t, 9.76, visitors.CustomMetricTotal, 0.001)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:             util.Today(),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, visitors.Visitors)
	assert.Equal(t, 3, visitors.Sessions)
	assert.Equal(t, 3, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 0.3333, visitors.BounceRate, 0.01)
	assert.InDelta(t, 0, visitors.CR, 0.01)
	assert.InDelta(t, 2.6633, visitors.CustomMetricAvg, 0.001)
	assert.InDelta(t, 7.99, visitors.CustomMetricTotal, 0.001)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:             util.Today(),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		Path:             []string{"/"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	assert.InDelta(t, 1.89, visitors.CustomMetricAvg, 0.001)
	assert.InDelta(t, 1.89, visitors.CustomMetricTotal, 0.001)
	visitors, err = analyzer.Visitors.Total(&Filter{
		From:             util.Today(),
		To:               util.Today(),
		EventName:        []string{"Unrelated"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeInteger,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 0, visitors.Bounces)
	assert.InDelta(t, 0, visitors.BounceRate, 0.01)
	assert.InDelta(t, 99, visitors.CustomMetricAvg, 0.001)
	assert.InDelta(t, 99, visitors.CustomMetricTotal, 0.001)
	filter := getMaxFilter("Sale")
	filter.CustomMetricType = pkg.CustomMetricTypeFloat
	filter.CustomMetricKey = "amount"
	filter.To = util.Today()
	visitors, err = analyzer.Visitors.Total(filter)
}

func TestAnalyzer_TotalVisitorsPageViews(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(9).Add(time.Minute * 15), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(9), Start: time.Now(), EntryPath: "/foo", ExitPath: "/", PageViews: 5},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4).Add(time.Minute * 15), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(4), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 15), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 3},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 2},
			{Sign: 1, VisitorID: 8, Time: util.PastDay(3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 9, Time: util.PastDay(3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 10, Time: util.PastDay(2).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 2},
			{Sign: 1, VisitorID: 11, Time: util.PastDay(2), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 12, Time: util.PastDay(2), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 13, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(9).Add(time.Minute * 10), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(9).Add(time.Minute * 12), Path: "/foo"},
		{VisitorID: 1, Time: util.PastDay(9).Add(time.Minute * 15), Path: "/bar"},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/foo"},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/foo", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/bar"},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 3, Time: util.PastDay(4).Add(time.Minute * 10), Path: "/"},
		{VisitorID: 3, Time: util.PastDay(4).Add(time.Minute * 12), Path: "/foo"},
		{VisitorID: 3, Time: util.PastDay(4).Add(time.Minute * 15), Path: "/bar"},
		{VisitorID: 4, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 5, Time: util.PastDay(4), Path: "/bar"},
		{VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 10), Path: "/"},
		{VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 12), Path: "/foo", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 15), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 7, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 7, Time: util.PastDay(3).Add(time.Minute * 10), Path: "/bar"},
		{VisitorID: 8, Time: util.PastDay(3), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 9, Time: util.PastDay(3), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 10, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 10, Time: util.PastDay(2).Add(time.Minute * 5), Path: "/bar"},
		{VisitorID: 11, Time: util.PastDay(2), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 12, Time: util.PastDay(2), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 13, Time: util.Today(), Path: "/"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: util.PastDay(9), Path: "/foo", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"a", "b"}},
		{VisitorID: 3, Time: util.PastDay(9), Path: "/", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"a", "b"}},
		{VisitorID: 4, Time: util.PastDay(9), Path: "/foo", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"a", "d"}},
		{VisitorID: 6, Time: util.PastDay(9), Path: "/foo", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"a", "b"}},
		{VisitorID: 8, Time: util.PastDay(9), Path: "/foo", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"c", "b"}},
		{VisitorID: 8, Time: util.PastDay(9), Path: "/", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"a", "b"}},
		{VisitorID: 10, Time: util.PastDay(9), Path: "/foo", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"a", "b"}},
		{VisitorID: 11, Time: util.PastDay(9), Path: "/foo", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"a", "d"}},
		{VisitorID: 13, Time: util.PastDay(9), Path: "/", Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"c", "d"}},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.TotalVisitorsPageViews(nil)
	assert.ErrorIs(t, err, ErrNoPeriodOrDay)
	assert.Nil(t, visitors)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From: util.PastDay(2),
		To:   util.PastDay(2),
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 3, visitors.Visitors)
	assert.Equal(t, 4, visitors.Views)
	assert.InDelta(t, 0.5, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, -0.5, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From: util.PastDay(3),
		To:   util.PastDay(2),
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 7, visitors.Visitors)
	assert.Equal(t, 11, visitors.Views)
	assert.InDelta(t, 1.3333, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1.2, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From:   util.PastDay(9),
		To:     util.PastDay(2),
		Sample: 10_000_000,
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 12, visitors.Visitors)
	assert.Equal(t, 24, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From:   util.PastDay(9),
		To:     util.PastDay(2),
		Path:   []string{"/foo"},
		Sample: 10_000_000,
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 4, visitors.Visitors)
	assert.Equal(t, 5, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From:      util.PastDay(9),
		To:        util.PastDay(2),
		EventName: []string{"event"},
		Sample:    10_000_000,
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 7, visitors.Visitors)
	assert.Equal(t, 14, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From:      util.PastDay(9),
		To:        util.PastDay(2),
		Path:      []string{"/foo"},
		EventName: []string{"event"},
		EventMeta: map[string]string{
			"foo": "a",
		},
		Sample: 10_000_000,
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 3, visitors.Visitors)
	assert.Equal(t, 3, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From:      util.PastDay(9),
		To:        util.PastDay(2),
		EntryPath: []string{"/"},
		ExitPath:  []string{"/bar"},
		Sample:    10_000_000,
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 4, visitors.Visitors)
	assert.Equal(t, 10, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From: util.PastDay(9),
		To:   util.PastDay(2),
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 6, visitors.Visitors)
	assert.Equal(t, 7, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From: util.PastDay(9),
		To:   util.PastDay(2),
		Tags: map[string]string{"author": "!John"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 9, visitors.Visitors)
	assert.Equal(t, 17, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From: util.PastDay(9),
		To:   util.PastDay(2),
		Tags: map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, visitors)
	assert.Equal(t, 2, visitors.Visitors)
	assert.Equal(t, 2, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	_, err = analyzer.Visitors.TotalVisitorsPageViews(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.TotalVisitorsPageViews(getMaxFilter("event"))
	assert.NoError(t, err)

	// imported statistics
	past10Days := util.PastDay(10).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past10Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From:          util.PastDay(10),
		To:            util.Today(),
		ImportedUntil: util.PastDay(9),
	})
	assert.NoError(t, err)
	assert.Equal(t, 15, visitors.Visitors)
	assert.Equal(t, 29, visitors.Views)
	assert.InDelta(t, 1, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, visitors.ViewsGrowth, 0.001)
	visitors, err = analyzer.Visitors.TotalVisitorsPageViews(&Filter{
		From:          util.PastDay(5),
		To:            util.Today(),
		ImportedUntil: util.PastDay(9),
	})
	assert.NoError(t, err)
	assert.Equal(t, 13, visitors.Visitors)
	assert.Equal(t, 21, visitors.Views)
	assert.InDelta(t, 2.25, visitors.VisitorsGrowth, 0.001)
	assert.InDelta(t, 0.75, visitors.ViewsGrowth, 0.001)
}

func TestAnalyzer_ByPeriodAndAvgSessionDuration(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, EntryPath: "/bar", ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 4, EntryPath: "/bar", ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 15), Start: time.Now(), SessionID: 4, EntryPath: "/bar", ExitPath: "/", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4), Start: time.Now(), SessionID: 4, EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), Start: time.Now(), SessionID: 3, EntryPath: "/bar", ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, EntryPath: "/bar", ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, EntryPath: "/bar", ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 15), Start: time.Now(), SessionID: 2, EntryPath: "/bar", ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 600},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(2), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: util.PastDay(2), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 4, Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 15), SessionID: 4, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4), SessionID: 4, Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4).Add(time.Minute * 10), SessionID: 3, Path: "/bar"},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 4, Time: util.PastDay(4), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 5, Time: util.PastDay(2).Add(time.Minute * 15), SessionID: 2, Path: "/bar"},
		{VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 8), SessionID: 2, Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 6, Time: util.PastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 7, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 8, Time: util.PastDay(2), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 9, Time: util.Today(), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByPeriod(&Filter{
		From:      util.PastDay(4),
		To:        util.Today(),
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, util.PastDay(4), visitors[0].Day.Time)
	assert.Equal(t, util.PastDay(3), visitors[1].Day.Time)
	assert.Equal(t, util.PastDay(2), visitors[2].Day.Time)
	assert.Equal(t, util.PastDay(1), visitors[3].Day.Time)
	assert.Equal(t, util.Today(), visitors[4].Day.Time)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 4, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 6, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 4, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 7, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 6, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 5, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 2, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 0.8333, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[0].CR, 0.01)
	assert.InDelta(t, 0, visitors[1].CR, 0.01)
	assert.InDelta(t, 1, visitors[2].CR, 0.01)
	assert.InDelta(t, 0, visitors[3].CR, 0.01)
	assert.InDelta(t, 1, visitors[4].CR, 0.01)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		Path: []string{"/"},
		From: util.PastDay(4),
		To:   util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 3, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 5, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 3, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 5, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 3, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 4, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 2, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 0.8, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.6666, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		EntryPath: []string{"/"},
		From:      util.PastDay(4),
		To:        util.Today(),
		Sample:    10_000_000,
		Limit:     10,
		Period:    pkg.PeriodDay,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 3, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 4, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 3, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 4, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 4, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 4, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 2, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.6666, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:   util.PastDay(4),
		To:     util.Today(),
		Period: pkg.PeriodDay,
		Tags:   map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 0, visitors[4].Visitors)
	assert.Equal(t, 2, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 2, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 0, visitors[4].Sessions)
	assert.Equal(t, 3, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 2, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 0, visitors[4].Views)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 0, visitors[4].Bounces)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[4].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:   util.PastDay(4),
		To:     util.Today(),
		Period: pkg.PeriodDay,
		Tags:   map[string]string{"author": "!John"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 3, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 4, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 3, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 4, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 4, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 4, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.3333, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:   util.PastDay(4),
		To:     util.Today(),
		Period: pkg.PeriodDay,
		Tags:   map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 0, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 0, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 1, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 0, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:   util.PastDay(4),
		To:     util.Today(),
		Period: pkg.PeriodDay,
		Tags:   map[string]string{"author": "!all"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 4, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 6, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 4, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 7, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 6, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 5, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 2, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 0.8333, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	asd, err := analyzer.Time.AvgSessionDuration(nil)
	assert.NoError(t, err)
	assert.Len(t, asd, 2)
	assert.Equal(t, util.PastDay(4), asd[0].Day.Time)
	assert.Equal(t, util.PastDay(2), asd[1].Day.Time)
	assert.Equal(t, 300, asd[0].AverageTimeSpentSeconds)
	assert.Equal(t, 450, asd[1].AverageTimeSpentSeconds)
	tsd, err := analyzer.Visitors.totalSessionDuration(&Filter{
		Ctx: context.Background(),
	})
	assert.NoError(t, err)
	assert.Equal(t, 1200, tsd)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{From: util.PastDay(4), To: util.PastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, util.PastDay(4), visitors[0].Day.Time)
	assert.Equal(t, util.PastDay(2), visitors[2].Day.Time)
	asd, err = analyzer.Time.AvgSessionDuration(&Filter{From: util.PastDay(3), To: util.PastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, asd, 3)
	tsd, err = analyzer.Visitors.totalSessionDuration(&Filter{
		Ctx:  context.Background(),
		From: util.PastDay(3),
		To:   util.PastDay(1),
	})
	assert.NoError(t, err)
	assert.Equal(t, 900, tsd)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:   util.PastDay(90),
		To:     util.Today(),
		Period: pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	assert.True(t, visitors[0].Week.Valid)
	_, err = analyzer.Visitors.ByPeriod(&Filter{
		From:   util.Today(),
		To:     util.Today(),
		Period: pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByPeriod(&Filter{
		From:   util.PastDay(1),
		To:     util.Today(),
		Period: pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByPeriod(&Filter{
		From:        util.PastDay(90),
		To:          util.Today(),
		PathPattern: []string{"(?i)^/bar"},
	})
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByPeriod(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByPeriod(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgSessionDuration(&Filter{
		From:   util.PastDay(90),
		To:     util.Today(),
		Period: pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgSessionDuration(&Filter{
		From:        util.PastDay(90),
		To:          util.Today(),
		PathPattern: []string{"(?i)^/bar"},
	})
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgSessionDuration(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgSessionDuration(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.totalSessionDuration(getMaxFilter(""))
	assert.NoError(t, err)

	// imported statistics
	past5Days := util.PastDay(5).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past5Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:                 util.PastDay(5),
		To:                   util.Today(),
		Period:               pkg.PeriodDay,
		ImportedUntil:        util.PastDay(4),
		IncludeCR:            true,
		Sample:               10_000,
		MaxTimeOnPageSeconds: 600,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, util.PastDay(5), visitors[0].Day.Time)
	assert.Equal(t, util.PastDay(4), visitors[1].Day.Time)
	assert.Equal(t, util.PastDay(3), visitors[2].Day.Time)
	assert.Equal(t, util.PastDay(2), visitors[3].Day.Time)
	assert.Equal(t, util.PastDay(1), visitors[4].Day.Time)
	assert.Equal(t, util.Today(), visitors[5].Day.Time)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 4, visitors[1].Visitors)
	assert.Equal(t, 0, visitors[2].Visitors)
	assert.Equal(t, 4, visitors[3].Visitors)
	assert.Equal(t, 0, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.Equal(t, 3, visitors[0].Sessions)
	assert.Equal(t, 6, visitors[1].Sessions)
	assert.Equal(t, 0, visitors[2].Sessions)
	assert.Equal(t, 4, visitors[3].Sessions)
	assert.Equal(t, 0, visitors[4].Sessions)
	assert.Equal(t, 1, visitors[5].Sessions)
	assert.Equal(t, 4, visitors[0].Views)
	assert.Equal(t, 7, visitors[1].Views)
	assert.Equal(t, 0, visitors[2].Views)
	assert.Equal(t, 6, visitors[3].Views)
	assert.Equal(t, 0, visitors[4].Views)
	assert.Equal(t, 1, visitors[5].Views)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 5, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.Equal(t, 2, visitors[3].Bounces)
	assert.Equal(t, 0, visitors[4].Bounces)
	assert.Equal(t, 1, visitors[5].Bounces)
	assert.InDelta(t, 0.3333, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.8333, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[0].CR, 0.01)
	assert.InDelta(t, 1, visitors[1].CR, 0.01)
	assert.InDelta(t, 0, visitors[2].CR, 0.01)
	assert.InDelta(t, 1, visitors[3].CR, 0.01)
	assert.InDelta(t, 0, visitors[4].CR, 0.01)
	assert.InDelta(t, 1, visitors[5].CR, 0.01)
	f := &Filter{
		Ctx:           context.Background(),
		From:          util.PastDay(10),
		To:            util.Today(),
		ImportedUntil: util.PastDay(4),
	}
	f.validate()
	tsd, err = analyzer.Visitors.totalSessionDuration(f)
	assert.NoError(t, err)
	assert.Equal(t, 1400, tsd)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:                 util.PastDay(5),
		To:                   util.Today(),
		Period:               pkg.PeriodWeek,
		ImportedUntil:        util.PastDay(4),
		IncludeCR:            true,
		Sample:               10_000,
		MaxTimeOnPageSeconds: 600,
	})
	assert.NoError(t, err)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:                 util.PastDay(5),
		To:                   util.Today(),
		Period:               pkg.PeriodMonth,
		ImportedUntil:        util.PastDay(4),
		IncludeCR:            true,
		Sample:               10_000,
		MaxTimeOnPageSeconds: 600,
	})
	assert.NoError(t, err)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:                 util.PastDay(5),
		To:                   util.Today(),
		Period:               pkg.PeriodYear,
		ImportedUntil:        util.PastDay(4),
		IncludeCR:            true,
		Sample:               10_000,
		MaxTimeOnPageSeconds: 600,
	})
	assert.NoError(t, err)
}

func TestAnalyzer_ByPeriodCustomMetric(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 5, Time: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Start: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), Path: "/"},
		{VisitorID: 2, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Path: "/foo"},
		{VisitorID: 3, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Path: "/bar"},
		{VisitorID: 4, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Path: "/"},
		{VisitorID: 5, Time: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Path: "/"},
		{VisitorID: 6, Time: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Path: "/foo"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.89", "EUR"}, Path: "/"},
		{VisitorID: 3, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"3.12", "EUR"}, Path: "/bar"},
		{VisitorID: 4, Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.77", "USD"}, Path: "/"},
		{VisitorID: 6, Time: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC), Name: "Sale", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "2.98"}, Path: "/foo"},
		{VisitorID: 6, Time: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Name: "Unrelated", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "99"}, Path: "/foo"},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByPeriod(&Filter{
		From:             time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),
		To:               time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
		EventName:        []string{"Sale"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), visitors[0].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), visitors[1].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), visitors[2].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC), visitors[3].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), visitors[4].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC), visitors[5].Day.Time)
	assert.InDelta(t, 1.89, visitors[0].CustomMetricAvg, 0.001)
	assert.InDelta(t, 2.445, visitors[1].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[5].CustomMetricAvg, 0.001)
	assert.InDelta(t, 1.89, visitors[0].CustomMetricTotal, 0.001)
	assert.InDelta(t, 4.89, visitors[1].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[5].CustomMetricTotal, 0.001)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:             time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC),
		To:               time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC), visitors[5].Day.Time)
	assert.InDelta(t, 1.89, visitors[0].CustomMetricAvg, 0.001)
	assert.InDelta(t, 3.12, visitors[1].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[5].CustomMetricAvg, 0.001)
	assert.InDelta(t, 1.89, visitors[0].CustomMetricTotal, 0.001)
	assert.InDelta(t, 3.12, visitors[1].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[5].CustomMetricTotal, 0.001)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:             time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
		To:               time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
		Path:             []string{"/"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), visitors[5].Day.Time)
	assert.InDelta(t, 0, visitors[0].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[1].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[5].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[0].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[1].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[5].CustomMetricTotal, 0.001)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:             time.Date(2023, 12, 22, 0, 0, 0, 0, time.UTC),
		To:               time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
		Path:             []string{"/"},
		Period:           pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.InDelta(t, 0, visitors[0].CR, 0.001)
	assert.InDelta(t, 0, visitors[1].CR, 0.001)
	assert.InDelta(t, 0.25, visitors[2].CR, 0.001)
	assert.InDelta(t, 0, visitors[3].CR, 0.001)
	assert.InDelta(t, 0, visitors[0].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[1].CustomMetricAvg, 0.001)
	assert.InDelta(t, 1.89, visitors[2].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[0].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[1].CustomMetricTotal, 0.001)
	assert.InDelta(t, 1.89, visitors[2].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricTotal, 0.001)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:      time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		EventName: []string{"Sale"},
		IncludeCR: true,
		Period:    pkg.PeriodMonth,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:      time.Date(2024, 10, 5, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		EventName: []string{"Sale"},
		IncludeCR: true,
		Period:    pkg.PeriodYear,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	filter := getMaxFilter("Sale")
	filter.CustomMetricType = pkg.CustomMetricTypeFloat
	filter.CustomMetricKey = "amount"
	filter.From = time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)
	filter.To = time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	visitors, err = analyzer.Visitors.ByPeriod(filter)
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)

	// imported statistics
	past5Days := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past5Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{
		From:             time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
		To:               time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		ImportedUntil:    time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), visitors[0].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC), visitors[1].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC), visitors[2].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), visitors[3].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC), visitors[4].Day.Time)
	assert.Equal(t, time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), visitors[5].Day.Time)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 0, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 0, visitors[5].Visitors)
	assert.Equal(t, 3, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 0, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 0, visitors[5].Sessions)
	assert.Equal(t, 4, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 0, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 0, visitors[5].Views)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 0, visitors[4].Bounces)
	assert.Equal(t, 0, visitors[5].Bounces)
	assert.InDelta(t, 0.3333, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[0].CR, 0.01)
	assert.InDelta(t, 0, visitors[1].CR, 0.01)
	assert.InDelta(t, 0, visitors[2].CR, 0.01)
	assert.InDelta(t, 0, visitors[3].CR, 0.01)
	assert.InDelta(t, 1, visitors[4].CR, 0.01)
	assert.InDelta(t, 0, visitors[5].CR, 0.01)
}

func TestAnalyzer_ByHour(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(2).Add(time.Hour * 3), Start: time.Now(), ExitPath: "/foo", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(2).Add(time.Hour * 3), Start: time.Now(), ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(2).Add(time.Hour * 3), Start: time.Now(), ExitPath: "/", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(2).Add(time.Hour * 8), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(1).Add(time.Hour * 4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(1).Add(time.Hour * 5), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(1).Add(time.Hour * 8), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.Today().Add(time.Hour * 5), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 7, Time: util.Today().Add(time.Hour * 10), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(2).Add(time.Hour*2 + time.Minute*30), Path: "/foo", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(2).Add(time.Hour * 3), Path: "/"},
		{VisitorID: 2, Time: util.PastDay(2).Add(time.Hour * 8), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 3, Time: util.PastDay(1).Add(time.Hour * 4), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 4, Time: util.PastDay(1).Add(time.Hour * 5), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: util.PastDay(1).Add(time.Hour * 8), Path: "/"},
		{VisitorID: 6, Time: util.Today().Add(time.Hour * 5), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 7, Time: util.Today().Add(time.Hour * 10), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByHour(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 2, visitors[5].Visitors)
	assert.Equal(t, 2, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)
	assert.Equal(t, 1, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 2, visitors[5].Views)
	assert.Equal(t, 2, visitors[8].Views)
	assert.Equal(t, 1, visitors[10].Views)
	assert.Equal(t, 1, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 2, visitors[5].Sessions)
	assert.Equal(t, 2, visitors[8].Sessions)
	assert.Equal(t, 1, visitors[10].Sessions)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.Equal(t, 2, visitors[5].Bounces)
	assert.Equal(t, 2, visitors[8].Bounces)
	assert.Equal(t, 1, visitors[10].Bounces)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[8].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[10].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByHour(&Filter{From: util.PastDay(1), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 2, visitors[5].Visitors)
	assert.Equal(t, 1, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 2, visitors[5].Views)
	assert.Equal(t, 1, visitors[8].Views)
	assert.Equal(t, 1, visitors[10].Views)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 2, visitors[5].Sessions)
	assert.Equal(t, 1, visitors[8].Sessions)
	assert.Equal(t, 1, visitors[10].Sessions)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.Equal(t, 2, visitors[5].Bounces)
	assert.Equal(t, 1, visitors[8].Bounces)
	assert.Equal(t, 1, visitors[10].Bounces)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[8].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[10].BounceRate, 0.01)

	visitors, err = analyzer.Visitors.ByHour(&Filter{
		From: util.PastDay(1),
		To:   util.Today(),
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.Equal(t, 0, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 1, visitors[5].Views)
	assert.Equal(t, 0, visitors[8].Views)
	assert.Equal(t, 1, visitors[10].Views)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 1, visitors[5].Sessions)
	assert.Equal(t, 0, visitors[8].Sessions)
	assert.Equal(t, 1, visitors[10].Sessions)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.Equal(t, 1, visitors[5].Bounces)
	assert.Equal(t, 0, visitors[8].Bounces)
	assert.Equal(t, 1, visitors[10].Bounces)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[8].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[10].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByHour(&Filter{
		From: util.PastDay(1),
		To:   util.Today(),
		Tags: map[string]string{"author": "!John"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 0, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.Equal(t, 1, visitors[8].Visitors)
	assert.Equal(t, 0, visitors[10].Visitors)
	assert.Equal(t, 0, visitors[4].Views)
	assert.Equal(t, 1, visitors[5].Views)
	assert.Equal(t, 1, visitors[8].Views)
	assert.Equal(t, 0, visitors[10].Views)
	assert.Equal(t, 0, visitors[4].Sessions)
	assert.Equal(t, 1, visitors[5].Sessions)
	assert.Equal(t, 1, visitors[8].Sessions)
	assert.Equal(t, 0, visitors[10].Sessions)
	assert.Equal(t, 0, visitors[4].Bounces)
	assert.Equal(t, 1, visitors[5].Bounces)
	assert.Equal(t, 1, visitors[8].Bounces)
	assert.Equal(t, 0, visitors[10].Bounces)
	assert.InDelta(t, 0, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[8].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[10].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByHour(&Filter{
		From: util.PastDay(1),
		To:   util.Today(),
		Tags: map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 0, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.Equal(t, 0, visitors[8].Visitors)
	assert.Equal(t, 0, visitors[10].Visitors)
	assert.Equal(t, 0, visitors[4].Views)
	assert.Equal(t, 1, visitors[5].Views)
	assert.Equal(t, 0, visitors[8].Views)
	assert.Equal(t, 0, visitors[10].Views)
	assert.Equal(t, 0, visitors[4].Sessions)
	assert.Equal(t, 1, visitors[5].Sessions)
	assert.Equal(t, 0, visitors[8].Sessions)
	assert.Equal(t, 0, visitors[10].Sessions)
	assert.Equal(t, 0, visitors[4].Bounces)
	assert.Equal(t, 1, visitors[5].Bounces)
	assert.Equal(t, 0, visitors[8].Bounces)
	assert.Equal(t, 0, visitors[10].Bounces)
	assert.InDelta(t, 0, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[8].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[10].BounceRate, 0.01)

	_, err = analyzer.Visitors.ByHour(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByHour(getMaxFilter("event"))
	assert.NoError(t, err)

	// imported statistics
	past3Days := util.PastDay(3).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past3Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.ByHour(&Filter{
		From:          util.PastDay(3),
		To:            util.PastDay(3),
		Period:        pkg.PeriodDay,
		ImportedUntil: util.PastDay(2),
		IncludeCR:     true,
		Sample:        10_000,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 0, visitors[0].Hour)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 4, visitors[0].Views)
	assert.Equal(t, 3, visitors[0].Sessions)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.3333, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[0].CR, 0.01)

	for i := 1; i < 24; i++ {
		assert.Equal(t, i, visitors[i].Hour)
		assert.Zero(t, visitors[i].Visitors)
		assert.Zero(t, visitors[i].Views)
		assert.Zero(t, visitors[i].Sessions)
		assert.Zero(t, visitors[i].Bounces)
		assert.InDelta(t, 0, visitors[i].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[i].CR, 0.01)
	}
}

func TestAnalyzer_ByHourEvent(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Hour*3 + time.Minute*15), Start: time.Now(), PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Hour * 5), Start: time.Now(), PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.Today().Add(time.Hour*6 + time.Minute*15), Start: time.Now(), PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Hour * 15), Start: time.Now(), PageViews: 1, IsBounce: true},
			{Sign: -1, VisitorID: 4, Time: util.Today().Add(time.Hour * 15), Start: time.Now(), PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Hour*15 + time.Minute*50), Start: time.Now(), PageViews: 2, IsBounce: false},
			{Sign: -1, VisitorID: 4, Time: util.Today().Add(time.Hour*15 + time.Minute*50), Start: time.Now(), PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Hour*15 + time.Minute*55), Start: time.Now(), PageViews: 3, IsBounce: false},
			{Sign: -1, VisitorID: 4, Time: util.Today().Add(time.Hour*15 + time.Minute*55), Start: time.Now(), PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Hour * 16), Start: time.Now(), PageViews: 4, IsBounce: false},
			{Sign: -1, VisitorID: 4, Time: util.Today().Add(time.Hour * 16), Start: time.Now(), PageViews: 4, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Hour*16 + time.Minute*5), Start: time.Now(), PageViews: 5, IsBounce: false},
			{Sign: 1, VisitorID: 5, Time: util.Today().Add(time.Hour * 19), Start: time.Now(), PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today().Add(time.Hour * 3), Path: "/"},
		{VisitorID: 1, Time: util.Today().Add(time.Hour*3 + time.Minute*15), Path: "/pricing"},
		{VisitorID: 2, Time: util.Today().Add(time.Hour * 5), Path: "/"},
		{VisitorID: 3, Time: util.Today().Add(time.Hour * 6), Path: "/"},
		{VisitorID: 3, Time: util.Today().Add(time.Hour*6 + time.Minute*15), Path: "/pricing"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour * 15), Path: "/"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour*15 + time.Minute*50), Path: "/about"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour*15 + time.Minute*55), Path: "/blog"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour * 16), Path: "/blog/1"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour*16 + time.Minute*5), Path: "/blog/2"},
		{VisitorID: 5, Time: util.Today().Add(time.Hour * 19), Path: "/"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: util.Today().Add(time.Hour*3 + time.Minute*10), Name: "event"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour*15 + time.Minute*52), Name: "event"},
		{VisitorID: 5, Time: util.Today().Add(time.Hour*19 + time.Minute), Name: "event"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByHour(&Filter{
		From:      util.Today(),
		To:        util.Today(),
		EventName: []string{"event"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Zero(t, visitors[0].Visitors)
	assert.Zero(t, visitors[1].Visitors)
	assert.Zero(t, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Zero(t, visitors[4].Visitors)
	assert.Zero(t, visitors[5].Visitors)
	assert.Zero(t, visitors[6].Visitors)
	assert.Zero(t, visitors[7].Visitors)
	assert.Zero(t, visitors[8].Visitors)
	assert.Zero(t, visitors[9].Visitors)
	assert.Zero(t, visitors[10].Visitors)
	assert.Zero(t, visitors[11].Visitors)
	assert.Zero(t, visitors[12].Visitors)
	assert.Zero(t, visitors[13].Visitors)
	assert.Zero(t, visitors[14].Visitors)
	assert.Equal(t, 1, visitors[15].Visitors)
	assert.Zero(t, visitors[16].Visitors)
	assert.Zero(t, visitors[17].Visitors)
	assert.Zero(t, visitors[18].Visitors)
	assert.Equal(t, 1, visitors[19].Visitors)
	assert.Zero(t, visitors[20].Visitors)
	assert.Zero(t, visitors[21].Visitors)
	assert.Zero(t, visitors[22].Visitors)
	assert.Zero(t, visitors[23].Visitors)
	total, err := analyzer.Visitors.Total(&Filter{
		From:      util.Today(),
		To:        util.Today(),
		EventName: []string{"event"},
		IncludeCR: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, total.Visitors)
}

func TestAnalyzer_ByHourCRAndCustomMetric(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Hour * 5), Start: util.Today().Add(time.Hour * 5), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today().Add(time.Hour * 5), Start: util.Today().Add(time.Hour * 5), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Hour * 5), Start: util.Today().Add(time.Hour * 5), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Hour * 5), Start: util.Today().Add(time.Hour * 5), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.Today().Add(time.Hour * 14), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Hour * 14), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 5, Time: util.Today().Add(time.Hour * 21), Start: util.Today().Add(time.Hour * 21), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.Today().Add(time.Hour * 21), Start: util.Today().Add(time.Hour * 21), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today().Add(time.Hour * 5), Path: "/"},
		{VisitorID: 2, Time: util.Today().Add(time.Hour * 5), Path: "/foo"},
		{VisitorID: 3, Time: util.Today().Add(time.Hour * 14), Path: "/bar"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour * 14), Path: "/"},
		{VisitorID: 5, Time: util.Today().Add(time.Hour * 21), Path: "/"},
		{VisitorID: 6, Time: util.Today().Add(time.Hour * 21), Path: "/foo"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: util.Today().Add(time.Hour * 5), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.89", "EUR"}, Path: "/"},
		{VisitorID: 3, Time: util.Today().Add(time.Hour * 14), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"3.12", "EUR"}, Path: "/bar"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour * 14), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.77", "USD"}, Path: "/"},
		{VisitorID: 6, Time: util.Today().Add(time.Hour * 21), Name: "Sale", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "2.98"}, Path: "/foo"},
		{VisitorID: 6, Time: util.Today().Add(time.Hour * 21), Name: "Unrelated", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "99"}, Path: "/foo"},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByHour(&Filter{
		From:             util.Today(),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.InDelta(t, 0, visitors[0].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[1].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricAvg, 0.001)
	assert.InDelta(t, 1.89, visitors[5].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[6].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[7].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[8].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[9].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[10].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[11].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[12].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[13].CustomMetricAvg, 0.001)
	assert.InDelta(t, 2.445, visitors[14].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[15].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[16].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[17].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[18].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[19].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[20].CustomMetricAvg, 0.001)
	assert.InDelta(t, 2.98, visitors[21].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[22].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[23].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[0].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[1].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[2].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[3].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[4].CustomMetricTotal, 0.001)
	assert.InDelta(t, 1.89, visitors[5].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[6].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[7].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[8].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[9].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[10].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[11].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[12].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[13].CustomMetricTotal, 0.001)
	assert.InDelta(t, 4.89, visitors[14].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[15].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[16].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[17].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[18].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[19].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[20].CustomMetricTotal, 0.001)
	assert.InDelta(t, 2.98, visitors[21].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[22].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[23].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[0].CR, 0.001)
	assert.InDelta(t, 0, visitors[1].CR, 0.001)
	assert.InDelta(t, 0, visitors[2].CR, 0.001)
	assert.InDelta(t, 0, visitors[3].CR, 0.001)
	assert.InDelta(t, 0, visitors[4].CR, 0.001)
	assert.InDelta(t, 0.5, visitors[5].CR, 0.001)
	assert.InDelta(t, 0, visitors[6].CR, 0.001)
	assert.InDelta(t, 0, visitors[7].CR, 0.001)
	assert.InDelta(t, 0, visitors[8].CR, 0.001)
	assert.InDelta(t, 0, visitors[9].CR, 0.001)
	assert.InDelta(t, 0, visitors[10].CR, 0.001)
	assert.InDelta(t, 0, visitors[11].CR, 0.001)
	assert.InDelta(t, 0, visitors[12].CR, 0.001)
	assert.InDelta(t, 0, visitors[13].CR, 0.001)
	assert.InDelta(t, 1, visitors[14].CR, 0.001)
	assert.InDelta(t, 0, visitors[15].CR, 0.001)
	assert.InDelta(t, 0, visitors[16].CR, 0.001)
	assert.InDelta(t, 0, visitors[17].CR, 0.001)
	assert.InDelta(t, 0, visitors[18].CR, 0.001)
	assert.InDelta(t, 0, visitors[19].CR, 0.001)
	assert.InDelta(t, 0, visitors[20].CR, 0.001)
	assert.InDelta(t, 0.5, visitors[21].CR, 0.001)
	assert.InDelta(t, 0, visitors[22].CR, 0.001)
	assert.InDelta(t, 0, visitors[23].CR, 0.001)
	visitors, err = analyzer.Visitors.ByHour(&Filter{
		From:             util.Today(),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.InDelta(t, 1.89, visitors[5].CustomMetricAvg, 0.001)
	assert.InDelta(t, 3.12, visitors[14].CustomMetricAvg, 0.001)
	assert.InDelta(t, 2.98, visitors[21].CustomMetricAvg, 0.001)
	assert.InDelta(t, 1.89, visitors[5].CustomMetricTotal, 0.001)
	assert.InDelta(t, 3.12, visitors[14].CustomMetricTotal, 0.001)
	assert.InDelta(t, 2.98, visitors[21].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0.5, visitors[5].CR, 0.001)
	assert.InDelta(t, 0.5, visitors[14].CR, 0.001)
	assert.InDelta(t, 0.5, visitors[21].CR, 0.001)
	visitors, err = analyzer.Visitors.ByHour(&Filter{
		From:             util.Today(),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
		Path:             []string{"/"},
	})
	assert.InDelta(t, 1.89, visitors[5].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[14].CustomMetricAvg, 0.001)
	assert.InDelta(t, 0, visitors[21].CustomMetricAvg, 0.001)
	assert.InDelta(t, 1.89, visitors[5].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[14].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0, visitors[21].CustomMetricTotal, 0.001)
	assert.InDelta(t, 0.5, visitors[5].CR, 0.001)
	assert.InDelta(t, 0, visitors[14].CR, 0.001)
	assert.InDelta(t, 0, visitors[21].CR, 0.001)
	filter := getMaxFilter("Sale")
	filter.CustomMetricType = pkg.CustomMetricTypeFloat
	filter.CustomMetricKey = "amount"
	filter.From = util.Today()
	filter.To = util.Today()
	_, err = analyzer.Visitors.ByHour(filter)
	assert.NoError(t, err)
}

func TestAnalyzer_ByHourTimeShift(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(1).Add(time.Hour * 3), Start: time.Now(), PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(1).Add(time.Hour * 3), Start: time.Now(), PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(1).Add(time.Hour * 4), Start: time.Now(), PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(1).Add(time.Hour * 5), Start: time.Now(), PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(1).Add(time.Hour * 6), Start: time.Now(), PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(1).Add(time.Hour * 3), Path: "/"},
		{VisitorID: 1, Time: util.PastDay(1).Add(time.Hour * 4), Path: "/"},
		{VisitorID: 2, Time: util.PastDay(1).Add(time.Hour * 5), Path: "/"},
		{VisitorID: 3, Time: util.PastDay(1).Add(time.Hour * 6), Path: "/"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{Name: "event", VisitorID: 1, Time: util.PastDay(1).Add(time.Hour * 4), Path: "/"},
		{Name: "event", VisitorID: 2, Time: util.PastDay(1).Add(time.Hour * 5), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByHour(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.Equal(t, 1, visitors[6].Visitors)
	assert.Equal(t, 1, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 1, visitors[5].Views)
	assert.Equal(t, 1, visitors[6].Views)
	assert.Equal(t, 1, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 1, visitors[5].Sessions)
	assert.Equal(t, 1, visitors[6].Sessions)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 0, visitors[4].Bounces)
	assert.Equal(t, 1, visitors[5].Bounces)
	assert.Equal(t, 1, visitors[6].Bounces)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[6].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByHour(&Filter{
		EventName: []string{"event"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.Equal(t, 0, visitors[6].Visitors)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 1, visitors[5].Views)
	assert.Equal(t, 0, visitors[6].Views)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 1, visitors[5].Sessions)
	assert.Equal(t, 0, visitors[6].Sessions)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 0, visitors[4].Bounces)
	assert.Equal(t, 1, visitors[5].Bounces)
	assert.Equal(t, 0, visitors[6].Bounces)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[6].BounceRate, 0.01)
}

func TestAnalyzer_ByMinute(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), ExitPath: "/foo", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today(), Start: time.Now(), ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Minute), Start: time.Now(), ExitPath: "/", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Minute * 8), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.Today().Add(time.Minute * 4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Minute * 5), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.Today().Add(time.Minute * 8), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.Today().Add(time.Minute * 5), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 7, Time: util.Today().Add(time.Minute * 10), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/foo"},
		{VisitorID: 1, Time: util.Today().Add(time.Minute), Path: "/"},
		{VisitorID: 2, Time: util.Today().Add(time.Minute * 8), Path: "/"},
		{VisitorID: 3, Time: util.Today().Add(time.Minute * 4), Path: "/"},
		{VisitorID: 4, Time: util.Today().Add(time.Minute * 5), Path: "/"},
		{VisitorID: 5, Time: util.Today().Add(time.Minute * 8), Path: "/"},
		{VisitorID: 6, Time: util.Today().Add(time.Minute * 5), Path: "/"},
		{VisitorID: 7, Time: util.Today().Add(time.Minute * 10), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByMinute(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 60)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 2, visitors[5].Visitors)
	assert.Equal(t, 2, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)
	assert.Equal(t, 1, visitors[1].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 2, visitors[5].Views)
	assert.Equal(t, 2, visitors[8].Views)
	assert.Equal(t, 1, visitors[10].Views)
	assert.Equal(t, 1, visitors[1].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 2, visitors[5].Sessions)
	assert.Equal(t, 2, visitors[8].Sessions)
	assert.Equal(t, 1, visitors[10].Sessions)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.Equal(t, 2, visitors[5].Bounces)
	assert.Equal(t, 2, visitors[8].Bounces)
	assert.Equal(t, 1, visitors[10].Bounces)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[8].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[10].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.ByMinute(&Filter{From: util.Today(), To: util.Today().Add(time.Hour)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 60)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 2, visitors[5].Visitors)
	assert.Equal(t, 2, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 2, visitors[5].Views)
	assert.Equal(t, 2, visitors[8].Views)
	assert.Equal(t, 1, visitors[10].Views)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 2, visitors[5].Sessions)
	assert.Equal(t, 2, visitors[8].Sessions)
	assert.Equal(t, 1, visitors[10].Sessions)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.Equal(t, 2, visitors[5].Bounces)
	assert.Equal(t, 2, visitors[8].Bounces)
	assert.Equal(t, 1, visitors[10].Bounces)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[5].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[8].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[10].BounceRate, 0.01)
	_, err = analyzer.Visitors.ByMinute(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByMinute(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_ByWeekdayAndHour(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Hour), Start: time.Now(), ExitPath: "/", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Hour * 8), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.Today().Add(time.Hour * 4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.Today().Add(time.Hour * 5), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.Today().Add(time.Hour * 8), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.Today().Add(time.Hour * 5), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 7, Time: util.Today().Add(time.Hour * 10), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/foo"},
		{VisitorID: 1, Time: util.Today().Add(time.Hour), Path: "/"},
		{VisitorID: 2, Time: util.Today().Add(time.Hour * 8), Path: "/"},
		{VisitorID: 3, Time: util.Today().Add(time.Hour * 4), Path: "/"},
		{VisitorID: 4, Time: util.Today().Add(time.Hour * 5), Path: "/"},
		{VisitorID: 5, Time: util.Today().Add(time.Hour * 8), Path: "/"},
		{VisitorID: 6, Time: util.Today().Add(time.Hour * 5), Path: "/"},
		{VisitorID: 7, Time: util.Today().Add(time.Hour * 10), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByWeekdayAndHour(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 168)
	visitors, err = analyzer.Visitors.ByWeekdayAndHour(&Filter{WeekdayMode: WeekdaySunday})
	assert.NoError(t, err)
	assert.Len(t, visitors, 168)
	_, err = analyzer.Visitors.ByWeekdayAndHour(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByWeekdayAndHour(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_Growth(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(9).Add(time.Minute * 15), Start: time.Now(), SessionID: 4, ExitPath: "/bar", DurationSeconds: 600, PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(9), Start: time.Now(), ExitPath: "/", PageViews: 5, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4).Add(time.Minute * 15), Start: time.Now(), SessionID: 4, ExitPath: "/bar", DurationSeconds: 600, PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(4), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), SessionID: 3, ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), SessionID: 3, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(3).Add(time.Minute * 5), Start: time.Now(), SessionID: 3, ExitPath: "/foo", DurationSeconds: 300, PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(3), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(3), Start: time.Now(), SessionID: 3, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), SessionID: 31, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: util.PastDay(3), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: util.PastDay(3), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 10, Time: util.PastDay(2).Add(time.Minute * 5), Start: time.Now(), SessionID: 2, ExitPath: "/bar", DurationSeconds: 300, PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 11, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 12, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 13, Time: util.Today(), Start: time.Now(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Visitors.Growth(nil)
	assert.ErrorIs(t, err, ErrNoPeriodOrDay)
	assert.Nil(t, growth)
	growth, err = analyzer.Visitors.Growth(&Filter{From: util.PastDay(2), To: util.PastDay(2), IncludeCR: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0.5, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, -0.5, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 0.5, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0.3333, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, -0.5, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 0, growth.CRGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{From: util.PastDay(3), To: util.PastDay(2), IncludeCR: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1.3333, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1.2, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 2, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0.1666, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 0, growth.CRGrowth, 0.001)
	_, err = analyzer.Visitors.Growth(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.Growth(getMaxFilter("event"))
	assert.NoError(t, err)

	// imported statistics
	past10Days := util.PastDay(10).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_visitors" (date, visitors, views, sessions, bounces, session_duration) VALUES
		('%s', 2, 4, 3, 1, 200)`, past10Days))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	growth, err = analyzer.Visitors.Growth(&Filter{
		From:          util.PastDay(5),
		To:            util.Today(),
		ImportedUntil: util.PastDay(9),
		IncludeCR:     true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 2.25, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 0.75, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 2.2, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0.7187, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0.5, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 0, growth.CRGrowth, 0.001)
}

func TestAnalyzer_GrowthDay(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions([]model.Session{
		{Sign: 1, VisitorID: 1, Time: util.PastDay(8).Add(time.Hour * 5), Start: time.Now()},
		{Sign: 1, VisitorID: 2, Time: util.PastDay(1).Add(time.Hour * 3), Start: time.Now()},
		{Sign: 1, VisitorID: 3, Time: util.PastDay(1).Add(time.Hour * 4), Start: time.Now()},
		{Sign: 1, VisitorID: 4, Time: util.PastDay(1).Add(time.Hour * 9), Start: time.Now()},
		{Sign: 1, VisitorID: 5, Time: util.Today().Add(time.Hour * 4), Start: time.Now()},
		{Sign: 1, VisitorID: 6, Time: util.Today().Add(time.Hour * 9), Start: time.Now()},
		{Sign: 1, VisitorID: 7, Time: util.Today().Add(time.Hour * 12), Start: time.Now()},
		{Sign: 1, VisitorID: 8, Time: util.Today().Add(time.Hour * 17), Start: time.Now()},
		{Sign: 1, VisitorID: 9, Time: util.Today().Add(time.Hour * 21), Start: time.Now()},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)

	// Testing for today is hard because it would require messing with the time.Now function.
	// I don't want to do that, so let's assume it works (tested in debug mode) and just get no error for today.
	growth, err := analyzer.Visitors.Growth(&Filter{From: util.Today(), To: util.Today()})
	assert.NoError(t, err)
	assert.NotNil(t, growth)

	growth, err = analyzer.Visitors.Growth(&Filter{From: util.PastDay(1), To: util.PastDay(1)})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 2, growth.VisitorsGrowth, 0.001)
}

func TestAnalyzer_GrowthDayFirstHour(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions([]model.Session{
		{Sign: 1, VisitorID: 1, Time: util.PastDay(1), Start: time.Now()},
		{Sign: 1, VisitorID: 2, Time: util.PastDay(1).Add(time.Hour * 4), Start: time.Now()},
		{Sign: 1, VisitorID: 3, Time: util.Today(), Start: time.Now()},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Visitors.Growth(&Filter{From: util.Today(), To: util.Today().Add(time.Hour * 4), IncludeTime: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, -0.5, growth.VisitorsGrowth, 0.01)
	growth, err = analyzer.Visitors.Growth(&Filter{From: util.Today(), To: util.Today().Add(time.Hour * 2), IncludeTime: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0, growth.VisitorsGrowth, 0.01)
}

func TestAnalyzer_GrowthNoData(t *testing.T) {
	db.CleanupDB(t, dbClient)
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Visitors.Growth(&Filter{From: util.PastDay(7), To: util.PastDay(7)})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 0, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 0, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	_, err = analyzer.Visitors.Growth(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.Growth(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_GrowthEvents(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 12, SessionID: 3, Time: util.PastDay(9).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: -1, VisitorID: 12, SessionID: 3, Time: util.PastDay(9).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 12, SessionID: 3, Time: util.PastDay(9).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 12, Time: util.PastDay(9).Add(time.Second * 5), EntryPath: "/", Start: time.Now(), ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 13, SessionID: 3, Time: util.PastDay(9).Add(time.Second * 6), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 13, SessionID: 31, Time: util.PastDay(9).Add(time.Minute * 10), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 14, Time: util.PastDay(9).Add(time.Second * 7), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 15, Time: util.PastDay(9).Add(time.Second * 8), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},

			{Sign: 1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(-time.Second), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(-time.Second), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: -1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(time.Minute * 15), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4).Add(time.Second * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},

			{Sign: 1, VisitorID: 4, SessionID: 3, Time: util.PastDay(3).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: -1, VisitorID: 4, SessionID: 3, Time: util.PastDay(3).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 4, SessionID: 3, Time: util.PastDay(3).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(3).Add(time.Second * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 5, SessionID: 3, Time: util.PastDay(3).Add(time.Second * 6), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 5, SessionID: 31, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(3).Add(time.Second * 7), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(3).Add(time.Second * 8), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},

			{Sign: 1, VisitorID: 8, SessionID: 2, Time: util.PastDay(2).Add(time.Second * 9), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: -1, VisitorID: 8, SessionID: 2, Time: util.PastDay(2).Add(time.Second * 9), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 8, SessionID: 2, Time: util.PastDay(2).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 2},
			{Sign: 1, VisitorID: 9, Time: util.PastDay(2).Add(time.Second * 10), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 10, Time: util.PastDay(2).Add(time.Second * 11), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},

			{Sign: 1, VisitorID: 11, Time: util.Today().Add(time.Second * 12), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 12, SessionID: 3, Time: util.PastDay(9), Path: "/"},
		{VisitorID: 12, SessionID: 3, Time: util.PastDay(9), Path: "/foo"},
		{VisitorID: 12, Time: util.PastDay(9), Path: "/"},
		{VisitorID: 13, SessionID: 3, Time: util.PastDay(9), Path: "/"},
		{VisitorID: 13, SessionID: 31, Time: util.PastDay(9), Path: "/bar"},
		{VisitorID: 14, Time: util.PastDay(9), Path: "/"},
		{VisitorID: 15, Time: util.PastDay(9), Path: "/"},

		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Path: "/foo"},
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Path: "/bar"},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/"},

		{VisitorID: 4, SessionID: 3, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 4, SessionID: 3, Time: util.PastDay(3), Path: "/foo"},
		{VisitorID: 4, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 5, SessionID: 3, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 5, SessionID: 31, Time: util.PastDay(3), Path: "/bar"},
		{VisitorID: 6, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 7, Time: util.PastDay(3), Path: "/"},

		{VisitorID: 8, SessionID: 2, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 8, SessionID: 2, Time: util.PastDay(2), Path: "/bar"},
		{VisitorID: 9, Time: util.PastDay(2), Path: "/"},
		{VisitorID: 10, Time: util.PastDay(2), Path: "/"},

		{VisitorID: 11, Time: util.Today(), Path: "/"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{Name: "event1", VisitorID: 13, Time: util.PastDay(9).Add(time.Second * 4), SessionID: 3, Path: "/"},
		{Name: "event1", DurationSeconds: 300, VisitorID: 14, Time: util.PastDay(9).Add(time.Minute * 5), SessionID: 3, Path: "/foo"},
		{Name: "event1", VisitorID: 14, Time: util.PastDay(9).Add(time.Second * 5), Path: "/"},
		{Name: "event1", VisitorID: 15, Time: util.PastDay(9).Add(time.Second * 6), SessionID: 3, Path: "/"},
		{Name: "event1", VisitorID: 15, Time: util.PastDay(9).Add(time.Minute * 10), SessionID: 31, Path: "/bar"},
		{Name: "event1", VisitorID: 16, Time: util.PastDay(9).Add(time.Second * 7), Path: "/"},
		{Name: "event1", VisitorID: 17, Time: util.PastDay(9).Add(time.Second * 8), Path: "/"},

		{Name: "event1", VisitorID: 1, Time: util.PastDay(4).Add(time.Second), SessionID: 4, Path: "/"},
		{Name: "event1", DurationSeconds: 300, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/foo"},
		{Name: "event1", DurationSeconds: 600, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 15), SessionID: 4, Path: "/bar"},
		{Name: "event1", VisitorID: 2, Time: util.PastDay(4).Add(time.Second * 2), Path: "/"},
		{Name: "event1", VisitorID: 3, Time: util.PastDay(4).Add(time.Second * 3), Path: "/"},
		{Name: "event1", VisitorID: 4, Time: util.PastDay(3).Add(time.Second * 4), SessionID: 3, Path: "/"},
		{Name: "event1", DurationSeconds: 300, VisitorID: 4, Time: util.PastDay(3).Add(time.Minute * 5), SessionID: 3, Path: "/foo"},
		{Name: "event1", VisitorID: 4, Time: util.PastDay(3).Add(time.Second * 5), Path: "/"},
		{Name: "event1", VisitorID: 5, Time: util.PastDay(3).Add(time.Second * 6), SessionID: 3, Path: "/"},
		{Name: "event1", VisitorID: 5, Time: util.PastDay(3).Add(time.Minute * 10), SessionID: 31, Path: "/bar"},
		{Name: "event1", VisitorID: 6, Time: util.PastDay(3).Add(time.Second * 7), Path: "/"},
		{Name: "event1", VisitorID: 7, Time: util.PastDay(3).Add(time.Second * 8), Path: "/"},
		{Name: "event1", VisitorID: 8, Time: util.PastDay(2).Add(time.Second * 9), SessionID: 2, Path: "/"},
		{Name: "event1", DurationSeconds: 300, VisitorID: 8, Time: util.PastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar"},
		{Name: "event1", VisitorID: 9, Time: util.PastDay(2).Add(time.Second * 10), Path: "/"},
		{Name: "event1", VisitorID: 10, Time: util.PastDay(2).Add(time.Second * 11), Path: "/"},
		{Name: "event1", VisitorID: 11, Time: util.Today().Add(time.Second * 12), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Visitors.Growth(nil)
	assert.ErrorIs(t, err, ErrNoPeriodOrDay)
	assert.Nil(t, growth)
	growth, err = analyzer.Visitors.Growth(&Filter{From: util.PastDay(2), To: util.PastDay(2), EventName: []string{"event1"}, IncludeCR: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0.5, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 0.5, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 1, growth.CRGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{From: util.PastDay(3), To: util.PastDay(2), EventName: []string{"event1"}, IncludeCR: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1.3333, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1.2, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 2, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, -0.3333, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 0, growth.CRGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{From: util.PastDay(3), To: util.PastDay(2), EventName: []string{"event1"}, Path: []string{"/bar"}, IncludeCR: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 1, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, -0.5, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, -0.1428, growth.CRGrowth, 0.001)
	_, err = analyzer.Visitors.Growth(getMaxFilter("event1"))
	assert.NoError(t, err)
}

func TestAnalyzer_GrowthCustomMetric(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(5), Start: util.PastDay(5), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(5), Start: util.PastDay(5), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(5), Start: util.PastDay(5), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4), Start: util.PastDay(4), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4), Start: util.PastDay(4), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(4), Start: util.PastDay(1), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
			{Sign: 1, VisitorID: 5, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: util.Today(), Start: util.Today(), EntryPath: "/foo", ExitPath: "/foo", PageViews: 1, IsBounce: false},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(5), Path: "/"},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/foo"},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/bar"},
		{VisitorID: 4, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 5, Time: util.Today(), Path: "/"},
		{VisitorID: 6, Time: util.Today(), Path: "/foo"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: util.PastDay(5), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.89", "EUR"}, Path: "/"},
		{VisitorID: 3, Time: util.PastDay(4), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"3.12", "EUR"}, Path: "/"},
		{VisitorID: 4, Time: util.PastDay(4), Name: "Sale", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"1.77", "USD"}, Path: "/"},
		{VisitorID: 6, Time: util.PastDay(1), Name: "Sale", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "2.98"}, Path: "/"},
		{VisitorID: 6, Time: util.Today(), Name: "Unrelated", MetaKeys: []string{"currency", "amount"}, MetaValues: []string{"EUR", "99"}, Path: "/"},
	}))
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Visitors.Growth(&Filter{
		From:             util.PastDay(2),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.NoError(t, err)
	assert.InDelta(t, 0.3185, growth.CustomMetricAvgGrowth, 0.001)
	assert.InDelta(t, -0.5604, growth.CustomMetricTotalGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{
		From:             util.PastDay(5),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
	})
	assert.NoError(t, err)
	assert.InDelta(t, 1, growth.CustomMetricAvgGrowth, 0.001)
	assert.InDelta(t, 1, growth.CustomMetricTotalGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{
		From:             util.PastDay(5),
		To:               util.Today(),
		EventName:        []string{"Sale"},
		EventMeta:        map[string]string{"currency": "EUR"},
		CustomMetricKey:  "amount",
		CustomMetricType: pkg.CustomMetricTypeFloat,
		IncludeCR:        true,
		Path:             []string{"/"},
	})
	assert.NoError(t, err)
	assert.InDelta(t, 1, growth.CustomMetricAvgGrowth, 0.001)
	assert.InDelta(t, 1, growth.CustomMetricTotalGrowth, 0.001)
	filter := getMaxFilter("Sale")
	filter.CustomMetricType = pkg.CustomMetricTypeFloat
	filter.CustomMetricKey = "amount"
	filter.From = util.PastDay(5)
	filter.To = util.Today()
	_, err = analyzer.Visitors.Growth(filter)
	assert.NoError(t, err)
}

func TestAnalyzer_GrowthTags(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(-time.Second), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(-time.Second), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: -1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: util.PastDay(4).Add(time.Minute * 15), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4).Add(time.Second * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 4, SessionID: 3, Time: util.PastDay(3).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: -1, VisitorID: 4, SessionID: 3, Time: util.PastDay(3).Add(time.Second * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 4, SessionID: 3, Time: util.PastDay(3).Add(time.Minute * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(3).Add(time.Second * 5), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 5, SessionID: 3, Time: util.PastDay(3).Add(time.Second * 6), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 5, SessionID: 31, Time: util.PastDay(3).Add(time.Minute * 10), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(3).Add(time.Second * 7), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 7, Time: util.PastDay(3).Add(time.Second * 8), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Path: "/foo", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 1, SessionID: 4, Time: util.PastDay(4), Path: "/bar"},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 4, SessionID: 3, Time: util.PastDay(3)},
		{VisitorID: 4, SessionID: 3, Time: util.PastDay(3), Path: "/foo", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 4, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 5, SessionID: 3, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 5, SessionID: 31, Time: util.PastDay(3), Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 6, Time: util.PastDay(3), Path: "/"},
		{VisitorID: 7, Time: util.PastDay(3), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Visitors.Growth(&Filter{
		From:      util.PastDay(5),
		To:        util.PastDay(1),
		IncludeCR: true,
		Tags:      map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 1, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 1, growth.CRGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{
		From:      util.PastDay(3),
		To:        util.PastDay(1),
		IncludeCR: true,
		Tags:      map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 1, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 1, growth.CRGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{
		From:      util.PastDay(3),
		To:        util.PastDay(1),
		IncludeCR: true,
		Tags:      map[string]string{"author": "!John"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0.3333, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 0, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 0.6666, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, 0, growth.CRGrowth, 0.001)
	growth, err = analyzer.Visitors.Growth(&Filter{
		From:      util.PastDay(3),
		To:        util.PastDay(1),
		IncludeCR: true,
		Tags:      map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 0, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 0, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	assert.InDelta(t, -0.2499, growth.CRGrowth, 0.001)
}

func TestAnalyzer_Referrer(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/exit", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/exit", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(time.Minute), Start: time.Now(), ExitPath: "/bar", Referrer: "ref3/foo", ReferrerName: "Ref3", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), ExitPath: "/", Referrer: "ref1/foo", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), ExitPath: "/", Referrer: "ref1/bar", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: time.Now(), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
		{VisitorID: 4, Time: time.Now(), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Referrer(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "Ref2", visitors[1].ReferrerName)
	assert.Equal(t, "Ref3", visitors[2].ReferrerName)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.25, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.25, visitors[2].RelativeVisitors, 0.01)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	_, err = analyzer.Visitors.Referrer(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.Referrer(getMaxFilter("event"))
	assert.NoError(t, err)
	visitors, err = analyzer.Visitors.Referrer(&Filter{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	_, err = analyzer.Visitors.Referrer(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldReferrerName,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldReferrerName,
			Input: "ref",
		},
	}})
	assert.NoError(t, err)

	// filter for referrer name
	visitors, err = analyzer.Visitors.Referrer(&Filter{ReferrerName: []string{"Ref1"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "Ref1", visitors[1].ReferrerName)
	assert.Equal(t, "ref1/bar", visitors[0].Referrer)
	assert.Equal(t, "ref1/foo", visitors[1].Referrer)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[1].BounceRate, 0.01)

	// filter for full referrer
	visitors, err = analyzer.Visitors.Referrer(&Filter{Referrer: []string{"ref1/foo"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref1/foo", visitors[0].Referrer)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)

	// filter for referrer name and full referrer
	visitors, err = analyzer.Visitors.Referrer(&Filter{ReferrerName: []string{"Ref1"}, Referrer: []string{"ref1/foo"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref1/foo", visitors[0].Referrer)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_referrer" (date, referrer, visitors, sessions, bounces) VALUES
		('%s', 'ref2/foo', 2, 3, 1), ('%s', 'ref3/foo', 1, 1, 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Sample:        10_000,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "ref1/foo", visitors[1].Referrer)
	assert.Equal(t, "ref3/foo", visitors[2].Referrer)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, "Ref1", visitors[1].ReferrerName)
	assert.Equal(t, "Ref3", visitors[2].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.InDelta(t, 0.4285, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2857, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2857, visitors[2].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 2, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.InDelta(t, 0.25, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[2].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Referrer:      []string{"ref2/foo"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.4285, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.25, visitors[0].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Referrer:      []string{"ref2/foo", "ref3/foo"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "ref3/foo", visitors[1].Referrer)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, "Ref3", visitors[1].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.InDelta(t, 0.4285, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2857, visitors[1].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.InDelta(t, 0.25, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[1].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Search: []Search{
			{
				Field: FieldReferrer,
				Input: "ref2",
			},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.4285, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.25, visitors[0].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Search: []Search{
			{
				Field: FieldReferrerName,
				Input: "ref2",
			},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.4285, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.25, visitors[0].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Search: []Search{
			{
				Field: FieldReferrer,
				Input: "ref2",
			},
			{
				Field: FieldReferrerName,
				Input: "ref3",
			},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "ref3/foo", visitors[1].Referrer)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, "Ref3", visitors[1].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.InDelta(t, 0.4285, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1428, visitors[1].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.InDelta(t, 0.25, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Referrer:      []string{"ref2/foo", "ref3/foo"},
		Limit:         1,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.4285, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.25, visitors[0].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		ReferrerName:  []string{"ref2/foo", "ref3/foo"},
		Limit:         1,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "ref2/foo", visitors[0].ReferrerName)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.2857, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.3333, visitors[0].BounceRate, 0.01)
}

func TestAnalyzer_ReferrerEvents(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/exit", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/exit", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(time.Minute), Start: time.Now(), ExitPath: "/bar", Referrer: "ref3/foo", ReferrerName: "Ref3", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), ExitPath: "/", Referrer: "ref1/foo", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), ExitPath: "/", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: time.Now(), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
		{VisitorID: 4, Time: time.Now(), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Referrer(&Filter{
		EventName: []string{"event"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[0].Sessions)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_referrer" (date, referrer, visitors, sessions, bounces) VALUES
		('%s', 'ref2/foo', 2, 3, 1), ('%s', 'ref1/bar', 1, 1, 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Sample:        10_000,
		EventName:     []string{"event"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "Ref2", visitors[0].ReferrerName)
	assert.Equal(t, "ref1/bar", visitors[1].ReferrerName)
	assert.Equal(t, "ref2/foo", visitors[0].Referrer)
	assert.Equal(t, "ref1/bar", visitors[1].Referrer)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 5, visitors[0].Sessions)
	assert.Equal(t, 1, visitors[1].Sessions)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.InDelta(t, 0.5714, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1428, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.4, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[1].BounceRate, 0.01)
}

func TestAnalyzer_ReferrerGrouping(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Referrer: "https://example.com", ReferrerName: "example.com", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Referrer: "https://www.example.com", ReferrerName: "example.com", PageViews: 1},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Referrer: "https://www.example.com", ReferrerName: "example.com", PageViews: 1},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), Referrer: "https://example.com", ReferrerName: "example.com", PageViews: 1},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Referrer(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "example.com", visitors[0].ReferrerName)
	assert.Equal(t, 4, visitors[0].Visitors)
	visitors, err = analyzer.Visitors.Referrer(&Filter{ReferrerName: []string{"example.com"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "example.com", visitors[0].ReferrerName)
	assert.Equal(t, "example.com", visitors[1].ReferrerName)
	assert.Equal(t, "https://example.com", visitors[0].Referrer)
	assert.Equal(t, "https://www.example.com", visitors[1].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
}

func TestAnalyzer_ReferrerUnknown(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), SessionID: 1, ExitPath: "/exit", PageViews: 3, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), SessionID: 1, ExitPath: "/exit", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), SessionID: 1, ExitPath: "/", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), SessionID: 1, ExitPath: "/", PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 3, Time: time.Now().Add(time.Minute), Start: time.Now(), SessionID: 3, ExitPath: "/bar", Referrer: "ref3", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), ExitPath: "/", Referrer: "ref1", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), ExitPath: "/", Referrer: "ref1", PageViews: 1, IsBounce: true},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Referrer(&Filter{Referrer: []string{pkg.Unknown}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Empty(t, visitors[0].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.4, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)
}

func TestAnalyzer_ReferrerTags(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", ReferrerName: "ref1", PageViews: 2},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", ReferrerName: "ref1", PageViews: 2},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 3), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", ReferrerName: "ref1", PageViews: 3},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(4), Start: time.Now(), EntryPath: "/", ExitPath: "/", ReferrerName: "ref2", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(4), Start: time.Now(), EntryPath: "/", ExitPath: "/", ReferrerName: "ref3", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(3).Add(time.Minute * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/", ReferrerName: "ref1", PageViews: 3},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(3).Add(time.Minute), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", ReferrerName: "ref2", PageViews: 2},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute), Path: "/", ReferrerName: "ref1", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 2), Path: "/foo", ReferrerName: "ref1"},
		{VisitorID: 1, Time: util.PastDay(4).Add(time.Minute * 3), Path: "/bar", ReferrerName: "ref1", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 2, Time: util.PastDay(4), Path: "/", ReferrerName: "ref2"},
		{VisitorID: 3, Time: util.PastDay(4), Path: "/", ReferrerName: "ref3"},
		{VisitorID: 4, Time: util.PastDay(3), Path: "/", ReferrerName: "ref1", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 4, Time: util.PastDay(3).Add(time.Minute), Path: "/foo", ReferrerName: "ref1", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 4, Time: util.PastDay(3).Add(time.Minute * 2), Path: "/", ReferrerName: "ref1"},
		{VisitorID: 5, Time: util.PastDay(3), Path: "/", ReferrerName: "ref1", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 5, Time: util.PastDay(3).Add(time.Minute), Path: "/bar", ReferrerName: "ref2", TagKeys: []string{"author"}, TagValues: []string{"John"}},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Referrer(&Filter{
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref2", visitors[1].ReferrerName)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[0].Sessions)
	assert.Equal(t, 1, visitors[1].Sessions)
	assert.InDelta(t, 0.4, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2, visitors[1].RelativeVisitors, 0.01)
	assert.Equal(t, 0, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.InDelta(t, 0, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		Tags: map[string]string{"author": "!John"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref2", visitors[1].ReferrerName)
	assert.Equal(t, "ref3", visitors[2].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 3, visitors[0].Sessions)
	assert.Equal(t, 1, visitors[1].Sessions)
	assert.Equal(t, 1, visitors[2].Sessions)
	assert.InDelta(t, 0.6, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2, visitors[2].RelativeVisitors, 0.01)
	assert.Equal(t, 0, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.InDelta(t, 0, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[2].BounceRate, 0.01)
	visitors, err = analyzer.Visitors.Referrer(&Filter{
		Tags: map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "ref1", visitors[0].ReferrerName)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[0].Sessions)
	assert.InDelta(t, 0.4, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 0, visitors[0].Bounces)
	assert.InDelta(t, 0, visitors[0].BounceRate, 0.01)
}

func TestAnalyzer_Channel(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/exit", Channel: "Channel 2", PageViews: 3, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/exit", Channel: "Channel 2", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), Start: time.Now(), ExitPath: "/", Channel: "Channel 2", PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(time.Minute), Start: time.Now(), ExitPath: "/bar", Channel: "Channel 3", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), ExitPath: "/", Channel: "Channel 1", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), ExitPath: "/", Channel: "Channel 1", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: time.Now(), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
		{VisitorID: 4, Time: time.Now(), Name: "event", MetaKeys: []string{"foo", "bar"}, MetaValues: []string{"val0", "val1"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.Channel(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Channel 1", visitors[0].Channel)
	assert.Equal(t, "Channel 2", visitors[1].Channel)
	assert.Equal(t, "Channel 3", visitors[2].Channel)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 2, visitors[0].Views)
	assert.Equal(t, 3, visitors[1].Views)
	assert.Equal(t, 2, visitors[2].Views)
	assert.Equal(t, 2, visitors[0].Sessions)
	assert.Equal(t, 1, visitors[1].Sessions)
	assert.Equal(t, 1, visitors[2].Sessions)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.25, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.25, visitors[2].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2857, visitors[0].RelativeViews, 0.01)
	assert.InDelta(t, 0.4285, visitors[1].RelativeViews, 0.01)
	assert.InDelta(t, 0.2857, visitors[2].RelativeViews, 0.01)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	_, err = analyzer.Visitors.Channel(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors.Channel(getMaxFilter("event"))
	assert.NoError(t, err)
	visitors, err = analyzer.Visitors.Channel(&Filter{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	_, err = analyzer.Visitors.Channel(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldChannel,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldChannel,
			Input: "cha",
		},
	}})
	assert.NoError(t, err)

	// filter for channel
	visitors, err = analyzer.Visitors.Channel(&Filter{Channel: []string{"Channel 1"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Channel 1", visitors[0].Channel)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[0].Views)
	assert.Equal(t, 2, visitors[0].Sessions)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2857, visitors[0].RelativeViews, 0.01)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)

	// filter for event
	visitors, err = analyzer.Visitors.Channel(&Filter{EventName: []string{"event"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "Channel 1", visitors[0].Channel)
	assert.Equal(t, "Channel 2", visitors[1].Channel)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[0].Views)
	assert.Equal(t, 3, visitors[1].Views)
	assert.Equal(t, 1, visitors[0].Sessions)
	assert.Equal(t, 1, visitors[1].Sessions)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.InDelta(t, 0.25, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.25, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1428, visitors[0].RelativeViews, 0.01)
	assert.InDelta(t, 0.4285, visitors[1].RelativeViews, 0.01)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
}

func TestAnalyzer_Timezone(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions([]model.Session{
		{Sign: 1, VisitorID: 1, Time: util.PastDay(3).Add(time.Hour * 18), Start: time.Now(), ExitPath: "/"}, // 18:00 UTC -> 03:00 Asia/Tokyo
		{Sign: 1, VisitorID: 2, Time: util.PastDay(2), Start: time.Now(), ExitPath: "/"},                     // 00:00 UTC -> 09:00 Asia/Tokyo
		{Sign: 1, VisitorID: 3, Time: util.PastDay(1).Add(time.Hour * 19), Start: time.Now(), ExitPath: "/"}, // 19:00 UTC -> 04:00 Asia/Tokyo
	}))
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(3).Add(time.Hour * 18), Path: "/"}, // 18:00 UTC -> 03:00 Asia/Tokyo
		{VisitorID: 2, Time: util.PastDay(2), Path: "/"},                     // 00:00 UTC -> 09:00 Asia/Tokyo
		{VisitorID: 3, Time: util.PastDay(1).Add(time.Hour * 19), Path: "/"}, // 19:00 UTC -> 04:00 Asia/Tokyo
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors.ByPeriod(&Filter{From: util.PastDay(3), To: util.PastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	hours, err := analyzer.Visitors.ByHour(&Filter{From: util.PastDay(3), To: util.PastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 1, hours[0].Visitors)
	assert.Equal(t, 1, hours[18].Visitors)
	assert.Equal(t, 1, hours[19].Visitors)
	timezone, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)
	visitors, err = analyzer.Visitors.ByPeriod(&Filter{From: util.PastDay(3), To: util.PastDay(1), Timezone: timezone})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, 0, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 0, visitors[2].Visitors)
	hours, err = analyzer.Visitors.ByHour(&Filter{From: util.PastDay(3), To: util.PastDay(1), Timezone: timezone})
	assert.NoError(t, err)
	assert.Equal(t, 1, hours[3].Visitors)
	assert.Equal(t, 0, hours[4].Visitors) // pushed to the next day, so outside of filter range
	assert.Equal(t, 1, hours[9].Visitors)
}

func TestAnalyzer_CalculateGrowth(t *testing.T) {
	growth := calculateGrowth(0, 0)
	assert.InDelta(t, 0, growth, 0.001)
	growth = calculateGrowth(1000, 0)
	assert.InDelta(t, 1, growth, 0.001)
	growth = calculateGrowth(0, 1000)
	assert.InDelta(t, -1, growth, 0.001)
	growth = calculateGrowth(100, 50)
	assert.InDelta(t, 1, growth, 0.001)
	growth = calculateGrowth(50, 100)
	assert.InDelta(t, -0.5, growth, 0.001)
	growth = calculateGrowth(0.0, 0.0)
	assert.InDelta(t, 0, growth, 0.001)
	growth = calculateGrowth(1000.0, 0.0)
	assert.InDelta(t, 1, growth, 0.001)
	growth = calculateGrowth(0.0, 1000.0)
	assert.InDelta(t, -1, growth, 0.001)
	growth = calculateGrowth(100.0, 50.0)
	assert.InDelta(t, 1, growth, 0.001)
	growth = calculateGrowth(50.0, 100.0)
	assert.InDelta(t, -0.5, growth, 0.001)
}

func TestAnalyzer_GetPreviousPeriod(t *testing.T) {
	analyzer := NewAnalyzer(dbClient)
	f := &Filter{
		From: util.PastDay(5),
		To:   util.Today(),
	}
	analyzer.Visitors.getPreviousPeriod(f)
	assert.Equal(t, util.PastDay(11), f.From)
	assert.Equal(t, util.PastDay(6), f.To)
	f = &Filter{
		From:          util.PastDay(5),
		To:            util.Today(),
		ImportedUntil: util.PastDay(6),
	}
	f.validate()
	analyzer.Visitors.getPreviousPeriod(f)
	assert.Equal(t, util.PastDay(6), f.From)
	assert.Equal(t, util.PastDay(6), f.To)
	assert.Equal(t, util.PastDay(11), f.importedFrom)
	assert.Equal(t, util.PastDay(7), f.importedTo)
	f = &Filter{
		From:          util.PastDay(10),
		To:            util.Today(),
		ImportedUntil: util.PastDay(5),
	}
	f.validate()
	analyzer.Visitors.getPreviousPeriod(f)
	assert.Equal(t, util.PastDay(21), f.From)
	assert.Equal(t, util.PastDay(11), f.To)
	assert.Equal(t, util.PastDay(21), f.importedFrom)
	assert.Equal(t, util.PastDay(11), f.importedTo)
	f = &Filter{
		From:          util.PastDay(10),
		To:            util.Today(),
		ImportedUntil: util.PastDay(9),
	}
	f.validate()
	analyzer.Visitors.getPreviousPeriod(f)
	assert.Equal(t, util.PastDay(21), f.From)
	assert.Equal(t, util.PastDay(11), f.To)
	assert.Equal(t, util.PastDay(21), f.importedFrom)
	assert.Equal(t, util.PastDay(11), f.importedTo)
}
