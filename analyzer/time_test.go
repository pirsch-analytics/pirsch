package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v5/db"
	"github.com/pirsch-analytics/pirsch/v5/model"
	"github.com/pirsch-analytics/pirsch/v5/util"
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
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.Time.AvgSessionDuration(nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 1)

	// average is (28+5+35+2)/4 because bounced visitors are not taken into the calculation
	assert.Equal(t, 18, stats[0].AverageTimeSpentSeconds)
}

func TestAnalyzer_AvgTimeOnPage(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.PastDay(3), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 1, Time: util.PastDay(3).Add(time.Second * 9), SessionID: 3, Path: "/foo", DurationSeconds: 9, Title: "Foo"},
		{VisitorID: 2, Time: util.PastDay(3), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 2, Time: util.PastDay(3).Add(time.Second * 7), SessionID: 3, Path: "/foo", DurationSeconds: 7, Title: "Foo"},
		{VisitorID: 3, Time: util.PastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 3, Time: util.PastDay(2).Add(time.Second * 5), SessionID: 2, Path: "/foo", DurationSeconds: 5, Title: "Foo"},
		{VisitorID: 4, Time: util.PastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: util.PastDay(2).Add(time.Second * 4), SessionID: 2, Path: "/foo", DurationSeconds: 4, Title: "Foo"},
		{VisitorID: 5, Time: util.PastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 5, Time: util.PastDay(1).Add(time.Second * 8), SessionID: 1, Path: "/foo", DurationSeconds: 8, Title: "Foo"},
		{VisitorID: 6, Time: util.PastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 6, Time: util.PastDay(1).Add(time.Second * 6), SessionID: 1, Path: "/foo", DurationSeconds: 6, Title: "Foo"},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.PastDay(3), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.PastDay(3), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, Time: util.PastDay(3), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 2, Time: util.PastDay(3), Start: time.Now(), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 3, Time: util.PastDay(2), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 4, Time: util.PastDay(2), Start: time.Now(), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 5, Time: util.PastDay(1), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 6, Time: util.PastDay(1), Start: time.Now(), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	byDay, err := analyzer.Time.AvgTimeOnPage(&Filter{Path: []string{"/"}, From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{Path: []string{"/foo"}, From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 0, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.Time.AvgTimeOnPage(&Filter{MaxTimeOnPageSeconds: 5})
	assert.NoError(t, err)
	assert.Len(t, byDay, 3)
	assert.Equal(t, 5, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 5, byDay[2].AverageTimeSpentSeconds)
	_, err = analyzer.Time.AvgTimeOnPage(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgTimeOnPage(getMaxFilter("event"))
	assert.NoError(t, err)
}
