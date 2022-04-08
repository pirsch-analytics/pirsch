package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_ActiveVisitors(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: time.Now().Add(-time.Minute * 30), Path: "/", Title: "Home"},
		{VisitorID: 1, Time: time.Now().Add(-time.Minute * 20), Path: "/", Title: "Home"},
		{VisitorID: 1, Time: time.Now().Add(-time.Minute * 15), Path: "/bar", Title: "Bar"},
		{VisitorID: 2, Time: time.Now().Add(-time.Minute * 4), Path: "/bar", Title: "Bar"},
		{VisitorID: 2, Time: time.Now().Add(-time.Minute * 3), Path: "/foo", Title: "Foo"},
		{VisitorID: 3, Time: time.Now().Add(-time.Minute * 3), Path: "/", Title: "Home"},
		{VisitorID: 4, Time: time.Now().Add(-time.Minute), Path: "/", Title: "Home"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(-time.Minute * 25)},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(-time.Minute * 25)},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(-time.Minute * 15)},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(-time.Minute * 3)},
			{Sign: 1, VisitorID: 3, Time: time.Now().Add(-time.Minute * 5)},
		},
		{
			{Sign: -1, VisitorID: 3, Time: time.Now().Add(-time.Minute * 5)},
			{Sign: 1, VisitorID: 3, Time: time.Now().Add(-time.Minute * 3)},
			{Sign: 1, VisitorID: 4, Time: time.Now().Add(-time.Minute)},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, count, err := analyzer.ActiveVisitors(nil, time.Minute*10)
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
	visitors, count, err = analyzer.ActiveVisitors(&Filter{Path: "/bar"}, time.Minute*30)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "/bar", visitors[0].Path)
	assert.Equal(t, 2, visitors[0].Visitors)
	_, _, err = analyzer.ActiveVisitors(getMaxFilter(""), time.Minute*10)
	assert.NoError(t, err)
	visitors, count, err = analyzer.ActiveVisitors(&Filter{IncludeTitle: true}, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Home", visitors[0].Title)
	assert.Equal(t, "Bar", visitors[1].Title)
	assert.Equal(t, "Foo", visitors[2].Title)
	_, _, err = analyzer.ActiveVisitors(getMaxFilter(""), time.Minute*10)
	assert.NoError(t, err)
}

