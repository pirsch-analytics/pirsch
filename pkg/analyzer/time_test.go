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

func TestAnalyzer_AvgSessionDuration(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), SessionID: 1, DurationSeconds: 25},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today(), Start: time.Now(), SessionID: 1, DurationSeconds: 25},
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), SessionID: 1, DurationSeconds: 28},
			{Sign: 1, VisitorID: 2, Time: util.Today(), Start: time.Now(), SessionID: 2, DurationSeconds: 5},
			{Sign: 1, VisitorID: 3, Time: util.Today(), Start: time.Now(), SessionID: 3, DurationSeconds: 0},
			{Sign: 1, VisitorID: 4, Time: util.Today(), Start: time.Now(), SessionID: 4, DurationSeconds: 35},
			{Sign: 1, VisitorID: 5, Time: util.Today(), Start: time.Now(), SessionID: 5, DurationSeconds: 2},
			{Sign: 1, VisitorID: 6, Time: util.Today(), Start: time.Now(), SessionID: 6, DurationSeconds: 0},
		},
	})
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, SessionID: 1, Time: util.Today(), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 25), Path: "/foo", DurationSeconds: 25, TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 28), Path: "/bar", DurationSeconds: 3, TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 2, SessionID: 2, Time: util.Today(), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, SessionID: 2, Time: util.Today().Add(time.Second * 5), Path: "/foo", DurationSeconds: 5, TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 3, SessionID: 3, Time: util.Today(), Path: "/"},
		{VisitorID: 4, SessionID: 4, Time: util.Today(), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 4, SessionID: 4, Time: util.Today().Add(time.Second * 35), Path: "/foo", DurationSeconds: 35},
		{VisitorID: 5, SessionID: 5, Time: util.Today(), Path: "/"},
		{VisitorID: 5, SessionID: 5, Time: util.Today().Add(time.Second * 2), Path: "/bar", DurationSeconds: 2, TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 6, SessionID: 6, Time: util.Today(), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.Time.AvgSessionDuration(nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, 18, stats[0].AverageTimeSpentSeconds) // (28+5+35+2)/4 because bounced visitors are not taken into the calculation
	stats, err = analyzer.Time.AvgSessionDuration(&Filter{
		Path: []string{"/foo"},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, 23, stats[0].AverageTimeSpentSeconds) // 28+5+35 because these sessions contain the path /foo
	stats, err = analyzer.Time.AvgSessionDuration(&Filter{
		PathPattern: []string{"(?i)^/.*$"},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, 18, stats[0].AverageTimeSpentSeconds) // all
	stats, err = analyzer.Time.AvgSessionDuration(&Filter{
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, 23, stats[0].AverageTimeSpentSeconds) // (28+5+35)/3
}

func TestAnalyzer_AvgSessionDurationPeriod(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, DurationSeconds: 25},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, DurationSeconds: 25},
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, DurationSeconds: 28},
			{Sign: 1, VisitorID: 2, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, DurationSeconds: 5},
			{Sign: 1, VisitorID: 3, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, DurationSeconds: 0},
			{Sign: 1, VisitorID: 4, Time: time.Date(2023, 9, 26, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 4, DurationSeconds: 35},
			{Sign: 1, VisitorID: 5, Time: time.Date(2023, 9, 26, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 5, DurationSeconds: 2},
			{Sign: 1, VisitorID: 6, Time: time.Date(2023, 9, 27, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 6, DurationSeconds: 10},
			{Sign: 1, VisitorID: 7, Time: time.Date(2023, 9, 27, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 7, DurationSeconds: 12},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.Time.AvgSessionDuration(&Filter{
		From: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 9, 28, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 4)
	assert.Equal(t, (28+5)/2, stats[0].AverageTimeSpentSeconds)
	assert.Equal(t, (35+2)/2, stats[1].AverageTimeSpentSeconds)
	assert.Equal(t, (10+12)/2, stats[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, stats[3].AverageTimeSpentSeconds)
	stats, err = analyzer.Time.AvgSessionDuration(&Filter{
		From:   time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2023, 9, 28, 0, 0, 0, 0, time.UTC),
		Period: pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, (28+5+35+2+10+12)/6, stats[0].AverageTimeSpentSeconds)
}

func TestAnalyzer_AvgSessionDurationTz(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 9, 24, 22, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, DurationSeconds: 21},
			{Sign: 1, VisitorID: 2, Time: time.Date(2023, 9, 24, 23, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, DurationSeconds: 18},
			{Sign: 1, VisitorID: 3, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, DurationSeconds: 25},
		},
		{
			{Sign: -1, VisitorID: 3, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, DurationSeconds: 25},
			{Sign: 1, VisitorID: 3, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, DurationSeconds: 28},
			{Sign: 1, VisitorID: 4, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 4, DurationSeconds: 5},
			{Sign: 1, VisitorID: 5, Time: time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 5, DurationSeconds: 0},
			{Sign: 1, VisitorID: 6, Time: time.Date(2023, 9, 26, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 6, DurationSeconds: 35},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	stats, err := analyzer.Time.AvgSessionDuration(&Filter{
		Timezone: tz,
		From:     time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2023, 9, 25, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, 18, stats[0].AverageTimeSpentSeconds)
}

func TestAnalyzer_AvgTimeOnPage(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 2, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 3, Time: time.Date(2023, 10, 23, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 4, Time: time.Date(2023, 10, 23, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 5, Time: time.Date(2023, 10, 24, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 6, Time: time.Date(2023, 10, 24, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
		},
	})
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC), SessionID: 3, Path: "/", Title: "Home", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC).Add(time.Second * 9), SessionID: 3, Path: "/foo", DurationSeconds: 9, Title: "Foo"},
		{VisitorID: 2, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC), SessionID: 3, Path: "/", Title: "Home", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, Time: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC).Add(time.Second * 7), SessionID: 3, Path: "/foo", DurationSeconds: 7, Title: "Foo"},
		{VisitorID: 3, Time: time.Date(2023, 10, 23, 0, 0, 0, 0, time.UTC), SessionID: 2, Path: "/", Title: "Home", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 3, Time: time.Date(2023, 10, 23, 0, 0, 0, 0, time.UTC).Add(time.Second * 5), SessionID: 2, Path: "/foo", DurationSeconds: 5, Title: "Foo"},
		{VisitorID: 4, Time: time.Date(2023, 10, 23, 0, 0, 0, 0, time.UTC), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: time.Date(2023, 10, 23, 0, 0, 0, 0, time.UTC).Add(time.Second * 4), SessionID: 2, Path: "/foo", DurationSeconds: 4, Title: "Foo", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: time.Date(2023, 10, 24, 0, 0, 0, 0, time.UTC), SessionID: 1, Path: "/", Title: "Home", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: time.Date(2023, 10, 24, 0, 0, 0, 0, time.UTC).Add(time.Second * 8), SessionID: 1, Path: "/foo", DurationSeconds: 8, Title: "Foo"},
		{VisitorID: 6, Time: time.Date(2023, 10, 24, 0, 0, 0, 0, time.UTC), SessionID: 1, Path: "/", Title: "Home", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 6, Time: time.Date(2023, 10, 24, 0, 0, 0, 0, time.UTC).Add(time.Second * 6), SessionID: 1, Path: "/foo", DurationSeconds: 6, Title: "Foo", TagKeys: []string{"author"}, TagValues: []string{"John"}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	byDay, err := analyzer.Time.AvgTimeOnPage(&Filter{
		From: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		Path: []string{"/"},
		From: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		Path: []string{"/foo"},
		From: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 0, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		MaxTimeOnPageSeconds: 5,
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 3)
	assert.Equal(t, 5, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 5, byDay[2].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		PathPattern: []string{"\\/.*"},
		From:        time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:          time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		EntryPath: []string{"/"},
		From:      time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		ExitPath: []string{"/foo"},
		From:     time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		From:   time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
		Period: pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 2)
	assert.Equal(t, 2023, byDay[0].Week.Time.Year())
	assert.Equal(t, time.Month(10), byDay[0].Week.Time.Month())
	assert.Equal(t, 16, byDay[0].Week.Time.Day())
	assert.Equal(t, 2023, byDay[1].Week.Time.Year())
	assert.Equal(t, time.Month(10), byDay[1].Week.Time.Month())
	assert.Equal(t, 23, byDay[1].Week.Time.Day())
	assert.Equal(t, 4, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 6, byDay[1].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		From:   time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
		Period: pkg.PeriodMonth,
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 1)
	assert.Equal(t, 2023, byDay[0].Month.Time.Year())
	assert.Equal(t, time.Month(10), byDay[0].Month.Time.Month())
	assert.Equal(t, 1, byDay[0].Month.Time.Day())
	assert.Equal(t, 5, byDay[0].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		From:   time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
		Period: pkg.PeriodYear,
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 1)
	assert.Equal(t, 2023, byDay[0].Year.Time.Year())
	assert.Equal(t, 5, byDay[0].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		From: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
		Tags: map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 8, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)

	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		From: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
		Tags: map[string]string{"author": "!John"},
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 0, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 6, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)

	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{
		From: time.Date(2023, 10, 22, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC),
		Tags: map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 0, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 5, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 6, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	_, err = analyzer.Time.AvgTimeOnPage(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgTimeOnPage(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_AvgTimeOnPagePeriod(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC).Add(time.Second * 9), SessionID: 3, Path: "/foo", DurationSeconds: 9, Title: "Foo"},
		{VisitorID: 2, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 2, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC).Add(time.Second * 7), SessionID: 3, Path: "/foo", DurationSeconds: 7, Title: "Foo"},
		{VisitorID: 3, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 3, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC).Add(time.Second * 5), SessionID: 2, Path: "/foo", DurationSeconds: 5, Title: "Foo"},
		{VisitorID: 4, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC).Add(time.Second * 4), SessionID: 2, Path: "/foo", DurationSeconds: 4, Title: "Foo"},
		{VisitorID: 5, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 5, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC).Add(time.Second * 8), SessionID: 1, Path: "/foo", DurationSeconds: 8, Title: "Foo"},
		{VisitorID: 6, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 6, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC).Add(time.Second * 6), SessionID: 1, Path: "/foo", DurationSeconds: 6, Title: "Foo"},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 2, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 3, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 4, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 5, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 6, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	byDay, err := analyzer.Time.AvgTimeOnPage(&Filter{
		From:   time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC),
		Path:   []string{"/"},
		Period: pkg.PeriodWeek,
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 1)
	assert.Equal(t, 6, byDay[0].AverageTimeSpentSeconds)
}

func TestAnalyzer_AvgTimeOnPageTz(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC).Add(time.Second * 9), SessionID: 3, Path: "/foo", DurationSeconds: 9, Title: "Foo"},
		{VisitorID: 2, Time: time.Date(2023, 10, 2, 23, 0, 0, 0, time.UTC), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 2, Time: time.Date(2023, 10, 2, 23, 0, 0, 0, time.UTC).Add(time.Second * 7), SessionID: 3, Path: "/foo", DurationSeconds: 7, Title: "Foo"},
		{VisitorID: 3, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 3, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC).Add(time.Second * 5), SessionID: 2, Path: "/foo", DurationSeconds: 5, Title: "Foo"},
		{VisitorID: 4, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC).Add(time.Second * 4), SessionID: 2, Path: "/foo", DurationSeconds: 4, Title: "Foo"},
		{VisitorID: 5, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 5, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC).Add(time.Second * 8), SessionID: 1, Path: "/foo", DurationSeconds: 8, Title: "Foo"},
		{VisitorID: 6, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 6, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC).Add(time.Second * 6), SessionID: 1, Path: "/foo", DurationSeconds: 6, Title: "Foo"},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 2, Time: time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 3, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 4, Time: time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 5, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 6, Time: time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	byDay, err := analyzer.Time.AvgTimeOnPage(&Filter{
		Timezone: tz,
		From:     time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC),
		Path:     []string{"/"},
	})
	assert.NoError(t, err)
	assert.Len(t, byDay, 1)
	assert.Equal(t, (7+5+4)/3, byDay[0].AverageTimeSpentSeconds)
}