func TestAnalyzer_TotalVisitors(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 10), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 10), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, ExitPath: "/", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 1, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: pastDay(4), SessionID: 4, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: pastDay(4).Add(time.Minute * 10), SessionID: 3, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 5, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 6, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: false, DurationSeconds: 600},
			{Sign: 1, VisitorID: 7, Time: pastDay(2), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: pastDay(2), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(4).Add(time.Minute * 10), SessionID: 4, Path: "/bar"},
		{VisitorID: 1, Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/"},
		{VisitorID: 1, Time: pastDay(4), Path: "/"},
		{VisitorID: 2, Time: pastDay(4), SessionID: 4, Path: "/"},
		{VisitorID: 2, Time: pastDay(4).Add(time.Minute * 10), SessionID: 3, Path: "/bar"},
		{VisitorID: 3, Time: pastDay(4), Path: "/"},
		{VisitorID: 4, Time: pastDay(4), Path: "/"},
		{VisitorID: 5, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 5, Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar"},
		{VisitorID: 6, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 7, Time: pastDay(2), Path: "/"},
		{VisitorID: 8, Time: pastDay(2), Path: "/"},
		{VisitorID: 9, Time: time.Now().UTC().Add(-time.Minute * 15), Path: "/"},
	}))
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.TotalVisitors(&Filter{From: pastDay(4), To: Today()})
	assert.NoError(t, err)
	assert.Equal(t, 9, visitors.Visitors)
	assert.Equal(t, 11, visitors.Sessions)
	assert.Equal(t, 13, visitors.Views)
	assert.Equal(t, 8, visitors.Bounces)
	assert.InDelta(t, 0.7272, visitors.BounceRate, 0.01)
	visitors, err = analyzer.TotalVisitors(&Filter{From: pastDay(2), To: Today()})
	assert.NoError(t, err)
	assert.Equal(t, 5, visitors.Visitors)
	assert.Equal(t, 5, visitors.Sessions)
	assert.Equal(t, 6, visitors.Views)
	assert.Equal(t, 3, visitors.Bounces)
	assert.InDelta(t, 0.6, visitors.BounceRate, 0.01)
	visitors, err = analyzer.TotalVisitors(&Filter{From: pastDay(1), To: Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	visitors, err = analyzer.TotalVisitors(&Filter{From: pastDay(1), To: Today()})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
	visitors, err = analyzer.TotalVisitors(&Filter{From: time.Now().UTC().Add(-time.Minute * 15), To: Today(), IncludeTime: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, visitors.Visitors)
	assert.Equal(t, 1, visitors.Sessions)
	assert.Equal(t, 1, visitors.Views)
	assert.Equal(t, 1, visitors.Bounces)
	assert.InDelta(t, 1, visitors.BounceRate, 0.01)
}

func TestAnalyzer_VisitorsAndAvgSessionDuration(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 10), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 10), SessionID: 4, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, ExitPath: "/", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 1, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: pastDay(4), SessionID: 4, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: pastDay(4).Add(time.Minute * 10), SessionID: 3, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 5, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, ExitPath: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
			{Sign: 1, VisitorID: 6, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, ExitPath: "/bar", PageViews: 1, IsBounce: false, DurationSeconds: 600},
			{Sign: 1, VisitorID: 7, Time: pastDay(2), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: pastDay(2), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 9, Time: Today(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(4).Add(time.Minute * 10), SessionID: 4, Path: "/bar"},
		{VisitorID: 1, Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/"},
		{VisitorID: 1, Time: pastDay(4), Path: "/"},
		{VisitorID: 2, Time: pastDay(4), SessionID: 4, Path: "/"},
		{VisitorID: 2, Time: pastDay(4).Add(time.Minute * 10), SessionID: 3, Path: "/bar"},
		{VisitorID: 3, Time: pastDay(4), Path: "/"},
		{VisitorID: 4, Time: pastDay(4), Path: "/"},
		{VisitorID: 5, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 5, Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar"},
		{VisitorID: 6, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar"},
		{VisitorID: 7, Time: pastDay(2), Path: "/"},
		{VisitorID: 8, Time: pastDay(2), Path: "/"},
		{VisitorID: 9, Time: Today(), Path: "/"},
	}))
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Visitors(&Filter{From: pastDay(4), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, pastDay(4), visitors[0].Day.Time)
	assert.Equal(t, pastDay(3), visitors[1].Day.Time)
	assert.Equal(t, pastDay(2), visitors[2].Day.Time)
	assert.Equal(t, pastDay(1), visitors[3].Day.Time)
	assert.Equal(t, Today(), visitors[4].Day.Time)
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
	assert.Equal(t, 5, visitors[2].Views)
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
	visitors, err = analyzer.Visitors(&Filter{Path: "/", From: pastDay(4), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.Equal(t, 0, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 5, visitors[0].Sessions)
	assert.Equal(t, 0, visitors[1].Sessions)
	assert.Equal(t, 2, visitors[2].Sessions)
	assert.Equal(t, 0, visitors[3].Sessions)
	assert.Equal(t, 1, visitors[4].Sessions)
	assert.Equal(t, 5, visitors[0].Views)
	assert.Equal(t, 0, visitors[1].Views)
	assert.Equal(t, 2, visitors[2].Views)
	assert.Equal(t, 0, visitors[3].Views)
	assert.Equal(t, 1, visitors[4].Views)
	assert.Equal(t, 4, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 2, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 0.8, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	asd, err := analyzer.AvgSessionDuration(nil)
	assert.NoError(t, err)
	assert.Len(t, asd, 2)
	assert.Equal(t, pastDay(4), asd[0].Day.Time)
	assert.Equal(t, pastDay(2), asd[1].Day.Time)
	assert.Equal(t, 300, asd[0].AverageTimeSpentSeconds)
	assert.Equal(t, 450, asd[1].AverageTimeSpentSeconds)
	tsd, err := analyzer.totalSessionDuration(&Filter{})
	assert.NoError(t, err)
	assert.Equal(t, 1200, tsd)
	visitors, err = analyzer.Visitors(&Filter{From: pastDay(4), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, pastDay(4), visitors[0].Day.Time)
	assert.Equal(t, pastDay(2), visitors[2].Day.Time)
	asd, err = analyzer.AvgSessionDuration(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, asd, 3)
	tsd, err = analyzer.totalSessionDuration(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 900, tsd)
	_, err = analyzer.Visitors(&Filter{
		From:   pastDay(90),
		To:     Today(),
		Period: PeriodWeek,
	})
	assert.NoError(t, err)
	_, err = analyzer.Visitors(&Filter{
		From:   Today(),
		To:     Today(),
		Period: PeriodWeek,
	})
	assert.NoError(t, err)
	_, err = analyzer.Visitors(&Filter{
		From:   pastDay(1),
		To:     Today(),
		Period: PeriodWeek,
	})
	assert.NoError(t, err)
	_, err = analyzer.Visitors(&Filter{
		From:        pastDay(90),
		To:          Today(),
		PathPattern: "(?i)^/bar",
	})
	assert.NoError(t, err)
	_, err = analyzer.Visitors(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Visitors(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.AvgSessionDuration(&Filter{
		From:   pastDay(90),
		To:     Today(),
		Period: PeriodWeek,
	})
	assert.NoError(t, err)
	_, err = analyzer.AvgSessionDuration(&Filter{
		From:        pastDay(90),
		To:          Today(),
		PathPattern: "(?i)^/bar",
	})
	assert.NoError(t, err)
	_, err = analyzer.AvgSessionDuration(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.AvgSessionDuration(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.totalSessionDuration(getMaxFilter(""))
	assert.NoError(t, err)
}

// FIXME test data or checks need to be changed so that they always work, and not rely on specific dates
/*func TestAnalyzer_VisitorsAvgSessionDurationAvgTimeOnPagePeriod(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveSessions([]Session{
		{Sign: 1, VisitorID: 1, Time: pastDay(391), PageViews: 1, DurationSeconds: 42},
		{Sign: 1, VisitorID: 2, Time: pastDay(14), PageViews: 1, DurationSeconds: 12},
		{Sign: 1, VisitorID: 3, Time: pastDay(7), PageViews: 1, DurationSeconds: 29},
		{Sign: 1, VisitorID: 4, Time: Today(), PageViews: 1, DurationSeconds: 34},
	}))
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(391), Path: "/", DurationSeconds: 5},
		{VisitorID: 2, Time: pastDay(14), Path: "/", DurationSeconds: 8},
		{VisitorID: 3, Time: pastDay(7), Path: "/", DurationSeconds: 7},
		{VisitorID: 4, Time: Today(), Path: "/", DurationSeconds: 3},
	}))
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Visitors(&Filter{From: pastDay(21), To: Today(), Period: PeriodWeek})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, pastWeek(4), getWeek(visitors[0].Week.Time))
	assert.Equal(t, pastWeek(3), getWeek(visitors[1].Week.Time))
	assert.Equal(t, pastWeek(2), getWeek(visitors[2].Week.Time))
	assert.Equal(t, pastWeek(1), getWeek(visitors[3].Week.Time))
	assert.Equal(t, 0, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	asd, err := analyzer.AvgSessionDuration(&Filter{From: pastDay(21), To: Today(), Period: PeriodWeek})
	assert.NoError(t, err)
	assert.Len(t, asd, 4)
	assert.Equal(t, pastWeek(4), getWeek(asd[0].Week.Time))
	assert.Equal(t, pastWeek(3), getWeek(asd[1].Week.Time))
	assert.Equal(t, pastWeek(2), getWeek(asd[2].Week.Time))
	assert.Equal(t, pastWeek(1), getWeek(asd[3].Week.Time))
	assert.Equal(t, 0, asd[0].AverageTimeSpentSeconds)
	assert.Equal(t, 2, asd[1].AverageTimeSpentSeconds)
	assert.Equal(t, 4, asd[2].AverageTimeSpentSeconds)
	assert.Equal(t, 11, asd[3].AverageTimeSpentSeconds)
	atop, err := analyzer.AvgTimeOnPage(&Filter{From: pastDay(21), To: Today(), Period: PeriodWeek})
	assert.NoError(t, err)
	assert.Len(t, atop, 4)
	assert.Equal(t, pastWeek(4), getWeek(atop[0].Week.Time))
	assert.Equal(t, pastWeek(3), getWeek(atop[1].Week.Time))
	assert.Equal(t, pastWeek(2), getWeek(atop[2].Week.Time))
	assert.Equal(t, pastWeek(1), getWeek(atop[3].Week.Time))
	assert.Equal(t, 0, atop[0].AverageTimeSpentSeconds)
	assert.Equal(t, 1, atop[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, atop[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, atop[3].AverageTimeSpentSeconds)
	visitors, err = analyzer.Visitors(&Filter{From: pastDay(85), To: Today(), Period: PeriodMonth})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, pastMonth(3), visitors[0].Month.Time)
	assert.Equal(t, pastMonth(2), visitors[1].Month.Time)
	assert.Equal(t, pastMonth(1), visitors[2].Month.Time)
	assert.Equal(t, pastMonth(0), visitors[3].Month.Time)
	assert.Equal(t, 0, visitors[0].Visitors)
	assert.Equal(t, 0, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	asd, err = analyzer.AvgSessionDuration(&Filter{From: pastDay(85), To: Today(), Period: PeriodMonth})
	assert.NoError(t, err)
	assert.Len(t, asd, 4)
	assert.Equal(t, pastMonth(3), asd[0].Month.Time)
	assert.Equal(t, pastMonth(2), asd[1].Month.Time)
	assert.Equal(t, pastMonth(1), asd[2].Month.Time)
	assert.Equal(t, pastMonth(0), asd[3].Month.Time)
	assert.Equal(t, 0, asd[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, asd[1].AverageTimeSpentSeconds)
	assert.Equal(t, 1, asd[2].AverageTimeSpentSeconds)
	assert.Equal(t, 7, asd[3].AverageTimeSpentSeconds)
	atop, err = analyzer.AvgTimeOnPage(&Filter{From: pastDay(85), To: Today(), Period: PeriodMonth})
	assert.NoError(t, err)
	assert.Len(t, atop, 4)
	assert.Equal(t, pastMonth(3), atop[0].Month.Time)
	assert.Equal(t, pastMonth(2), atop[1].Month.Time)
	assert.Equal(t, pastMonth(1), atop[2].Month.Time)
	assert.Equal(t, pastMonth(0), atop[3].Month.Time)
	assert.Equal(t, 0, atop[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, atop[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, atop[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, atop[3].AverageTimeSpentSeconds)
	year := time.Now().Year()
	visitors, err = analyzer.Visitors(&Filter{From: pastDay(400), To: Today(), Period: PeriodYear})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, year-1, visitors[0].Year.Time.Year())
	assert.Equal(t, year, visitors[1].Year.Time.Year())
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 3, visitors[1].Visitors)
	asd, err = analyzer.AvgSessionDuration(&Filter{From: pastDay(400), To: Today(), Period: PeriodYear})
	assert.NoError(t, err)
	assert.Len(t, asd, 2)
	assert.Equal(t, year-1, asd[0].Year.Time.Year())
	assert.Equal(t, year, asd[1].Year.Time.Year())
	assert.Equal(t, 0, asd[0].AverageTimeSpentSeconds)
	assert.Equal(t, 1, asd[1].AverageTimeSpentSeconds)
	asd, err = analyzer.AvgTimeOnPage(&Filter{From: pastDay(400), To: Today(), Period: PeriodYear})
	assert.NoError(t, err)
	assert.Len(t, asd, 2)
	assert.Equal(t, year-1, asd[0].Year.Time.Year())
	assert.Equal(t, year, asd[1].Year.Time.Year())
	assert.Equal(t, 0, asd[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, asd[1].AverageTimeSpentSeconds)
}*/

func TestAnalyzer_Growth(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 15), SessionID: 4, ExitPath: "/bar", DurationSeconds: 600, PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: pastDay(4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: pastDay(3).Add(time.Minute * 10), SessionID: 3, ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 4, Time: pastDay(3).Add(time.Minute * 10), SessionID: 3, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: pastDay(3).Add(time.Minute * 5), SessionID: 3, ExitPath: "/foo", DurationSeconds: 300, PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: pastDay(3), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: pastDay(3), SessionID: 3, ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: pastDay(3).Add(time.Minute * 10), SessionID: 31, ExitPath: "/bar", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: pastDay(3), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 7, Time: pastDay(3), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 8, Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, ExitPath: "/bar", DurationSeconds: 300, PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 9, Time: pastDay(2), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 10, Time: pastDay(2), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 11, Time: Today(), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	growth, err := analyzer.Growth(nil)
	assert.ErrorIs(t, err, ErrNoPeriodOrDay)
	assert.Nil(t, growth)
	growth, err = analyzer.Growth(&Filter{From: pastDay(2), To: pastDay(2)})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, -0.25, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, -0.4285, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, -0.5, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, -0.2, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	growth, err = analyzer.Growth(&Filter{From: pastDay(3), To: pastDay(2)})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1.3333, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1.2, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 2, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0.1666, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	_, err = analyzer.Growth(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Growth(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_GrowthDay(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveSessions([]Session{
		{Sign: 1, VisitorID: 1, Time: pastDay(2).Add(time.Hour * 5)},
		{Sign: 1, VisitorID: 2, Time: pastDay(1).Add(time.Hour * 3)},
		{Sign: 1, VisitorID: 3, Time: pastDay(1).Add(time.Hour * 4)},
		{Sign: 1, VisitorID: 4, Time: pastDay(1).Add(time.Hour * 9)},
		{Sign: 1, VisitorID: 5, Time: Today().Add(time.Hour * 4)},
		{Sign: 1, VisitorID: 6, Time: Today().Add(time.Hour * 9)},
		{Sign: 1, VisitorID: 7, Time: Today().Add(time.Hour * 12)},
		{Sign: 1, VisitorID: 8, Time: Today().Add(time.Hour * 17)},
		{Sign: 1, VisitorID: 9, Time: Today().Add(time.Hour * 21)},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)

	// Testing for today is hard because it would require messing with the time.Now function.
	// I don't want to do that, so let's assume it works (tested in debug mode) and just get no error for today.
	growth, err := analyzer.Growth(&Filter{From: Today(), To: Today()})
	assert.NoError(t, err)
	assert.NotNil(t, growth)

	growth, err = analyzer.Growth(&Filter{From: pastDay(1), To: pastDay(1)})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 2, growth.VisitorsGrowth, 0.001)
}

func TestAnalyzer_GrowthDayFirstHour(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveSessions([]Session{
		{Sign: 1, VisitorID: 1, Time: pastDay(1)},
		{Sign: 1, VisitorID: 2, Time: pastDay(1).Add(time.Hour * 4)},
		{Sign: 1, VisitorID: 3, Time: Today()},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	growth, err := analyzer.Growth(&Filter{From: Today(), To: Today().Add(time.Hour * 4), IncludeTime: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, -0.5, growth.VisitorsGrowth, 0.01)
	growth, err = analyzer.Growth(&Filter{From: Today(), To: Today().Add(time.Hour * 2), IncludeTime: true})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0, growth.VisitorsGrowth, 0.01)
}

func TestAnalyzer_GrowthNoData(t *testing.T) {
	cleanupDB()
	analyzer := NewAnalyzer(dbClient, nil)
	growth, err := analyzer.Growth(&Filter{From: pastDay(7), To: pastDay(7)})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 0, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 0, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 0, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	_, err = analyzer.Growth(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Growth(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_GrowthEvents(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: pastDay(4).Add(-time.Second), EntryPath: "/", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 4, Time: pastDay(4).Add(-time.Second), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: pastDay(4).Add(time.Minute * 5), EntryPath: "/", ExitPath: "/foo"},
			{Sign: -1, VisitorID: 1, SessionID: 4, Time: pastDay(4).Add(time.Minute * 5), EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 1, SessionID: 4, Time: pastDay(4).Add(time.Minute * 15), EntryPath: "/", ExitPath: "/bar"},
			{Sign: 1, VisitorID: 2, Time: pastDay(4).Add(time.Second * 2), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 3, Time: pastDay(4).Add(time.Second * 3), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 4, SessionID: 3, Time: pastDay(3).Add(time.Second * 3), EntryPath: "/", ExitPath: "/"},
			{Sign: -1, VisitorID: 4, SessionID: 3, Time: pastDay(3).Add(time.Second * 3), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 4, SessionID: 3, Time: pastDay(3).Add(time.Minute * 5), EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 4, Time: pastDay(3).Add(time.Second * 5), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 5, SessionID: 3, Time: pastDay(3).Add(time.Second * 6), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 5, SessionID: 31, Time: pastDay(3).Add(time.Minute * 10), EntryPath: "/bar", ExitPath: "/bar"},
			{Sign: 1, VisitorID: 6, Time: pastDay(3).Add(time.Second * 7), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 7, Time: pastDay(3).Add(time.Second * 8), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 8, SessionID: 2, Time: pastDay(2).Add(time.Second * 9), EntryPath: "/", ExitPath: "/"},
			{Sign: -1, VisitorID: 8, SessionID: 2, Time: pastDay(2).Add(time.Second * 9), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 8, SessionID: 2, Time: pastDay(2).Add(time.Minute * 5), EntryPath: "/", ExitPath: "/bar"},
			{Sign: 1, VisitorID: 9, Time: pastDay(2).Add(time.Second * 10), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 10, Time: pastDay(2).Add(time.Second * 11), EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 11, Time: Today().Add(time.Second * 12), EntryPath: "/", ExitPath: "/"},
		},
	})
	assert.NoError(t, dbClient.SaveEvents([]Event{
		{Name: "event1", VisitorID: 1, Time: pastDay(4).Add(time.Second), SessionID: 4, Path: "/"},
		{Name: "event1", DurationSeconds: 300, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/foo"},
		{Name: "event1", DurationSeconds: 600, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 15), SessionID: 4, Path: "/bar"},
		{Name: "event1", VisitorID: 2, Time: pastDay(4).Add(time.Second * 2), Path: "/"},
		{Name: "event1", VisitorID: 3, Time: pastDay(4).Add(time.Second * 3), Path: "/"},
		{Name: "event1", VisitorID: 4, Time: pastDay(3).Add(time.Second * 4), SessionID: 3, Path: "/"},
		{Name: "event1", DurationSeconds: 300, VisitorID: 4, Time: pastDay(3).Add(time.Minute * 5), SessionID: 3, Path: "/foo"},
		{Name: "event1", VisitorID: 4, Time: pastDay(3).Add(time.Second * 5), Path: "/"},
		{Name: "event1", VisitorID: 5, Time: pastDay(3).Add(time.Second * 6), SessionID: 3, Path: "/"},
		{Name: "event1", VisitorID: 5, Time: pastDay(3).Add(time.Minute * 10), SessionID: 31, Path: "/bar"},
		{Name: "event1", VisitorID: 6, Time: pastDay(3).Add(time.Second * 7), Path: "/"},
		{Name: "event1", VisitorID: 7, Time: pastDay(3).Add(time.Second * 8), Path: "/"},
		{Name: "event1", VisitorID: 8, Time: pastDay(2).Add(time.Second * 9), SessionID: 2, Path: "/"},
		{Name: "event1", DurationSeconds: 300, VisitorID: 8, Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar"},
		{Name: "event1", VisitorID: 9, Time: pastDay(2).Add(time.Second * 10), Path: "/"},
		{Name: "event1", VisitorID: 10, Time: pastDay(2).Add(time.Second * 11), Path: "/"},
		{Name: "event1", VisitorID: 11, Time: Today().Add(time.Second * 12), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	growth, err := analyzer.Growth(nil)
	assert.ErrorIs(t, err, ErrNoPeriodOrDay)
	assert.Nil(t, growth)
	growth, err = analyzer.Growth(&Filter{From: pastDay(2), To: pastDay(2), EventName: "event1"})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, -0.25, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, -0.4285, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, -0.5, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	analyzer = NewAnalyzer(dbClient, &AnalyzerConfig{
		DisableBotFilter: true,
	})
	growth, err = analyzer.Growth(&Filter{From: pastDay(3), To: pastDay(2), EventName: "event1"})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1.3333, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1.2, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 2, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, -0.3333, growth.TimeSpentGrowth, 0.001)
	growth, err = analyzer.Growth(&Filter{From: pastDay(3), To: pastDay(2), EventName: "event1", Path: "/bar"})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 1, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, -0.5, growth.TimeSpentGrowth, 0.001)
	_, err = analyzer.Growth(getMaxFilter("event1"))
	assert.NoError(t, err)
}

func TestAnalyzer_VisitorHours(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(2).Add(time.Hour * 3), ExitPath: "/foo", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: pastDay(2).Add(time.Hour * 3), ExitPath: "/foo", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: pastDay(2).Add(time.Hour * 3), ExitPath: "/", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: pastDay(2).Add(time.Hour * 8), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 3, Time: pastDay(1).Add(time.Hour * 4), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: pastDay(1).Add(time.Hour * 5), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: pastDay(1).Add(time.Hour * 8), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 6, Time: Today().Add(time.Hour * 5), ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 7, Time: Today().Add(time.Hour * 10), ExitPath: "/", PageViews: 1, IsBounce: true},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(2).Add(time.Hour*2 + time.Minute*30), Path: "/foo"},
		{VisitorID: 1, Time: pastDay(2).Add(time.Hour * 3), Path: "/"},
		{VisitorID: 2, Time: pastDay(2).Add(time.Hour * 8), Path: "/"},
		{VisitorID: 3, Time: pastDay(1).Add(time.Hour * 4), Path: "/"},
		{VisitorID: 4, Time: pastDay(1).Add(time.Hour * 5), Path: "/"},
		{VisitorID: 5, Time: pastDay(1).Add(time.Hour * 8), Path: "/"},
		{VisitorID: 6, Time: Today().Add(time.Hour * 5), Path: "/"},
		{VisitorID: 7, Time: Today().Add(time.Hour * 10), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.VisitorHours(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 2, visitors[5].Visitors)
	assert.Equal(t, 2, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)

	assert.Equal(t, 2, visitors[3].Views)
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

	visitors, err = analyzer.VisitorHours(&Filter{From: pastDay(1), To: Today()})
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

	_, err = analyzer.VisitorHours(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.VisitorHours(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_PagesAndAvgTimeOnPage(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home"},
		{VisitorID: 1, Time: pastDay(4).Add(time.Minute * 3), SessionID: 4, DurationSeconds: 180, Path: "/foo", Title: "Foo"},
		{VisitorID: 1, Time: pastDay(4).Add(time.Hour), SessionID: 41, Path: "/", Title: "Home"},
		{VisitorID: 2, Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home"},
		{VisitorID: 2, Time: pastDay(4).Add(time.Minute * 2), SessionID: 4, DurationSeconds: 120, Path: "/bar", Title: "Bar"},
		{VisitorID: 3, Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home"},
		{VisitorID: 5, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 5, Time: pastDay(2).Add(time.Minute * 5), SessionID: 21, Path: "/bar", Title: "Bar"},
		{VisitorID: 6, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 6, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, DurationSeconds: 600, Path: "/bar", Title: "Bar"},
		{VisitorID: 6, Time: pastDay(2).Add(time.Minute * 11), SessionID: 21, Path: "/bar", Title: "Bar"},
		{VisitorID: 6, Time: pastDay(2).Add(time.Minute * 21), SessionID: 21, DurationSeconds: 600, Path: "/foo", Title: "Foo"},
		{VisitorID: 7, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 8, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 9, Time: Today(), SessionID: 2, Path: "/", Title: "Home"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(4).Add(time.Minute * 3), SessionID: 4, DurationSeconds: 180, ExitPath: "/foo", EntryTitle: "Foo", IsBounce: false, PageViews: 2},
			{Sign: 1, VisitorID: 1, Time: pastDay(4).Add(time.Hour), SessionID: 41, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: pastDay(4).Add(time.Minute * 2), SessionID: 4, DurationSeconds: 120, ExitPath: "/bar", EntryTitle: "Bar", IsBounce: false, PageViews: 2},
			{Sign: 1, VisitorID: 3, Time: pastDay(4), SessionID: 4, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 4, Time: pastDay(4), SessionID: 4, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 4, Time: pastDay(4), SessionID: 4, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 4, Time: pastDay(4), SessionID: 4, ExitPath: "/foo", EntryTitle: "Foo", IsBounce: false, PageViews: 2},
			{Sign: 1, VisitorID: 5, Time: pastDay(2), SessionID: 2, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 5, Time: pastDay(2).Add(time.Minute * 5), SessionID: 21, ExitPath: "/bar", EntryTitle: "Bar", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 6, Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, DurationSeconds: 600, ExitPath: "/bar", EntryTitle: "Bar", IsBounce: false, PageViews: 2},
			{Sign: 1, VisitorID: 6, Time: pastDay(2).Add(time.Minute * 21), SessionID: 21, DurationSeconds: 600, ExitPath: "/foo", EntryTitle: "Foo", IsBounce: false, PageViews: 2},
			{Sign: 1, VisitorID: 7, Time: pastDay(2), SessionID: 2, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 8, Time: pastDay(2), SessionID: 2, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 9, Time: Today(), SessionID: 2, ExitPath: "/", EntryTitle: "Home", IsBounce: true, PageViews: 1},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Pages(&Filter{IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path)
	assert.Equal(t, "/bar", visitors[1].Path)
	assert.Equal(t, "/foo", visitors[2].Path)
	assert.Empty(t, visitors[0].Title)
	assert.Empty(t, visitors[1].Title)
	assert.Empty(t, visitors[2].Title)
	assert.Equal(t, 9, visitors[0].Visitors)
	assert.Equal(t, 3, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.InDelta(t, 1, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2222, visitors[2].RelativeVisitors, 0.01)
	assert.Equal(t, 10, visitors[0].Sessions)
	assert.Equal(t, 4, visitors[1].Sessions)
	assert.Equal(t, 2, visitors[2].Sessions)
	assert.Equal(t, 10, visitors[0].Views)
	assert.Equal(t, 4, visitors[1].Views)
	assert.Equal(t, 2, visitors[2].Views)
	assert.InDelta(t, 0.5882, visitors[0].RelativeViews, 0.01)
	assert.InDelta(t, 0.2352, visitors[1].RelativeViews, 0.01)
	assert.InDelta(t, 0.125, visitors[2].RelativeViews, 0.01)
	assert.Equal(t, 6, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.InDelta(t, 0.6, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.25, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	assert.Equal(t, 300, visitors[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, visitors[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, visitors[2].AverageTimeSpentSeconds)
	top, err := analyzer.AvgTimeOnPage(nil)
	assert.NoError(t, err)
	assert.Len(t, top, 2)
	assert.Equal(t, pastDay(4), top[0].Day.Time)
	assert.Equal(t, pastDay(2), top[1].Day.Time)
	assert.Equal(t, 150, top[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, top[1].AverageTimeSpentSeconds)
	ttop, err := analyzer.totalTimeOnPage(&Filter{})
	assert.NoError(t, err)
	assert.Equal(t, 1500, ttop)
	visitors, err = analyzer.Pages(&Filter{From: pastDay(3), To: pastDay(1), IncludeTitle: true, IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path)
	assert.Equal(t, "/bar", visitors[1].Path)
	assert.Equal(t, "/foo", visitors[2].Path)
	assert.Equal(t, "Home", visitors[0].Title)
	assert.Equal(t, "Bar", visitors[1].Title)
	assert.Equal(t, "Foo", visitors[2].Title)
	assert.Equal(t, 600, visitors[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, visitors[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, visitors[2].AverageTimeSpentSeconds)
	top, err = analyzer.AvgTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1), IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, top, 3)
	assert.Equal(t, pastDay(3), top[0].Day.Time)
	assert.Equal(t, pastDay(2), top[1].Day.Time)
	assert.Equal(t, pastDay(1), top[2].Day.Time)
	assert.Equal(t, 0, top[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, top[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, top[2].AverageTimeSpentSeconds)
	ttop, err = analyzer.totalTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 1200, ttop)
	_, err = analyzer.Pages(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Pages(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.totalTimeOnPage(getMaxFilter(""))
	assert.NoError(t, err)
	visitors, err = analyzer.Pages(&Filter{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	ttop, err = analyzer.totalTimeOnPage(&Filter{MaxTimeOnPageSeconds: 200})
	assert.NoError(t, err)
	assert.Equal(t, 180+120+200+200, ttop)
}

func TestAnalyzer_PageTitle(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		// these need to be at the same day, because otherwise they will be in different partitions
		// and the neighbor function doesn't work for the time on page calculation (visitor ID 2 is unrelated, so next day is fine)
		{VisitorID: 1, Time: pastDay(1).Add(time.Hour), SessionID: 1, Path: "/", Title: "Home 1"},
		{VisitorID: 1, Time: pastDay(1).Add(time.Hour * 2), SessionID: 1, Path: "/", Title: "Home 2", DurationSeconds: 42},
		{VisitorID: 2, Time: Today(), SessionID: 3, Path: "/foo", Title: "Foo"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(2), SessionID: 1, ExitPath: "/foo", EntryTitle: "Foo"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: pastDay(2), SessionID: 1, ExitPath: "/foo", EntryTitle: "Foo"},
			{Sign: 1, VisitorID: 1, Time: pastDay(2), SessionID: 1, ExitPath: "/", EntryTitle: "Home 1"},
			{Sign: 1, VisitorID: 1, Time: pastDay(1), SessionID: 2, ExitPath: "/", EntryTitle: "Home 2", DurationSeconds: 42},
			{Sign: 1, VisitorID: 2, Time: Today(), SessionID: 3, ExitPath: "/foo", EntryTitle: "Foo"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Pages(&Filter{IncludeTitle: true, IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Home 1", visitors[0].Title)
	assert.Equal(t, "Home 2", visitors[1].Title)
	assert.Equal(t, "Foo", visitors[2].Title)
	assert.Equal(t, 42, visitors[0].AverageTimeSpentSeconds)
	assert.Equal(t, 42, visitors[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, visitors[2].AverageTimeSpentSeconds)
}

func TestAnalyzer_PageTitleEvent(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: pastDay(2), EntryPath: "/", ExitPath: "/", EntryTitle: "Home 1", ExitTitle: "Home 1"},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 1, Time: pastDay(2), EntryPath: "/", ExitPath: "/", EntryTitle: "Home 1", ExitTitle: "Home 1"},
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: pastDay(1), EntryPath: "/", ExitPath: "/", EntryTitle: "Home 1", ExitTitle: "Home 2"},
			{Sign: 1, VisitorID: 2, SessionID: 3, Time: pastDay(1), EntryPath: "/foo", ExitPath: "/foo", EntryTitle: "Foo", ExitTitle: "Foo"},
		},
	})
	assert.NoError(t, dbClient.SaveEvents([]Event{
		{Name: "event", VisitorID: 1, Time: pastDay(2), SessionID: 1, Path: "/", Title: "Home 1"},
		{Name: "event", VisitorID: 1, Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home 2", DurationSeconds: 42},
		{Name: "event", VisitorID: 2, Time: Today(), SessionID: 3, Path: "/foo", Title: "Foo"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Pages(&Filter{EventName: "event", IncludeTitle: true, IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Home 1", visitors[0].Title)
	assert.Equal(t, "Home 2", visitors[1].Title)
	assert.Equal(t, "Foo", visitors[2].Title)
	assert.Equal(t, 0, visitors[0].AverageTimeSpentSeconds)
	assert.Equal(t, 42, visitors[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, visitors[2].AverageTimeSpentSeconds)
}

func TestAnalyzer_EntryExitPages(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(2), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 1, Time: pastDay(2).Add(time.Second), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 1, Time: pastDay(2).Add(time.Second * 10), SessionID: 2, DurationSeconds: 10, Path: "/foo", Title: "Foo"},
		{VisitorID: 2, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 3, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: pastDay(1).Add(time.Second * 20), SessionID: 1, DurationSeconds: 20, Path: "/bar", Title: "Bar"},
		{VisitorID: 5, Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 5, Time: pastDay(1).Add(time.Second * 40), SessionID: 1, DurationSeconds: 40, Path: "/bar", Title: "Bar"},
		{VisitorID: 6, Time: pastDay(1), SessionID: 1, Path: "/bar", Title: "Bar"},
		{VisitorID: 7, Time: pastDay(1), SessionID: 1, Path: "/bar", Title: "Bar"},
		{VisitorID: 7, Time: pastDay(1).Add(time.Minute), SessionID: 2, Path: "/", Title: "Home"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(2).Add(time.Second * 10), SessionID: 1, DurationSeconds: 10, EntryPath: "/bar", ExitPath: "/foo", EntryTitle: "Bar", ExitTitle: "Foo"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: pastDay(2).Add(time.Second * 10), SessionID: 1, DurationSeconds: 10, EntryPath: "/bar", ExitPath: "/foo", EntryTitle: "Bar", ExitTitle: "Foo"},
			{Sign: 1, VisitorID: 1, Time: pastDay(2).Add(time.Second * 10), SessionID: 1, DurationSeconds: 10, EntryPath: "/", ExitPath: "/", EntryTitle: "Home", ExitTitle: "Home"},
			{Sign: 1, VisitorID: 1, Time: pastDay(2).Add(time.Second * 10), SessionID: 2, DurationSeconds: 10, EntryPath: "/", ExitPath: "/foo", EntryTitle: "Home", ExitTitle: "Foo"},
			{Sign: 1, VisitorID: 2, Time: pastDay(2), SessionID: 2, EntryPath: "/", ExitPath: "/", EntryTitle: "Home", ExitTitle: "Home"},
			{Sign: 1, VisitorID: 3, Time: pastDay(2), SessionID: 2, EntryPath: "/", ExitPath: "/", EntryTitle: "Home", ExitTitle: "Home"},
			{Sign: 1, VisitorID: 4, Time: pastDay(1).Add(time.Second * 20), SessionID: 1, DurationSeconds: 20, EntryPath: "/", ExitPath: "/bar", EntryTitle: "Home", ExitTitle: "Bar"},
			{Sign: 1, VisitorID: 5, Time: pastDay(1).Add(time.Second * 40), SessionID: 1, DurationSeconds: 40, EntryPath: "/", ExitPath: "/bar", EntryTitle: "Home", ExitTitle: "Bar"},
			{Sign: 1, VisitorID: 6, Time: pastDay(1), SessionID: 1, EntryPath: "/bar", ExitPath: "/bar", EntryTitle: "Bar", ExitTitle: "Bar"},
			{Sign: 1, VisitorID: 7, Time: pastDay(1).Add(time.Minute), SessionID: 1, EntryPath: "/bar", ExitPath: "/bar", EntryTitle: "Bar", ExitTitle: "Bar"},
			{Sign: 1, VisitorID: 7, Time: pastDay(1).Add(time.Minute), SessionID: 2, EntryPath: "/", ExitPath: "/", EntryTitle: "Home", ExitTitle: "Home"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	entries, err := analyzer.EntryPages(&Filter{IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "/", entries[0].Path)
	assert.Equal(t, "/bar", entries[1].Path)
	assert.Empty(t, entries[0].Title)
	assert.Empty(t, entries[1].Title)
	assert.Equal(t, 6, entries[0].Visitors)
	assert.Equal(t, 4, entries[1].Visitors)
	assert.Equal(t, 7, entries[0].Sessions)
	assert.Equal(t, 4, entries[1].Sessions)
	assert.Equal(t, 7, entries[0].Entries)
	assert.Equal(t, 2, entries[1].Entries)
	assert.InDelta(t, 1, entries[0].EntryRate, 0.001)
	assert.InDelta(t, 0.5, entries[1].EntryRate, 0.001)
	assert.Equal(t, 23, entries[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, entries[1].AverageTimeSpentSeconds)
	entries, err = analyzer.EntryPages(&Filter{From: pastDay(1), To: Today(), IncludeTitle: true, IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "/", entries[0].Path)
	assert.Equal(t, "/bar", entries[1].Path)
	assert.Equal(t, "Home", entries[0].Title)
	assert.Equal(t, "Bar", entries[1].Title)
	assert.Equal(t, 3, entries[0].Visitors)
	assert.Equal(t, 4, entries[1].Visitors)
	assert.Equal(t, 3, entries[0].Sessions)
	assert.Equal(t, 4, entries[1].Sessions)
	assert.Equal(t, 3, entries[0].Entries)
	assert.Equal(t, 2, entries[1].Entries)
	assert.InDelta(t, 1, entries[0].EntryRate, 0.001)
	assert.InDelta(t, 0.5, entries[1].EntryRate, 0.001)
	assert.Equal(t, 30, entries[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, entries[1].AverageTimeSpentSeconds)
	entries, err = analyzer.EntryPages(&Filter{From: pastDay(1), To: Today(), EntryPath: "/", IncludeTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "/", entries[0].Path)
	assert.Equal(t, 3, entries[0].Visitors)
	assert.Equal(t, 3, entries[0].Entries)
	assert.InDelta(t, 1, entries[0].EntryRate, 0.001)
	assert.Equal(t, 30, entries[0].AverageTimeSpentSeconds)
	_, err = analyzer.EntryPages(getMaxFilter(""))
	assert.NoError(t, err)
	exits, err := analyzer.ExitPages(nil)
	assert.NoError(t, err)
	assert.Len(t, exits, 3)
	assert.Equal(t, "/", exits[0].Path)
	assert.Equal(t, "/bar", exits[1].Path)
	assert.Equal(t, "/foo", exits[2].Path)
	assert.Empty(t, exits[0].Title)
	assert.Empty(t, exits[1].Title)
	assert.Empty(t, exits[2].Title)
	assert.Equal(t, 6, exits[0].Visitors)
	assert.Equal(t, 4, exits[1].Visitors)
	assert.Equal(t, 1, exits[2].Visitors)
	assert.Equal(t, 7, exits[0].Sessions)
	assert.Equal(t, 4, exits[1].Sessions)
	assert.Equal(t, 1, exits[2].Sessions)
	assert.Equal(t, 4, exits[0].Exits)
	assert.Equal(t, 4, exits[1].Exits)
	assert.Equal(t, 1, exits[2].Exits)
	assert.InDelta(t, 0.5714, exits[0].ExitRate, 0.001)
	assert.InDelta(t, 1, exits[1].ExitRate, 0.001)
	assert.InDelta(t, 1, exits[2].ExitRate, 0.001)
	exits, err = analyzer.ExitPages(&Filter{From: pastDay(1), To: Today(), IncludeTitle: true})
	assert.NoError(t, err)
	assert.Len(t, exits, 2)
	assert.Equal(t, "/bar", exits[0].Path)
	assert.Equal(t, "/", exits[1].Path)
	assert.Equal(t, "Bar", exits[0].Title)
	assert.Equal(t, "Home", exits[1].Title)
	assert.Equal(t, 4, exits[0].Visitors)
	assert.Equal(t, 3, exits[1].Visitors)
	assert.Equal(t, 4, exits[0].Exits)
	assert.Equal(t, 1, exits[1].Exits)
	assert.InDelta(t, 1, exits[0].ExitRate, 0.001)
	assert.InDelta(t, 0.33, exits[1].ExitRate, 0.01)
	exits, err = analyzer.ExitPages(&Filter{From: pastDay(1), To: Today(), ExitPath: "/"})
	assert.NoError(t, err)
	assert.Len(t, exits, 1)
	assert.Equal(t, "/", exits[0].Path)
	assert.Equal(t, 3, exits[0].Visitors)
	assert.Equal(t, 1, exits[0].Exits)
	assert.InDelta(t, 0.3333, exits[0].ExitRate, 0.01)
	_, err = analyzer.ExitPages(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.ExitPages(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_EntryExitPagesEvents(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: Today(), SessionID: 1, EntryPath: "/", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today(), SessionID: 1, EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, Time: Today().Add(time.Second), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
			{Sign: -1, VisitorID: 1, Time: Today().Add(time.Second), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 1, Time: Today().Add(time.Second * 2), SessionID: 1, EntryPath: "/", ExitPath: "/bar"},
		},
	})
	assert.NoError(t, dbClient.SaveEvents([]Event{
		{Name: "event", VisitorID: 1, SessionID: 1, Time: Today(), Path: "/foo"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	entries, err := analyzer.EntryPages(&Filter{EventName: "event"})
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "/", entries[0].Path)
	assert.Equal(t, 1, entries[0].Entries)
	exits, err := analyzer.ExitPages(&Filter{EventName: "event"})
	assert.NoError(t, err)
	assert.Len(t, exits, 1)
	assert.Equal(t, "/bar", exits[0].Path)
	assert.Equal(t, 1, exits[0].Exits)
	_, err = analyzer.EntryPages(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.ExitPages(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_PageConversions(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: Today(), Path: "/"},
		{VisitorID: 2, Time: Today(), Path: "/simple/page"},
		{VisitorID: 2, Time: Today().Add(time.Minute), Path: "/simple/page"},
		{VisitorID: 3, Time: Today(), Path: "/siMple/page/"},
		{VisitorID: 3, Time: Today().Add(time.Minute), Path: "/siMple/page/"},
		{VisitorID: 4, Time: Today(), Path: "/simple/page/with/many/slashes"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: Today(), ExitPath: "/foo", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today(), ExitPath: "/foo", PageViews: 1},
			{Sign: 1, VisitorID: 1, Time: Today(), ExitPath: "/", PageViews: 2},
			{Sign: 1, VisitorID: 2, Time: Today().Add(time.Minute), ExitPath: "/simple/page", PageViews: 1},
			{Sign: 1, VisitorID: 3, Time: Today(), ExitPath: "/siMple/page/", PageViews: 1},
			{Sign: 1, VisitorID: 3, Time: Today().Add(time.Minute), ExitPath: "/siMple/page/", PageViews: 2},
			{Sign: 1, VisitorID: 4, Time: Today(), ExitPath: "/simple/page/with/many/slashes", PageViews: 1},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	stats, err := analyzer.PageConversions(nil)
	assert.NoError(t, err)
	assert.Nil(t, stats)
	stats, err = analyzer.PageConversions(&Filter{PathPattern: ".*"})
	assert.NoError(t, err)
	assert.Equal(t, 4, stats.Visitors)
	assert.Equal(t, 6, stats.Views)
	assert.InDelta(t, 1, stats.CR, 0.01)
	stats, err = analyzer.PageConversions(&Filter{PathPattern: "(?i)^/simple/[^/]+/.*"})
	assert.NoError(t, err)
	assert.Equal(t, 2, stats.Visitors)
	assert.Equal(t, 3, stats.Views)
	assert.InDelta(t, 0.5, stats.CR, 0.01)
	_, err = analyzer.PageConversions(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.PageConversions(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_Events(t *testing.T) {
	cleanupDB()

	// create sessions for the conversion rate
	for i := 0; i < 10; i++ {
		saveSessions(t, [][]Session{
			{
				{Sign: 1, VisitorID: uint64(i), Time: Today(), EntryPath: "/", ExitPath: "/exit"},
			},
			{
				{Sign: -1, VisitorID: uint64(i), Time: Today(), EntryPath: "/", ExitPath: "/exit"},
				{Sign: 1, VisitorID: uint64(i), Time: Today(), EntryPath: "/", ExitPath: "/exit"},
			},
		})
	}

	assert.NoError(t, dbClient.SaveEvents([]Event{
		{Name: "event1", DurationSeconds: 5, MetaKeys: []string{"status", "price"}, MetaValues: []string{"in", "34.56"}, VisitorID: 1, Time: Today(), Path: "/"},
		{Name: "event1", DurationSeconds: 8, MetaKeys: []string{"status", "price"}, MetaValues: []string{"out", "34.56"}, VisitorID: 2, Time: Today().Add(time.Second), Path: "/simple/page"},
		{Name: "event1", DurationSeconds: 3, VisitorID: 3, Time: Today().Add(time.Second * 2), Path: "/simple/page/1"},
		{Name: "event1", DurationSeconds: 8, VisitorID: 3, Time: Today().Add(time.Minute), Path: "/simple/page/2"},
		{Name: "event1", DurationSeconds: 2, MetaKeys: []string{"status"}, MetaValues: []string{"in"}, VisitorID: 4, Time: Today().Add(time.Second * 3), Path: "/"},
		{Name: "event2", DurationSeconds: 1, VisitorID: 1, Time: Today().Add(time.Second * 4), Path: "/"},
		{Name: "event2", DurationSeconds: 5, VisitorID: 2, Time: Today().Add(time.Second * 5), Path: "/"},
		{Name: "event2", DurationSeconds: 7, MetaKeys: []string{"status", "price"}, MetaValues: []string{"in", "34.56"}, VisitorID: 2, Time: Today().Add(time.Minute), Path: "/simple/page"},
		{Name: "event2", DurationSeconds: 9, MetaKeys: []string{"status", "price", "third"}, MetaValues: []string{"in", "13.74", "param"}, VisitorID: 3, Time: Today().Add(time.Second * 6), Path: "/simple/page"},
		{Name: "event2", DurationSeconds: 3, MetaKeys: []string{"price"}, MetaValues: []string{"34.56"}, VisitorID: 4, Time: Today().Add(time.Second * 7), Path: "/"},
		{Name: "event2", DurationSeconds: 4, VisitorID: 5, Time: Today().Add(time.Second * 8), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	stats, err := analyzer.Events(nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.Equal(t, 5, stats[0].Visitors)
	assert.Equal(t, 4, stats[1].Visitors)
	assert.Equal(t, 6, stats[0].Views)
	assert.Equal(t, 5, stats[1].Views)
	assert.InDelta(t, 0.5, stats[0].CR, 0.001)
	assert.InDelta(t, 0.4, stats[1].CR, 0.001)
	assert.InDelta(t, 4, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 5, stats[1].AverageDurationSeconds, 0.001)
	assert.Len(t, stats[0].MetaKeys, 3)
	assert.Len(t, stats[1].MetaKeys, 2)
	stats, err = analyzer.Events(&Filter{EntryPath: "/exit"})
	assert.NoError(t, err)
	assert.Len(t, stats, 0)
	stats, err = analyzer.Events(&Filter{EntryPath: "/", ExitPath: "/exit"})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.Equal(t, 5, stats[0].Visitors)
	assert.Equal(t, 4, stats[1].Visitors)
	assert.Equal(t, 6, stats[0].Views)
	assert.Equal(t, 5, stats[1].Views)
	assert.InDelta(t, 0.5, stats[0].CR, 0.001)
	assert.InDelta(t, 0.4, stats[1].CR, 0.001)
	assert.InDelta(t, 4, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 5, stats[1].AverageDurationSeconds, 0.001)
	assert.Len(t, stats[0].MetaKeys, 3)
	assert.Len(t, stats[1].MetaKeys, 2)
	stats, err = analyzer.Events(&Filter{EventName: "event2"})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 5, stats[0].Visitors)
	assert.Equal(t, 6, stats[0].Views)
	assert.InDelta(t, 0.5, stats[0].CR, 0.001)
	assert.InDelta(t, 4, stats[0].AverageDurationSeconds, 0.001)
	assert.Len(t, stats[0].MetaKeys, 3)
	stats, err = analyzer.Events(&Filter{EventName: "does-not-exist"})
	assert.NoError(t, err)
	assert.Empty(t, stats)
	_, err = analyzer.Events(getMaxFilter(""))
	assert.NoError(t, err)
	stats, err = analyzer.EventBreakdown(&Filter{EventName: "event1", EventMetaKey: "status"})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 2, stats[0].Views)
	assert.Equal(t, 1, stats[1].Views)
	assert.InDelta(t, 0.2, stats[0].CR, 0.001)
	assert.InDelta(t, 0.1, stats[1].CR, 0.001)
	assert.InDelta(t, 3, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 8, stats[1].AverageDurationSeconds, 0.001)
	assert.Equal(t, "in", stats[0].MetaValue)
	assert.Equal(t, "out", stats[1].MetaValue)
	stats, err = analyzer.EventBreakdown(&Filter{EventName: "event2", EventMetaKey: "status"})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 2, stats[0].Views)
	assert.InDelta(t, 0.2, stats[0].CR, 0.001)
	assert.InDelta(t, 8, stats[0].AverageDurationSeconds, 0.001)
	assert.Equal(t, "in", stats[0].MetaValue)
	stats, err = analyzer.EventBreakdown(&Filter{EventName: "event2", EventMetaKey: "price"})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 2, stats[0].Views)
	assert.Equal(t, 1, stats[1].Views)
	assert.InDelta(t, 0.2, stats[0].CR, 0.001)
	assert.InDelta(t, 0.1, stats[1].CR, 0.001)
	assert.InDelta(t, 5, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 9, stats[1].AverageDurationSeconds, 0.001)
	assert.Equal(t, "34.56", stats[0].MetaValue)
	assert.Equal(t, "13.74", stats[1].MetaValue)
	stats, err = analyzer.EventBreakdown(&Filter{EventName: "event2", EventMetaKey: "third"})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 1, stats[0].Views)
	assert.InDelta(t, 0.1, stats[0].CR, 0.001)
	assert.InDelta(t, 9, stats[0].AverageDurationSeconds, 0.001)
	assert.Equal(t, "param", stats[0].MetaValue)
	stats, err = analyzer.EventBreakdown(&Filter{EventName: "does-not-exist", EventMetaKey: "status"})
	assert.NoError(t, err)
	assert.Empty(t, stats)
	stats, err = analyzer.EventBreakdown(&Filter{EventName: "event1", EventMetaKey: "does-not-exist"})
	assert.NoError(t, err)
	assert.Empty(t, stats)
	_, err = analyzer.EventBreakdown(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.EventBreakdown(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_EventList(t *testing.T) {
	cleanupDB()

	// create sessions for the conversion rate
	for i := 0; i < 5; i++ {
		saveSessions(t, [][]Session{
			{
				{Sign: 1, VisitorID: uint64(i + 1), Time: Today(), EntryPath: "/", ExitPath: "/exit"},
			},
			{
				{Sign: -1, VisitorID: uint64(i + 1), Time: Today(), EntryPath: "/", ExitPath: "/exit"},
				{Sign: 1, VisitorID: uint64(i + 1), Time: Today(), EntryPath: "/", ExitPath: "/exit"},
			},
		})
	}

	assert.NoError(t, dbClient.SaveEvents([]Event{
		{Name: "event1", MetaKeys: []string{"a", "b"}, MetaValues: []string{"foo", "42"}, VisitorID: 1, Time: Today(), Path: "/"},
		{Name: "event1", MetaKeys: []string{"a", "b"}, MetaValues: []string{"foo", "42"}, VisitorID: 2, Time: Today(), Path: "/foo"},
		{Name: "event1", MetaKeys: []string{"a", "b"}, MetaValues: []string{"bar", "42"}, VisitorID: 1, Time: Today(), Path: "/bar"},
		{Name: "event2", MetaKeys: []string{"a", "b"}, MetaValues: []string{"foo", "42"}, VisitorID: 3, Time: Today(), Path: "/"},
		{Name: "event2", MetaKeys: []string{"a", "b"}, MetaValues: []string{"foo", "56"}, VisitorID: 4, Time: Today(), Path: "/"},
		{Name: "event2", MetaKeys: []string{"a", "b"}, MetaValues: []string{"foo", "42"}, VisitorID: 5, Time: Today(), Path: "/foo"},
	}))
	analyzer := NewAnalyzer(dbClient, nil)
	stats, err := analyzer.EventList(nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 4)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	assert.Equal(t, "event1", stats[2].Name)
	assert.Equal(t, "event2", stats[3].Name)
	assert.Equal(t, 2, stats[0].Count)
	assert.Equal(t, 2, stats[1].Count)
	assert.Equal(t, 1, stats[2].Count)
	assert.Equal(t, 1, stats[3].Count)
	assert.Equal(t, "foo", stats[0].Meta["a"])
	assert.Equal(t, "42", stats[0].Meta["b"])
	assert.Equal(t, "foo", stats[1].Meta["a"])
	assert.Equal(t, "42", stats[1].Meta["b"])
	assert.Equal(t, "bar", stats[2].Meta["a"])
	assert.Equal(t, "42", stats[2].Meta["b"])
	assert.Equal(t, "foo", stats[3].Meta["a"])
	assert.Equal(t, "56", stats[3].Meta["b"])
	stats, err = analyzer.EventList(&Filter{EventName: "event1", Path: "/foo"})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, "foo", stats[0].Meta["a"])
	assert.Equal(t, "42", stats[0].Meta["b"])
	stats, err = analyzer.EventList(&Filter{Path: "/foo"})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	stats, err = analyzer.EventList(&Filter{EventMeta: map[string]string{"a": "bar"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, "bar", stats[0].Meta["a"])
	stats, err = analyzer.EventList(&Filter{EventMeta: map[string]string{"a": "foo", "b": "56"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, "foo", stats[0].Meta["a"])
	assert.Equal(t, "56", stats[0].Meta["b"])
	stats, err = analyzer.EventList(&Filter{EventMeta: map[string]string{"a": "no", "b": "result"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 0)
}

func TestAnalyzer_Referrer(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), ExitPath: "/exit", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), ExitPath: "/exit", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), ExitPath: "/", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(time.Minute), ExitPath: "/bar", Referrer: "ref3/foo", ReferrerName: "Ref3", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 3, Time: time.Now(), ExitPath: "/", Referrer: "ref1/foo", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 4, Time: time.Now(), ExitPath: "/", Referrer: "ref1/bar", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Referrer(nil)
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
	_, err = analyzer.Referrer(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Referrer(getMaxFilter("event"))
	assert.NoError(t, err)
	visitors, err = analyzer.Referrer(&Filter{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)

	// filter for referrer name
	visitors, err = analyzer.Referrer(&Filter{ReferrerName: "Ref1"})
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
	visitors, err = analyzer.Referrer(&Filter{Referrer: "ref1/foo"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref1/foo", visitors[0].Referrer)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)

	// filter for referrer name and full referrer
	visitors, err = analyzer.Referrer(&Filter{ReferrerName: "Ref1", Referrer: "ref1/foo"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref1/foo", visitors[0].Referrer)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
}

func TestAnalyzer_ReferrerUnknown(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), SessionID: 1, ExitPath: "/exit", PageViews: 3, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), SessionID: 1, ExitPath: "/exit", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 1, Time: time.Now().Add(time.Minute * 2), SessionID: 1, ExitPath: "/", PageViews: 3, IsBounce: true},
			{Sign: 1, VisitorID: 2, Time: time.Now().Add(time.Minute * 2), SessionID: 1, ExitPath: "/", PageViews: 3, IsBounce: false},
			{Sign: 1, VisitorID: 3, Time: time.Now().Add(time.Minute), SessionID: 3, ExitPath: "/bar", Referrer: "ref3", PageViews: 2, IsBounce: false},
			{Sign: 1, VisitorID: 4, Time: time.Now(), ExitPath: "/", Referrer: "ref1", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 5, Time: time.Now(), ExitPath: "/", Referrer: "ref1", PageViews: 1, IsBounce: true},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Referrer(&Filter{Referrer: Unknown})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Empty(t, visitors[0].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.4, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)
}

func TestAnalyzer_Platform(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: time.Now(), Path: "/"},
		{VisitorID: 1, Time: time.Now(), Path: "/foo"},
		{VisitorID: 1, Time: time.Now(), Path: "/bar"},
		{VisitorID: 2, Time: time.Now(), Path: "/"},
		{VisitorID: 3, Time: time.Now(), Path: "/"},
		{VisitorID: 4, Time: time.Now(), Path: "/"},
		{VisitorID: 5, Time: time.Now(), Path: "/"},
		{VisitorID: 6, Time: time.Now(), Path: "/"},
	}))
	saveSessions(t, [][]Session{
		{
			// set mobile which we overwrite with desktop to be sure the results get collapsed
			{Sign: 1, VisitorID: 1, Time: time.Now(), Mobile: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Mobile: true},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Desktop: true},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Mobile: true},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Mobile: true},
			{Sign: 1, VisitorID: 4, Time: time.Now()},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Desktop: true},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Desktop: true},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	platform, err := analyzer.Platform(&Filter{From: pastDay(5), To: Today()})
	assert.NoError(t, err)
	assert.Equal(t, 3, platform.PlatformDesktop)
	assert.Equal(t, 2, platform.PlatformMobile)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0.5, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0.1666, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Platform(&Filter{Path: "/foo"})
	assert.NoError(t, err)
	assert.Equal(t, 1, platform.PlatformDesktop)
	assert.Equal(t, 0, platform.PlatformMobile)
	assert.Equal(t, 0, platform.PlatformUnknown)
	assert.InDelta(t, 1, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformUnknown, 0.01)
	_, err = analyzer.Platform(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Platform(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_Languages(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Language: "ru"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Language: "ru"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Language: "en"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Language: "de"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Language: "de"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Language: "jp"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Language: "en"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Language: "en"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Languages(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].Language)
	assert.Equal(t, "de", visitors[1].Language)
	assert.Equal(t, "jp", visitors[2].Language)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Languages(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Languages(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_Countries(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), CountryCode: "ru"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), CountryCode: "ru"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), CountryCode: "en"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), CountryCode: "de"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), CountryCode: "de"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), CountryCode: "jp"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), CountryCode: "en"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), CountryCode: "en"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Countries(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].CountryCode)
	assert.Equal(t, "de", visitors[1].CountryCode)
	assert.Equal(t, "jp", visitors[2].CountryCode)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Countries(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Countries(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_Cities(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), CountryCode: "no", City: "Oslo"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), CountryCode: "no", City: "Oslo"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), CountryCode: "gb", City: "London"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), CountryCode: "de", City: "Berlin"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), CountryCode: "de", City: "Berlin"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), CountryCode: "jp", City: "Tokyo"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), CountryCode: "gb", City: "London"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), CountryCode: "gb", City: "London"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Cities(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "gb", visitors[0].CountryCode)
	assert.Equal(t, "de", visitors[1].CountryCode)
	assert.Equal(t, "jp", visitors[2].CountryCode)
	assert.Equal(t, "London", visitors[0].City)
	assert.Equal(t, "Berlin", visitors[1].City)
	assert.Equal(t, "Tokyo", visitors[2].City)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Cities(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Cities(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_Browser(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Browser: BrowserEdge},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Browser: BrowserEdge},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Browser: BrowserChrome},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Browser: BrowserFirefox},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Browser: BrowserFirefox},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Browser: BrowserSafari},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Browser: BrowserChrome},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Browser: BrowserChrome},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Browser(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, BrowserChrome, visitors[0].Browser)
	assert.Equal(t, BrowserFirefox, visitors[1].Browser)
	assert.Equal(t, BrowserSafari, visitors[2].Browser)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Browser(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Browser(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_BrowserVersion(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Browser: BrowserEdge, BrowserVersion: "85.0"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Browser: BrowserEdge, BrowserVersion: "85.0"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Browser: BrowserFirefox, BrowserVersion: "89.0.0"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Browser: BrowserFirefox, BrowserVersion: "89.0.1"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Browser: BrowserSafari, BrowserVersion: "14.1.2"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "87.2"},
			{Sign: 1, VisitorID: 7, Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "86.0"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.BrowserVersion(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, BrowserChrome, visitors[0].Browser)
	assert.Equal(t, BrowserChrome, visitors[1].Browser)
	assert.Equal(t, BrowserChrome, visitors[2].Browser)
	assert.Equal(t, BrowserFirefox, visitors[3].Browser)
	assert.Equal(t, BrowserFirefox, visitors[4].Browser)
	assert.Equal(t, BrowserSafari, visitors[5].Browser)
	assert.Equal(t, "85.1", visitors[0].BrowserVersion)
	assert.Equal(t, "86.0", visitors[1].BrowserVersion)
	assert.Equal(t, "87.2", visitors[2].BrowserVersion)
	assert.Equal(t, "89.0.0", visitors[3].BrowserVersion)
	assert.Equal(t, "89.0.1", visitors[4].BrowserVersion)
	assert.Equal(t, "14.1.2", visitors[5].BrowserVersion)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.InDelta(t, 0.2857, visitors[0].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[1].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[2].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[3].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[4].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[5].RelativeVisitors, 0.001)
	_, err = analyzer.BrowserVersion(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.BrowserVersion(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_OS(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), OS: OSLinux},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), OS: OSLinux},
			{Sign: 1, VisitorID: 1, Time: time.Now(), OS: OSWindows},
			{Sign: 1, VisitorID: 2, Time: time.Now(), OS: OSMac},
			{Sign: 1, VisitorID: 3, Time: time.Now(), OS: OSMac},
			{Sign: 1, VisitorID: 4, Time: time.Now(), OS: OSAndroid},
			{Sign: 1, VisitorID: 5, Time: time.Now(), OS: OSWindows},
			{Sign: 1, VisitorID: 6, Time: time.Now(), OS: OSWindows},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.OS(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, OSWindows, visitors[0].OS)
	assert.Equal(t, OSMac, visitors[1].OS)
	assert.Equal(t, OSAndroid, visitors[2].OS)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.OS(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.OS(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_OSVersion(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), OS: OSLinux, OSVersion: "1"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), OS: OSLinux, OSVersion: "1"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), OS: OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), OS: OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), OS: OSMac, OSVersion: "14.0.0"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), OS: OSMac, OSVersion: "13.1.0"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), OS: OSLinux},
			{Sign: 1, VisitorID: 6, Time: time.Now(), OS: OSWindows, OSVersion: "9"},
			{Sign: 1, VisitorID: 7, Time: time.Now(), OS: OSWindows, OSVersion: "8"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.OSVersion(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, OSWindows, visitors[0].OS)
	assert.Equal(t, OSLinux, visitors[1].OS)
	assert.Equal(t, OSMac, visitors[2].OS)
	assert.Equal(t, OSMac, visitors[3].OS)
	assert.Equal(t, OSWindows, visitors[4].OS)
	assert.Equal(t, OSWindows, visitors[5].OS)
	assert.Equal(t, "10", visitors[0].OSVersion)
	assert.Empty(t, visitors[1].OSVersion)
	assert.Equal(t, "13.1.0", visitors[2].OSVersion)
	assert.Equal(t, "14.0.0", visitors[3].OSVersion)
	assert.Equal(t, "8", visitors[4].OSVersion)
	assert.Equal(t, "9", visitors[5].OSVersion)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.InDelta(t, 0.2857, visitors[0].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[1].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[2].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[3].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[4].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[5].RelativeVisitors, 0.001)
	_, err = analyzer.OSVersion(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.OSVersion(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_ScreenClass(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), ScreenClass: "S", ScreenWidth: 415, ScreenHeight: 600},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), ScreenClass: "S", ScreenWidth: 415, ScreenHeight: 600},
			{Sign: 1, VisitorID: 1, Time: time.Now(), ScreenClass: "XXL", ScreenWidth: 3840, ScreenHeight: 2080},
			{Sign: 1, VisitorID: 2, Time: time.Now(), ScreenClass: "XL", ScreenWidth: 2560, ScreenHeight: 1440},
			{Sign: 1, VisitorID: 3, Time: time.Now(), ScreenClass: "XL", ScreenWidth: 2560, ScreenHeight: 1440},
			{Sign: 1, VisitorID: 4, Time: time.Now(), ScreenClass: "L", ScreenWidth: 1980, ScreenHeight: 1080},
			{Sign: 1, VisitorID: 5, Time: time.Now(), ScreenClass: "XXL", ScreenWidth: 3840, ScreenHeight: 2080},
			{Sign: 1, VisitorID: 6, Time: time.Now(), ScreenClass: "XXL", ScreenWidth: 3840, ScreenHeight: 2080},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.ScreenClass(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "XXL", visitors[0].ScreenClass)
	assert.Equal(t, "XL", visitors[1].ScreenClass)
	assert.Equal(t, "L", visitors[2].ScreenClass)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	visitors, err = analyzer.ScreenClass(&Filter{ScreenWidth: "2560"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "XL", visitors[0].ScreenClass)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
	visitors, err = analyzer.ScreenClass(&Filter{ScreenHeight: "1080"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "L", visitors[0].ScreenClass)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.InDelta(t, 0.1666, visitors[0].RelativeVisitors, 0.01)
	_, err = analyzer.ScreenClass(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.ScreenClass(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_UTM(t *testing.T) {
	cleanupDB()
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), UTMSource: "sourceX", UTMMedium: "mediumX", UTMCampaign: "campaignX", UTMContent: "contentX", UTMTerm: "termX"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), UTMSource: "sourceX", UTMMedium: "mediumX", UTMCampaign: "campaignX", UTMContent: "contentX", UTMTerm: "termX"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), UTMSource: "source2", UTMMedium: "medium2", UTMCampaign: "campaign2", UTMContent: "content2", UTMTerm: "term2"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), UTMSource: "source2", UTMMedium: "medium2", UTMCampaign: "campaign2", UTMContent: "content2", UTMTerm: "term2"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), UTMSource: "source3", UTMMedium: "medium3", UTMCampaign: "campaign3", UTMContent: "content3", UTMTerm: "term3"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	source, err := analyzer.UTMSource(nil)
	assert.NoError(t, err)
	assert.Len(t, source, 3)
	assert.Equal(t, "source1", source[0].UTMSource)
	assert.Equal(t, "source2", source[1].UTMSource)
	assert.Equal(t, "source3", source[2].UTMSource)
	assert.Equal(t, 3, source[0].Visitors)
	assert.Equal(t, 2, source[1].Visitors)
	assert.Equal(t, 1, source[2].Visitors)
	assert.InDelta(t, 0.5, source[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, source[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, source[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMSource(getMaxFilter(""))
	assert.NoError(t, err)
	medium, err := analyzer.UTMMedium(nil)
	assert.NoError(t, err)
	assert.Len(t, medium, 3)
	assert.Equal(t, "medium1", medium[0].UTMMedium)
	assert.Equal(t, "medium2", medium[1].UTMMedium)
	assert.Equal(t, "medium3", medium[2].UTMMedium)
	assert.Equal(t, 3, medium[0].Visitors)
	assert.Equal(t, 2, medium[1].Visitors)
	assert.Equal(t, 1, medium[2].Visitors)
	assert.InDelta(t, 0.5, medium[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, medium[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, medium[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMMedium(getMaxFilter(""))
	assert.NoError(t, err)
	campaign, err := analyzer.UTMCampaign(nil)
	assert.NoError(t, err)
	assert.Len(t, campaign, 3)
	assert.Equal(t, "campaign1", campaign[0].UTMCampaign)
	assert.Equal(t, "campaign2", campaign[1].UTMCampaign)
	assert.Equal(t, "campaign3", campaign[2].UTMCampaign)
	assert.Equal(t, 3, campaign[0].Visitors)
	assert.Equal(t, 2, campaign[1].Visitors)
	assert.Equal(t, 1, campaign[2].Visitors)
	assert.InDelta(t, 0.5, campaign[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, campaign[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, campaign[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMCampaign(getMaxFilter(""))
	assert.NoError(t, err)
	content, err := analyzer.UTMContent(nil)
	assert.NoError(t, err)
	assert.Len(t, content, 3)
	assert.Equal(t, "content1", content[0].UTMContent)
	assert.Equal(t, "content2", content[1].UTMContent)
	assert.Equal(t, "content3", content[2].UTMContent)
	assert.Equal(t, 3, content[0].Visitors)
	assert.Equal(t, 2, content[1].Visitors)
	assert.Equal(t, 1, content[2].Visitors)
	assert.InDelta(t, 0.5, content[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, content[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, content[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMContent(getMaxFilter(""))
	assert.NoError(t, err)
	term, err := analyzer.UTMTerm(nil)
	assert.NoError(t, err)
	assert.Len(t, term, 3)
	assert.Equal(t, "term1", term[0].UTMTerm)
	assert.Equal(t, "term2", term[1].UTMTerm)
	assert.Equal(t, "term3", term[2].UTMTerm)
	assert.Equal(t, 3, term[0].Visitors)
	assert.Equal(t, 2, term[1].Visitors)
	assert.Equal(t, 1, term[2].Visitors)
	assert.InDelta(t, 0.5, term[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, term[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, term[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMTerm(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.UTMTerm(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_AvgTimeOnPage(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(3), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 1, Time: pastDay(3).Add(time.Second * 9), SessionID: 3, Path: "/foo", DurationSeconds: 9, Title: "Foo"},
		{VisitorID: 2, Time: pastDay(3), SessionID: 3, Path: "/", Title: "Home"},
		{VisitorID: 2, Time: pastDay(3).Add(time.Second * 7), SessionID: 3, Path: "/foo", DurationSeconds: 7, Title: "Foo"},
		{VisitorID: 3, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 3, Time: pastDay(2).Add(time.Second * 5), SessionID: 2, Path: "/foo", DurationSeconds: 5, Title: "Foo"},
		{VisitorID: 4, Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{VisitorID: 4, Time: pastDay(2).Add(time.Second * 4), SessionID: 2, Path: "/foo", DurationSeconds: 4, Title: "Foo"},
		{VisitorID: 5, Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 5, Time: pastDay(1).Add(time.Second * 8), SessionID: 1, Path: "/foo", DurationSeconds: 8, Title: "Foo"},
		{VisitorID: 6, Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{VisitorID: 6, Time: pastDay(1).Add(time.Second * 6), SessionID: 1, Path: "/foo", DurationSeconds: 6, Title: "Foo"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: pastDay(3), SessionID: 3, EntryPath: "/", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: pastDay(3), SessionID: 3, EntryPath: "/", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, Time: pastDay(3), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 2, Time: pastDay(3), SessionID: 3, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 3, Time: pastDay(2), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 4, Time: pastDay(2), SessionID: 2, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 5, Time: pastDay(1), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
			{Sign: 1, VisitorID: 6, Time: pastDay(1), SessionID: 1, EntryPath: "/", ExitPath: "/foo"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	byDay, err := analyzer.AvgTimeOnPage(&Filter{Path: "/", From: pastDay(3), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.AvgTimeOnPage(&Filter{Path: "/foo", From: pastDay(3), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 0, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.AvgTimeOnPage(&Filter{MaxTimeOnPageSeconds: 5})
	assert.NoError(t, err)
	assert.Len(t, byDay, 3)
	assert.Equal(t, 5, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 5, byDay[2].AverageTimeSpentSeconds)
	_, err = analyzer.AvgTimeOnPage(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.AvgTimeOnPage(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_CalculateGrowth(t *testing.T) {
	analyzer := NewAnalyzer(dbClient, nil)
	growth := analyzer.calculateGrowth(0, 0)
	assert.InDelta(t, 0, growth, 0.001)
	growth = analyzer.calculateGrowth(1000, 0)
	assert.InDelta(t, 1, growth, 0.001)
	growth = analyzer.calculateGrowth(0, 1000)
	assert.InDelta(t, -1, growth, 0.001)
	growth = analyzer.calculateGrowth(100, 50)
	assert.InDelta(t, 1, growth, 0.001)
	growth = analyzer.calculateGrowth(50, 100)
	assert.InDelta(t, -0.5, growth, 0.001)
}

func TestAnalyzer_CalculateGrowthFloat64(t *testing.T) {
	analyzer := NewAnalyzer(dbClient, nil)
	growth := analyzer.calculateGrowthFloat64(0, 0)
	assert.InDelta(t, 0, growth, 0.001)
	growth = analyzer.calculateGrowthFloat64(1000, 0)
	assert.InDelta(t, 1, growth, 0.001)
	growth = analyzer.calculateGrowthFloat64(0, 1000)
	assert.InDelta(t, -1, growth, 0.001)
	growth = analyzer.calculateGrowthFloat64(100, 50)
	assert.InDelta(t, 1, growth, 0.001)
	growth = analyzer.calculateGrowthFloat64(50, 100)
	assert.InDelta(t, -0.5, growth, 0.001)
}

func TestAnalyzer_Timezone(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveSessions([]Session{
		{Sign: 1, VisitorID: 1, Time: pastDay(3).Add(time.Hour * 18), ExitPath: "/"}, // 18:00 UTC -> 03:00 Asia/Tokyo
		{Sign: 1, VisitorID: 2, Time: pastDay(2), ExitPath: "/"},                     // 00:00 UTC -> 09:00 Asia/Tokyo
		{Sign: 1, VisitorID: 3, Time: pastDay(1).Add(time.Hour * 19), ExitPath: "/"}, // 19:00 UTC -> 04:00 Asia/Tokyo
	}))
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: pastDay(3).Add(time.Hour * 18), Path: "/"}, // 18:00 UTC -> 03:00 Asia/Tokyo
		{VisitorID: 2, Time: pastDay(2), Path: "/"},                     // 00:00 UTC -> 09:00 Asia/Tokyo
		{VisitorID: 3, Time: pastDay(1).Add(time.Hour * 19), Path: "/"}, // 19:00 UTC -> 04:00 Asia/Tokyo
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Visitors(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	hours, err := analyzer.VisitorHours(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 1, hours[0].Visitors)
	assert.Equal(t, 1, hours[18].Visitors)
	assert.Equal(t, 1, hours[19].Visitors)
	timezone, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)
	visitors, err = analyzer.Visitors(&Filter{From: pastDay(3), To: pastDay(1), Timezone: timezone})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, 0, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 0, visitors[2].Visitors)
	hours, err = analyzer.VisitorHours(&Filter{From: pastDay(3), To: pastDay(1), Timezone: timezone})
	assert.NoError(t, err)
	assert.Equal(t, 1, hours[3].Visitors)
	assert.Equal(t, 0, hours[4].Visitors) // pushed to the next day, so outside of filter range
	assert.Equal(t, 1, hours[9].Visitors)
}

func TestAnalyzer_PathPattern(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: Today(), Path: "/"},
		{VisitorID: 2, Time: Today(), Path: "/simple/page"},
		{VisitorID: 3, Time: Today(), Path: "/siMple/page/"},
		{VisitorID: 4, Time: Today(), Path: "/simple/page/with/many/slashes"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: Today(), ExitPath: "/exit"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today(), ExitPath: "/exit"},
			{Sign: 1, VisitorID: 1, Time: Today(), ExitPath: "/"},
			{Sign: 1, VisitorID: 2, Time: Today(), ExitPath: "/simple/page"},
			{Sign: 1, VisitorID: 3, Time: Today(), ExitPath: "/siMple/page/"},
			{Sign: 1, VisitorID: 4, Time: Today(), ExitPath: "/simple/page/with/many/slashes"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Pages(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	visitors, err = analyzer.Pages(&Filter{PathPattern: "(?i)^/simple/[^/]+$"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	visitors, err = analyzer.Pages(&Filter{PathPattern: "(?i)^/simple/[^/]+/.*"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	visitors, err = analyzer.Pages(&Filter{PathPattern: "(?i)^/simple/[^/]+/slashes$"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 0)
	visitors, err = analyzer.Pages(&Filter{PathPattern: "(?i)^/simple/.+/slashes$"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
}

func TestAnalyzer_EntryExitPagePathFilter(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, SessionID: 1, Time: Today(), DurationSeconds: 0, Path: "/"},
		{VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 3), DurationSeconds: 3, Path: "/account/billing/"},
		{VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 5), DurationSeconds: 2, Path: "/settings/general/"},
		{VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 7), DurationSeconds: 2, Path: "/integrations/wordpress/"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 7), DurationSeconds: 7, ExitPath: "/", EntryPath: "/settings/general", PageViews: 4, IsBounce: false},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 7), DurationSeconds: 7, ExitPath: "/", EntryPath: "/settings/general", PageViews: 4, IsBounce: false},
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 7), DurationSeconds: 7, ExitPath: "/integrations/wordpress/", EntryPath: "/", PageViews: 4, IsBounce: false},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	filter := &Filter{
		Path:  "/account/billing/",
		Limit: 11,
	}
	entry, err := analyzer.EntryPages(filter)
	assert.NoError(t, err)
	assert.Len(t, entry, 1)
	assert.Equal(t, "/", entry[0].Path)
	assert.Equal(t, 1, entry[0].Visitors)
	assert.Equal(t, 1, entry[0].Entries)
	exit, err := analyzer.ExitPages(filter)
	assert.NoError(t, err)
	assert.Len(t, exit, 1)
	assert.Equal(t, "/integrations/wordpress/", exit[0].Path)
	assert.Equal(t, 1, exit[0].Visitors)
	assert.Equal(t, 1, exit[0].Exits)
}

func TestAnalyzer_EntryExitPageFilterCombination(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		// / -> /foo -> /bar -> /exit
		{VisitorID: 1, SessionID: 1, Time: Today(), Path: "/"},
		{VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 10), Path: "/foo"},
		{VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 20), Path: "/bar"},
		{VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 30), Path: "/exit"},

		// / -> /bar -> /
		{VisitorID: 2, SessionID: 2, Time: Today(), Path: "/"},
		{VisitorID: 2, SessionID: 2, Time: Today().Add(time.Second * 10), Path: "/bar"},
		{VisitorID: 2, SessionID: 2, Time: Today().Add(time.Second * 20), Path: "/"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 30), ExitPath: "/", EntryPath: "/exit", PageViews: 4, IsBounce: false},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 30), ExitPath: "/", EntryPath: "/exit", PageViews: 4, IsBounce: false},
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: Today().Add(time.Second * 30), ExitPath: "/exit", EntryPath: "/", PageViews: 4, IsBounce: false},
			{Sign: 1, VisitorID: 2, SessionID: 2, Time: Today().Add(time.Second * 20), ExitPath: "/", EntryPath: "/", PageViews: 3, IsBounce: false},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)

	// no filter
	pages, err := analyzer.Pages(nil)
	assert.NoError(t, err)
	assert.Len(t, pages, 4)
	assert.Equal(t, "/", pages[0].Path)
	assert.Equal(t, "/bar", pages[1].Path)
	assert.Equal(t, "/exit", pages[2].Path)
	assert.Equal(t, "/foo", pages[3].Path)
	assert.Equal(t, 2, pages[0].Visitors)
	assert.Equal(t, 2, pages[1].Visitors)
	assert.Equal(t, 1, pages[2].Visitors)
	assert.Equal(t, 1, pages[3].Visitors)
	entryPages, err := analyzer.EntryPages(nil)
	assert.NoError(t, err)
	assert.Len(t, entryPages, 1)
	assert.Equal(t, "/", entryPages[0].Path)
	assert.Equal(t, 2, entryPages[0].Visitors)
	assert.Equal(t, 2, entryPages[0].Sessions)
	assert.Equal(t, 2, entryPages[0].Entries)
	exitPages, err := analyzer.ExitPages(nil)
	assert.NoError(t, err)
	assert.Len(t, exitPages, 2)
	assert.Equal(t, "/", exitPages[0].Path)
	assert.Equal(t, "/exit", exitPages[1].Path)
	assert.Equal(t, 2, exitPages[0].Visitors)
	assert.Equal(t, 1, exitPages[1].Visitors)
	assert.Equal(t, 1, exitPages[0].Exits)
	assert.Equal(t, 1, exitPages[1].Exits)

	// filter for a path
	filter := &Filter{Path: "/bar"}
	pages, err = analyzer.Pages(filter)
	assert.NoError(t, err)
	assert.Len(t, pages, 1)
	assert.Equal(t, "/bar", pages[0].Path)
	assert.Equal(t, 2, pages[0].Visitors)
	entryPages, err = analyzer.EntryPages(filter)
	assert.NoError(t, err)
	assert.Len(t, entryPages, 1)
	assert.Equal(t, "/", entryPages[0].Path)
	assert.Equal(t, 2, entryPages[0].Visitors)
	assert.Equal(t, 2, entryPages[0].Entries)
	exitPages, err = analyzer.ExitPages(filter)
	assert.NoError(t, err)
	assert.Len(t, exitPages, 2)
	assert.Equal(t, "/", exitPages[0].Path)
	assert.Equal(t, "/exit", exitPages[1].Path)
	assert.Equal(t, 2, exitPages[0].Visitors)
	assert.Equal(t, 1, exitPages[1].Visitors)
	assert.Equal(t, 1, exitPages[0].Exits)
	assert.Equal(t, 1, exitPages[1].Exits)

	// filter entry page
	filter.Path = ""
	filter.EntryPath = "/bar"
	pages, err = analyzer.Pages(filter)
	assert.NoError(t, err)
	assert.Len(t, pages, 0)
	entryPages, err = analyzer.EntryPages(filter)
	assert.NoError(t, err)
	assert.Len(t, entryPages, 0)
	exitPages, err = analyzer.ExitPages(filter)
	assert.NoError(t, err)
	assert.Len(t, exitPages, 0)

	filter.EntryPath = "/"
	pages, err = analyzer.Pages(filter)
	assert.NoError(t, err)
	assert.Len(t, pages, 4)
	assert.Equal(t, "/", pages[0].Path)
	assert.Equal(t, "/bar", pages[1].Path)
	assert.Equal(t, "/exit", pages[2].Path)
	assert.Equal(t, "/foo", pages[3].Path)
	assert.Equal(t, 2, pages[0].Visitors)
	assert.Equal(t, 2, pages[1].Visitors)
	assert.Equal(t, 1, pages[2].Visitors)
	assert.Equal(t, 1, pages[3].Visitors)
	entryPages, err = analyzer.EntryPages(filter)
	assert.NoError(t, err)
	assert.Len(t, entryPages, 1)
	assert.Equal(t, "/", entryPages[0].Path)
	assert.Equal(t, 2, entryPages[0].Visitors)
	assert.Equal(t, 2, entryPages[0].Entries)
	exitPages, err = analyzer.ExitPages(filter)
	assert.NoError(t, err)
	assert.Len(t, exitPages, 2)
	assert.Equal(t, "/", exitPages[0].Path)
	assert.Equal(t, "/exit", exitPages[1].Path)
	assert.Equal(t, 2, exitPages[0].Visitors)
	assert.Equal(t, 1, exitPages[1].Visitors)
	assert.Equal(t, 1, exitPages[0].Exits)
	assert.Equal(t, 1, exitPages[1].Exits)

	// filter entry + exit page
	filter.ExitPath = "/bar"
	pages, err = analyzer.Pages(filter)
	assert.NoError(t, err)
	assert.Len(t, pages, 0)
	entryPages, err = analyzer.EntryPages(filter)
	assert.NoError(t, err)
	assert.Len(t, entryPages, 0)
	exitPages, err = analyzer.ExitPages(filter)
	assert.NoError(t, err)
	assert.Len(t, exitPages, 0)

	filter.ExitPath = "/exit"
	pages, err = analyzer.Pages(filter)
	assert.NoError(t, err)
	assert.Len(t, pages, 4)
	assert.Equal(t, "/", pages[0].Path)
	assert.Equal(t, "/bar", pages[1].Path)
	assert.Equal(t, "/exit", pages[2].Path)
	assert.Equal(t, "/foo", pages[3].Path)
	assert.Equal(t, 1, pages[0].Visitors)
	assert.Equal(t, 1, pages[1].Visitors)
	assert.Equal(t, 1, pages[2].Visitors)
	assert.Equal(t, 1, pages[3].Visitors)
	entryPages, err = analyzer.EntryPages(filter)
	assert.NoError(t, err)
	assert.Len(t, entryPages, 1)
	assert.Equal(t, "/", entryPages[0].Path)
	assert.Equal(t, 2, entryPages[0].Visitors)
	assert.Equal(t, 1, entryPages[0].Entries)
	exitPages, err = analyzer.ExitPages(filter)
	assert.NoError(t, err)
	assert.Len(t, exitPages, 1)
	assert.Equal(t, "/exit", exitPages[0].Path)
	assert.Equal(t, 1, exitPages[0].Visitors)
	assert.Equal(t, 1, exitPages[0].Exits)

	// filter entry + exit page + page
	filter.Path = "/bar"
	pages, err = analyzer.Pages(filter)
	assert.NoError(t, err)
	assert.Len(t, pages, 1)
	assert.Equal(t, "/bar", pages[0].Path)
	assert.Equal(t, 1, pages[0].Visitors)
	entryPages, err = analyzer.EntryPages(filter)
	assert.NoError(t, err)
	assert.Len(t, entryPages, 1)
	assert.Equal(t, "/", entryPages[0].Path)
	assert.Equal(t, 2, entryPages[0].Visitors)
	assert.Equal(t, 1, entryPages[0].Entries)
	exitPages, err = analyzer.ExitPages(filter)
	assert.NoError(t, err)
	assert.Len(t, exitPages, 1)
	assert.Equal(t, "/exit", exitPages[0].Path)
	assert.Equal(t, 1, exitPages[0].Visitors)
	assert.Equal(t, 1, exitPages[0].Exits)
}

func TestAnalyzer_totalVisitorsSessions(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, SessionID: 1, Time: Today(), Path: "/"},
		{VisitorID: 1, SessionID: 1, Time: Today(), Path: "/foo"},
		{VisitorID: 1, SessionID: 1, Time: Today(), Path: "/bar"},
		{VisitorID: 1, SessionID: 1, Time: Today(), Path: "/bar"},
		{VisitorID: 1, SessionID: 2, Time: Today(), Path: "/foo"},
		{VisitorID: 2, SessionID: 1, Time: Today(), Path: "/"},
		{VisitorID: 2, SessionID: 2, Time: Today(), Path: "/foo"},
		{VisitorID: 3, SessionID: 1, Time: Today(), Path: "/"},
		{VisitorID: 3, SessionID: 1, Time: Today(), Path: "/foo"},
	}))
	assert.NoError(t, dbClient.SaveSessions([]Session{
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: Today()},
		{Sign: 1, VisitorID: 1, SessionID: 2, Time: Today()},
		{Sign: 1, VisitorID: 2, SessionID: 1, Time: Today()},
		{Sign: 1, VisitorID: 2, SessionID: 2, Time: Today()},
		{Sign: 1, VisitorID: 3, SessionID: 1, Time: Today()},
		{Sign: 1, VisitorID: 3, SessionID: 1, Time: Today()},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	total, err := analyzer.totalVisitorsSessions(nil, []string{"/", "/foo", "/bar"})
	assert.NoError(t, err)
	assert.Len(t, total, 3)
	assert.Equal(t, "/foo", total[0].Path)
	assert.Equal(t, "/", total[1].Path)
	assert.Equal(t, "/bar", total[2].Path)
	assert.Equal(t, 4, total[0].Views)
	assert.Equal(t, 3, total[1].Views)
	assert.Equal(t, 2, total[2].Views)
	assert.Equal(t, 3, total[0].Visitors)
	assert.Equal(t, 3, total[1].Visitors)
	assert.Equal(t, 1, total[2].Visitors)
	assert.Equal(t, 4, total[0].Sessions)
	assert.Equal(t, 3, total[1].Sessions)
	assert.Equal(t, 1, total[2].Sessions)
	total, err = analyzer.totalVisitorsSessions(nil, []string{"/"})
	assert.NoError(t, err)
	assert.Len(t, total, 1)
	assert.Equal(t, "/", total[0].Path)
	assert.Equal(t, 3, total[0].Views)
	assert.Equal(t, 3, total[0].Visitors)
	assert.Equal(t, 3, total[0].Sessions)
	_, err = analyzer.totalVisitorsSessions(getMaxFilter(""), []string{"/", "/foo", "/bar"})
	assert.NoError(t, err)
}

func TestAnalyzer_avgTimeOnPage(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: Today(), Path: "/"},
		{VisitorID: 1, Time: Today().Add(time.Minute * 2), Path: "/foo", DurationSeconds: 120},
		{VisitorID: 1, Time: Today().Add(time.Minute*2 + time.Second*23), Path: "/bar", DurationSeconds: 23},

		{VisitorID: 2, Time: Today(), Path: "/bar"},
		{VisitorID: 2, Time: Today().Add(time.Second * 16), Path: "/foo", DurationSeconds: 16},
		{VisitorID: 2, Time: Today().Add(time.Second*16 + time.Second*8), Path: "/", DurationSeconds: 8},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: Today(), EntryPath: "/bar", ExitPath: "/"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today(), EntryPath: "/bar", ExitPath: "/"},
			{Sign: 1, VisitorID: 1, Time: Today(), EntryPath: "/", ExitPath: "/bar"},
			{Sign: 1, VisitorID: 2, Time: Today(), EntryPath: "/bar", ExitPath: "/"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	stats, err := analyzer.avgTimeOnPage(nil, []string{"/", "/foo", "/bar"})
	assert.NoError(t, err)
	assert.Len(t, stats, 3)
	paths := []string{stats[0].Path, stats[1].Path, stats[2].Path}
	assert.Contains(t, paths, "/")
	assert.Contains(t, paths, "/foo")
	assert.Contains(t, paths, "/bar")
	top := []int{stats[0].AverageTimeSpentSeconds, stats[1].AverageTimeSpentSeconds, stats[2].AverageTimeSpentSeconds}
	assert.Contains(t, top, 120)
	assert.Contains(t, top, (23+8)/2)
	assert.Contains(t, top, 16)
	_, err = analyzer.avgTimeOnPage(getMaxFilter(""), []string{"/", "/foo", "/bar"})
	assert.NoError(t, err)
}

func TestAnalyzer_NoData(t *testing.T) {
	cleanupDB()
	analyzer := NewAnalyzer(dbClient, nil)
	_, _, err := analyzer.ActiveVisitors(nil, time.Minute*15)
	assert.NoError(t, err)
	_, err = analyzer.Visitors(nil)
	assert.NoError(t, err)
	_, err = analyzer.Growth(&Filter{From: pastDay(7), To: Today()})
	assert.NoError(t, err)
	_, err = analyzer.VisitorHours(nil)
	assert.NoError(t, err)
	_, err = analyzer.Pages(nil)
	assert.NoError(t, err)
	_, err = analyzer.EntryPages(nil)
	assert.NoError(t, err)
	_, err = analyzer.ExitPages(nil)
	assert.NoError(t, err)
	_, err = analyzer.PageConversions(nil)
	assert.NoError(t, err)
	_, err = analyzer.Events(nil)
	assert.NoError(t, err)
	_, err = analyzer.EventBreakdown(&Filter{EventName: "event"})
	assert.NoError(t, err)
	_, err = analyzer.Referrer(nil)
	assert.NoError(t, err)
	_, err = analyzer.Platform(nil)
	assert.NoError(t, err)
	_, err = analyzer.Languages(nil) // other metadata works the same...
	assert.NoError(t, err)
	_, err = analyzer.OSVersion(nil)
	assert.NoError(t, err)
	_, err = analyzer.BrowserVersion(nil)
	assert.NoError(t, err)
	_, err = analyzer.AvgSessionDuration(nil)
	assert.NoError(t, err)
	_, err = analyzer.AvgTimeOnPage(nil)
	assert.NoError(t, err)
}

func getMaxFilter(eventName string) *Filter {
	return &Filter{
		ClientID:       42,
		From:           pastDay(5),
		To:             pastDay(2),
		Path:           "/path",
		EntryPath:      "/entry",
		ExitPath:       "/exit",
		Language:       "en",
		Country:        "en",
		City:           "London",
		Referrer:       "ref",
		ReferrerName:   "refname",
		OS:             OSWindows,
		OSVersion:      "10",
		Browser:        BrowserChrome,
		BrowserVersion: "90",
		Platform:       PlatformDesktop,
		ScreenClass:    "XL",
		UTMSource:      "source",
		UTMMedium:      "medium",
		UTMCampaign:    "campaign",
		UTMContent:     "content",
		UTMTerm:        "term",
		EventName:      eventName,
		Limit:          42,
	}
}

func saveSessions(t *testing.T, sessions [][]Session) {
	for _, entries := range sessions {
		assert.NoError(t, dbClient.SaveSessions(entries))
		time.Sleep(time.Millisecond * 20)
	}
}

func getWeek(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}
