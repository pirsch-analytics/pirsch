package pirsch

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_ActiveVisitors(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 30), Path: "/", Title: "Home"},
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 15), Path: "/", Title: "Home"},
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 5), Path: "/bar", Title: "Bar"},
		{Fingerprint: "fp2", Time: time.Now().Add(-time.Minute * 4), Path: "/bar", Title: "Bar"},
		{Fingerprint: "fp2", Time: time.Now().Add(-time.Minute * 3), Path: "/foo", Title: "Foo"},
		{Fingerprint: "fp3", Time: time.Now().Add(-time.Minute * 3), Path: "/", Title: "Home"},
		{Fingerprint: "fp4", Time: time.Now().Add(-time.Minute), Path: "/", Title: "Home"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, count, err := analyzer.ActiveVisitors(nil, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path)
	assert.Equal(t, "/bar", visitors[1].Path)
	assert.Equal(t, "/foo", visitors[2].Path)
	// TODO
	/*assert.Empty(t, visitors[0].Title)
	assert.Empty(t, visitors[1].Title)
	assert.Empty(t, visitors[2].Title)*/
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	visitors, count, err = analyzer.ActiveVisitors(&Filter{Path: "/bar"}, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "/bar", visitors[0].Path)
	assert.Equal(t, 2, visitors[0].Visitors)
	_, _, err = analyzer.ActiveVisitors(getMaxFilter(), time.Minute*10)
	assert.NoError(t, err)
	// TODO
	/*visitors, count, err = analyzer.ActiveVisitors(&Filter{IncludeTitle: true}, time.Minute*10)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Home", visitors[0].Title)
	assert.Equal(t, "Bar", visitors[1].Title)
	assert.Equal(t, "Foo", visitors[2].Title)*/
}

func TestAnalyzer_VisitorsAndAvgSessionDuration(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(4), SessionID: 4, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/foo", PageViews: 2, IsBounce: false, DurationSeconds: 300},
		{Fingerprint: "fp1", Time: pastDay(4), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp2", Time: pastDay(4), SessionID: 4, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp2", Time: pastDay(4).Add(time.Minute * 10), SessionID: 41, Path: "/bar", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp3", Time: pastDay(4), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp4", Time: pastDay(4), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp5", Time: pastDay(2), SessionID: 2, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp5", Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 300},
		{Fingerprint: "fp6", Time: pastDay(2), SessionID: 2, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, Path: "/bar", PageViews: 2, IsBounce: false, DurationSeconds: 600},
		{Fingerprint: "fp7", Time: pastDay(2), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp8", Time: pastDay(2), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp9", Time: Today(), Path: "/", PageViews: 1, IsBounce: true},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Visitors(&Filter{From: pastDay(4), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, pastDay(4), visitors[0].Day)
	assert.Equal(t, pastDay(3), visitors[1].Day)
	assert.Equal(t, pastDay(2), visitors[2].Day)
	assert.Equal(t, pastDay(1), visitors[3].Day)
	assert.Equal(t, Today(), visitors[4].Day)
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
	assert.InDelta(t, 0.83, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[2].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[3].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[4].BounceRate, 0.01)
	asd, err := analyzer.AvgSessionDuration(nil)
	assert.NoError(t, err)
	assert.Len(t, asd, 2)
	assert.Equal(t, pastDay(4), asd[0].Day)
	assert.Equal(t, pastDay(2), asd[1].Day)
	assert.Equal(t, 300, asd[0].AverageTimeSpentSeconds)
	assert.Equal(t, 450, asd[1].AverageTimeSpentSeconds)
	tsd, err := analyzer.totalSessionDuration(nil)
	assert.NoError(t, err)
	assert.Equal(t, 1200, tsd)
	visitors, err = analyzer.Visitors(&Filter{From: pastDay(4), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, pastDay(4), visitors[0].Day)
	assert.Equal(t, pastDay(2), visitors[2].Day)
	asd, err = analyzer.AvgSessionDuration(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, asd, 3)
	tsd, err = analyzer.totalSessionDuration(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 900, tsd)
	_, err = analyzer.Visitors(getMaxFilter())
	assert.NoError(t, err)
	_, err = analyzer.AvgSessionDuration(getMaxFilter())
	assert.NoError(t, err)
	_, err = analyzer.totalSessionDuration(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Growth(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(4), SessionID: 4, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/foo", DurationSeconds: 300, PageViews: 2, IsBounce: false},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 15), SessionID: 4, Path: "/bar", DurationSeconds: 600, PageViews: 3, IsBounce: false},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp3", Time: pastDay(4), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp4", Time: pastDay(3), SessionID: 3, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp4", Time: pastDay(3).Add(time.Minute * 5), SessionID: 3, Path: "/foo", DurationSeconds: 300, PageViews: 2, IsBounce: false},
		{Fingerprint: "fp4", Time: pastDay(3), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp5", Time: pastDay(3), SessionID: 3, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp5", Time: pastDay(3).Add(time.Minute * 10), SessionID: 31, Path: "/bar", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp6", Time: pastDay(3), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp7", Time: pastDay(3), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp8", Time: pastDay(2), SessionID: 2, Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp8", Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar", DurationSeconds: 300, PageViews: 2, IsBounce: false},
		{Fingerprint: "fp9", Time: pastDay(2), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp10", Time: pastDay(2), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp11", Time: Today(), Path: "/", PageViews: 1, IsBounce: true},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Growth(nil)
	assert.ErrorIs(t, err, ErrNoPeriodOrDay)
	assert.Nil(t, growth)
	growth, err = analyzer.Growth(&Filter{Day: pastDay(2)})
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
	assert.InDelta(t, -0.3333, growth.TimeSpentGrowth, 0.001)
	_, err = analyzer.Growth(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_GrowthEvents(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveEvents([]Event{
		{Name: "event1", Hit: Hit{Fingerprint: "fp1", Time: pastDay(4), SessionID: 4, Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", DurationSeconds: 300, Hit: Hit{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 5), SessionID: 4, Path: "/foo", DurationSeconds: 300, PageViews: 2, IsBounce: false}},
		{Name: "event1", DurationSeconds: 600, Hit: Hit{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 15), SessionID: 4, Path: "/bar", DurationSeconds: 600, PageViews: 3, IsBounce: false}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp2", Time: pastDay(4), Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp3", Time: pastDay(4), Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp4", Time: pastDay(3), SessionID: 3, Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", DurationSeconds: 300, Hit: Hit{Fingerprint: "fp4", Time: pastDay(3).Add(time.Minute * 5), SessionID: 3, Path: "/foo", DurationSeconds: 300, PageViews: 2, IsBounce: false}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp4", Time: pastDay(3), Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp5", Time: pastDay(3), SessionID: 3, Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp5", Time: pastDay(3).Add(time.Minute * 10), SessionID: 31, Path: "/bar", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp6", Time: pastDay(3), Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp7", Time: pastDay(3), Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp8", Time: pastDay(2), SessionID: 2, Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", DurationSeconds: 300, Hit: Hit{Fingerprint: "fp8", Time: pastDay(2).Add(time.Minute * 5), SessionID: 2, Path: "/bar", DurationSeconds: 300, PageViews: 2, IsBounce: false}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp9", Time: pastDay(2), Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp10", Time: pastDay(2), Path: "/", PageViews: 1, IsBounce: true}},
		{Name: "event1", Hit: Hit{Fingerprint: "fp11", Time: Today(), Path: "/", PageViews: 1, IsBounce: true}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	growth, err := analyzer.Growth(nil)
	assert.ErrorIs(t, err, ErrNoPeriodOrDay)
	assert.Nil(t, growth)
	growth, err = analyzer.Growth(&Filter{Day: pastDay(2), EventName: "event1"})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, -0.25, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, -0.4285, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, -0.5, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, -0.2, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	growth, err = analyzer.Growth(&Filter{From: pastDay(3), To: pastDay(2), EventName: "event1"})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1.3333, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1.2, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 2, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 0.1666, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, -0.3333, growth.TimeSpentGrowth, 0.001)
	maxFilter := getMaxFilter()
	maxFilter.EventName = "event1"
	_, err = analyzer.Growth(maxFilter)
	assert.NoError(t, err)
}

func TestAnalyzer_VisitorHours(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp1", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp1", Time: pastDay(2).Add(time.Hour * 3), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(2).Add(time.Hour * 5), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(2).Add(time.Hour * 8), Path: "/"},
		{Fingerprint: "fp3", Time: pastDay(1).Add(time.Hour * 4), Path: "/"},
		{Fingerprint: "fp4", Time: pastDay(1).Add(time.Hour * 5), Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(1).Add(time.Hour * 8), Path: "/"},
		{Fingerprint: "fp6", Time: Today().Add(time.Hour * 3), Path: "/"},
		{Fingerprint: "fp6", Time: Today().Add(time.Hour * 5), Path: "/"},
		{Fingerprint: "fp7", Time: Today().Add(time.Hour * 10), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.VisitorHours(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 0, visitors[0].Hour)
	assert.Equal(t, 3, visitors[3].Hour)
	assert.Equal(t, 4, visitors[4].Hour)
	assert.Equal(t, 5, visitors[5].Hour)
	assert.Equal(t, 8, visitors[8].Hour)
	assert.Equal(t, 10, visitors[10].Hour)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 3, visitors[5].Visitors)
	assert.Equal(t, 2, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)
	visitors, err = analyzer.VisitorHours(&Filter{From: pastDay(1), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 24)
	assert.Equal(t, 3, visitors[3].Hour)
	assert.Equal(t, 4, visitors[4].Hour)
	assert.Equal(t, 5, visitors[5].Hour)
	assert.Equal(t, 8, visitors[8].Hour)
	assert.Equal(t, 10, visitors[10].Hour)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 2, visitors[5].Visitors)
	assert.Equal(t, 1, visitors[8].Visitors)
	assert.Equal(t, 1, visitors[10].Visitors)
	_, err = analyzer.VisitorHours(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_PagesAndAvgTimeOnPage(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 3), SessionID: 4, DurationSeconds: 180, Path: "/foo", Title: "Foo", IsBounce: false, PageViews: 2},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Hour), SessionID: 41, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp2", Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp2", Time: pastDay(4).Add(time.Minute * 2), SessionID: 4, DurationSeconds: 120, Path: "/bar", Title: "Bar", IsBounce: false, PageViews: 2},
		{Fingerprint: "fp3", Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp4", Time: pastDay(4), SessionID: 4, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp5", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp5", Time: pastDay(2).Add(time.Minute * 5), SessionID: 21, Path: "/bar", Title: "Bar", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp6", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 10), SessionID: 2, DurationSeconds: 600, Path: "/bar", Title: "Bar", IsBounce: false, PageViews: 2},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 11), SessionID: 21, Path: "/bar", Title: "Bar", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 21), SessionID: 21, DurationSeconds: 600, Path: "/foo", Title: "Foo", IsBounce: false, PageViews: 2},
		{Fingerprint: "fp7", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp8", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
		{Fingerprint: "fp9", Time: Today(), SessionID: 2, Path: "/", Title: "Home", IsBounce: true, PageViews: 1},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Pages(nil)
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
	assert.InDelta(t, 0.625, visitors[0].RelativeViews, 0.01)
	assert.InDelta(t, 0.25, visitors[1].RelativeViews, 0.01)
	assert.InDelta(t, 0.125, visitors[2].RelativeViews, 0.01)
	assert.Equal(t, 7, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.InDelta(t, 0.7, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.25, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	assert.Equal(t, 300, visitors[0].AverageTimeSpentSeconds)
	assert.Equal(t, 440, visitors[1].AverageTimeSpentSeconds)
	assert.Equal(t, 390, visitors[2].AverageTimeSpentSeconds)
	top, err := analyzer.AvgTimeOnPage(nil)
	assert.NoError(t, err)
	assert.Len(t, top, 2)
	assert.Equal(t, pastDay(4), top[0].Day)
	assert.Equal(t, pastDay(2), top[1].Day)
	assert.Equal(t, 150, top[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, top[1].AverageTimeSpentSeconds)
	ttop, err := analyzer.totalTimeOnPage(nil)
	assert.NoError(t, err)
	assert.Equal(t, 1500, ttop)
	visitors, err = analyzer.Pages(&Filter{From: pastDay(3), To: pastDay(1), IncludeTitle: true})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path)
	assert.Equal(t, "/bar", visitors[1].Path)
	assert.Equal(t, "/foo", visitors[2].Path)
	// TODO
	/*assert.Equal(t, "Home", visitors[0].Title)
	assert.Equal(t, "Bar", visitors[1].Title)*/
	assert.Equal(t, 600, visitors[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, visitors[1].AverageTimeSpentSeconds)
	assert.Equal(t, 600, visitors[2].AverageTimeSpentSeconds)
	top, err = analyzer.AvgTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, top, 3)
	assert.Equal(t, pastDay(3), top[0].Day)
	assert.Equal(t, pastDay(2), top[1].Day)
	assert.Equal(t, pastDay(1), top[2].Day)
	assert.Equal(t, 0, top[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, top[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, top[2].AverageTimeSpentSeconds)
	ttop, err = analyzer.totalTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 1200, ttop)
	_, err = analyzer.Pages(getMaxFilter())
	assert.NoError(t, err)
	_, err = analyzer.totalTimeOnPage(getMaxFilter())
	assert.NoError(t, err)
	visitors, err = analyzer.Pages(&Filter{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	ttop, err = analyzer.totalTimeOnPage(&Filter{MaxTimeOnPageSeconds: 200})
	assert.NoError(t, err)
	assert.Equal(t, 180+120+200+200, ttop)
}

// TODO
/*func TestAnalyzer_PageTitle(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(2), Path: "/", Title: "Home 1"},
		{Fingerprint: "fp1", Time: pastDay(1), Path: "/", Title: "Home 2", DurationSeconds: 42},
		{Fingerprint: "fp2", Time: Today(), Path: "/foo", Title: "Foo"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Pages(&Filter{IncludeTitle: true, IncludeAvgTimeOnPage: true})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Foo", visitors[0].Title)
	assert.Equal(t, "Home 1", visitors[1].Title)
	assert.Equal(t, "Home 2", visitors[2].Title)
	assert.Equal(t, 0, visitors[0].AverageTimeSpentSeconds)
	assert.Equal(t, 42, visitors[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, visitors[2].AverageTimeSpentSeconds)
}*/

func TestAnalyzer_EntryExitPages(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home", EntryPath: "/"},
		{Fingerprint: "fp1", Time: pastDay(2).Add(time.Second), SessionID: 2, Path: "/", Title: "Home", EntryPath: "/"},
		{Fingerprint: "fp1", Time: pastDay(2).Add(time.Second * 10), SessionID: 2, DurationSeconds: 10, Path: "/foo", Title: "Foo", EntryPath: "/"},
		{Fingerprint: "fp2", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home", EntryPath: "/"},
		{Fingerprint: "fp3", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home", EntryPath: "/"},
		{Fingerprint: "fp4", Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home", EntryPath: "/"},
		{Fingerprint: "fp4", Time: pastDay(1).Add(time.Second * 20), SessionID: 1, DurationSeconds: 20, Path: "/bar", Title: "Bar", EntryPath: "/"},
		{Fingerprint: "fp5", Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home", EntryPath: "/"},
		{Fingerprint: "fp5", Time: pastDay(1).Add(time.Second * 40), SessionID: 1, DurationSeconds: 40, Path: "/bar", Title: "Bar", EntryPath: "/"},
		{Fingerprint: "fp6", Time: pastDay(1), SessionID: 1, Path: "/bar", Title: "Bar", EntryPath: "/bar"},
		{Fingerprint: "fp7", Time: pastDay(1), SessionID: 1, Path: "/bar", Title: "Bar", EntryPath: "/bar"},
		{Fingerprint: "fp7", Time: pastDay(1).Add(time.Minute), SessionID: 1, Path: "/", Title: "Home", EntryPath: "/bar"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	entries, err := analyzer.EntryPages(nil)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "/", entries[0].Path)
	assert.Equal(t, "/bar", entries[1].Path)
	assert.Empty(t, entries[0].Title)
	assert.Empty(t, entries[1].Title)
	assert.Equal(t, 6, entries[0].Visitors)
	assert.Equal(t, 4, entries[1].Visitors)
	assert.Equal(t, 5, entries[0].Entries)
	assert.Equal(t, 2, entries[1].Entries)
	assert.InDelta(t, 0.8333, entries[0].EntryRate, 0.001)
	assert.InDelta(t, 0.5, entries[1].EntryRate, 0.001)
	assert.Equal(t, 23, entries[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, entries[1].AverageTimeSpentSeconds)
	entries, err = analyzer.EntryPages(&Filter{From: pastDay(1), To: Today(), IncludeTitle: true})
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "/", entries[0].Path)
	assert.Equal(t, "/bar", entries[1].Path)
	// TODO
	/*assert.Equal(t, "Home", entries[0].Title)
	assert.Equal(t, "Bar", entries[1].Title)*/
	assert.Equal(t, 3, entries[0].Visitors)
	assert.Equal(t, 4, entries[1].Visitors)
	assert.Equal(t, 2, entries[0].Entries)
	assert.Equal(t, 2, entries[1].Entries)
	assert.InDelta(t, 0.6666, entries[0].EntryRate, 0.001)
	assert.InDelta(t, 0.5, entries[1].EntryRate, 0.001)
	assert.Equal(t, 30, entries[0].AverageTimeSpentSeconds)
	assert.Equal(t, 0, entries[1].AverageTimeSpentSeconds)
	entries, err = analyzer.EntryPages(&Filter{From: pastDay(1), To: Today(), EntryPath: "/"})
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "/", entries[0].Path)
	assert.Equal(t, 3, entries[0].Visitors)
	assert.Equal(t, 2, entries[0].Entries)
	assert.InDelta(t, 0.6666, entries[0].EntryRate, 0.001)
	assert.Equal(t, 30, entries[0].AverageTimeSpentSeconds)
	/*_, err = analyzer.EntryPages(getMaxFilter())
	assert.NoError(t, err)*/
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
	assert.Equal(t, 3, exits[0].Exits)
	assert.Equal(t, 3, exits[1].Exits)
	assert.Equal(t, 1, exits[2].Exits)
	assert.InDelta(t, 0.5, exits[0].ExitRate, 0.001)
	assert.InDelta(t, 0.75, exits[1].ExitRate, 0.001)
	assert.InDelta(t, 1, exits[2].ExitRate, 0.001)
	// TODO
	exits, err = analyzer.ExitPages(&Filter{From: pastDay(1), To: Today() /*IncludeTitle: true,*/})
	assert.NoError(t, err)
	assert.Len(t, exits, 2)
	assert.Equal(t, "/bar", exits[0].Path)
	assert.Equal(t, "/", exits[1].Path)
	/*assert.Equal(t, "Bar", exits[0].Title)
	assert.Equal(t, "Home", exits[1].Title)*/
	assert.Equal(t, 4, exits[0].Visitors)
	assert.Equal(t, 3, exits[1].Visitors)
	assert.Equal(t, 3, exits[0].Exits)
	assert.Equal(t, 1, exits[1].Exits)
	assert.InDelta(t, 0.75, exits[0].ExitRate, 0.001)
	assert.InDelta(t, 0.33, exits[1].ExitRate, 0.01)
	exits, err = analyzer.ExitPages(&Filter{From: pastDay(1), To: Today(), ExitPath: "/"})
	assert.NoError(t, err)
	assert.Len(t, exits, 1)
	assert.Equal(t, "/", exits[0].Path)
	assert.Equal(t, 3, exits[0].Visitors)
	assert.Equal(t, 1, exits[0].Exits)
	assert.InDelta(t, 0.3333, exits[0].ExitRate, 0.01)
	_, err = analyzer.ExitPages(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_PageConversions(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: Today(), Path: "/", PageViews: 1},
		{Fingerprint: "fp2", Time: Today(), Path: "/simple/page", PageViews: 1},
		{Fingerprint: "fp2", Time: Today().Add(time.Minute), Path: "/simple/page", PageViews: 2},
		{Fingerprint: "fp3", Time: Today(), Path: "/siMple/page/", PageViews: 1},
		{Fingerprint: "fp3", Time: Today().Add(time.Minute), Path: "/siMple/page/", PageViews: 2},
		{Fingerprint: "fp4", Time: Today(), Path: "/simple/page/with/many/slashes", PageViews: 1},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.PageConversions(nil)
	assert.NoError(t, err)
	assert.Equal(t, 4, stats.Visitors)
	assert.Equal(t, 6, stats.Views)
	assert.InDelta(t, 1, stats.CR, 0.01)
	stats, err = analyzer.PageConversions(&Filter{PathPattern: "(?i)^/simple/[^/]+/.*"})
	assert.NoError(t, err)
	assert.Equal(t, 2, stats.Visitors)
	assert.Equal(t, 3, stats.Views)
	assert.InDelta(t, 0.5, stats.CR, 0.01)
	_, err = analyzer.PageConversions(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Events(t *testing.T) {
	cleanupDB()

	// create hits for the conversion rate
	for i := 0; i < 10; i++ {
		assert.NoError(t, dbClient.SaveHits([]Hit{
			{Fingerprint: fmt.Sprintf("fp%d", i), Time: Today(), Path: "/"},
		}))
	}

	assert.NoError(t, dbClient.SaveEvents([]Event{
		{Name: "event1", DurationSeconds: 5, MetaKeys: []string{"status", "price"}, MetaValues: []string{"in", "34.56"}, Hit: Hit{Fingerprint: "fp1", Time: Today(), Path: "/", PageViews: 1}},
		{Name: "event1", DurationSeconds: 8, MetaKeys: []string{"status", "price"}, MetaValues: []string{"out", "34.56"}, Hit: Hit{Fingerprint: "fp2", Time: Today(), Path: "/simple/page", PageViews: 1}},
		{Name: "event1", DurationSeconds: 3, Hit: Hit{Fingerprint: "fp3", Time: Today(), Path: "/simple/page/1", PageViews: 1}},
		{Name: "event1", DurationSeconds: 8, Hit: Hit{Fingerprint: "fp3", Time: Today().Add(time.Minute), Path: "/simple/page/2", PageViews: 2}},
		{Name: "event1", DurationSeconds: 2, MetaKeys: []string{"status"}, MetaValues: []string{"in"}, Hit: Hit{Fingerprint: "fp4", Time: Today(), Path: "/", PageViews: 1}},
		{Name: "event2", DurationSeconds: 1, Hit: Hit{Fingerprint: "fp1", Time: Today(), Path: "/", PageViews: 1}},
		{Name: "event2", DurationSeconds: 5, Hit: Hit{Fingerprint: "fp2", Time: Today(), Path: "/", PageViews: 1}},
		{Name: "event2", DurationSeconds: 7, MetaKeys: []string{"status", "price"}, MetaValues: []string{"in", "34.56"}, Hit: Hit{Fingerprint: "fp2", Time: Today().Add(time.Minute), Path: "/simple/page", PageViews: 2}},
		{Name: "event2", DurationSeconds: 9, MetaKeys: []string{"status", "price", "third"}, MetaValues: []string{"in", "13.74", "param"}, Hit: Hit{Fingerprint: "fp3", Time: Today(), Path: "/simple/page", PageViews: 1}},
		{Name: "event2", DurationSeconds: 3, MetaKeys: []string{"price"}, MetaValues: []string{"34.56"}, Hit: Hit{Fingerprint: "fp4", Time: Today(), Path: "/", PageViews: 1}},
		{Name: "event2", DurationSeconds: 4, Hit: Hit{Fingerprint: "fp5", Time: Today(), Path: "/", PageViews: 1}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.Events(getMaxFilter())
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
	assert.Equal(t, 3, stats[0].Views)
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
	assert.Equal(t, 3, stats[0].Views)
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
	_, err = analyzer.EventBreakdown(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Referrer(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Referrer: "ref1/foo", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp1", Time: time.Now().Add(time.Minute), Path: "/foo", Referrer: "ref1/bar", ReferrerName: "Ref1", PageViews: 2, IsBounce: false},
		{Fingerprint: "fp1", Time: time.Now().Add(time.Minute * 2), Path: "/", Referrer: "ref2/foo", ReferrerName: "Ref2", PageViews: 3, IsBounce: false},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/", Referrer: "ref2/path", ReferrerName: "Ref2", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp2", Time: time.Now().Add(time.Minute), Path: "/bar", Referrer: "ref3/foo", ReferrerName: "Ref3", PageViews: 2, IsBounce: false},
		{Fingerprint: "fp3", Time: time.Now(), Path: "/", Referrer: "ref1/foo", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp4", Time: time.Now(), Path: "/", Referrer: "ref1/bar", ReferrerName: "Ref1", PageViews: 1, IsBounce: true},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Referrer(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "Ref2", visitors[1].ReferrerName)
	assert.Equal(t, "Ref3", visitors[2].ReferrerName)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.75, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.5, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.25, visitors[2].RelativeVisitors, 0.01)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.Equal(t, 0, visitors[2].Bounces)
	assert.InDelta(t, 0.6666, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
	_, err = analyzer.Referrer(getMaxFilter())
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
	assert.Equal(t, "ref1/foo", visitors[0].Referrer)
	assert.Equal(t, "ref1/bar", visitors[1].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[1].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[1].BounceRate, 0.01)

	// filter for full referrer
	visitors, err = analyzer.Referrer(&Filter{Referrer: "ref1/foo"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref1/foo", visitors[0].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)

	// filter for referrer name and full referrer
	visitors, err = analyzer.Referrer(&Filter{ReferrerName: "Ref1", Referrer: "ref1/foo"})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "Ref1", visitors[0].ReferrerName)
	assert.Equal(t, "ref1/foo", visitors[0].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.InDelta(t, 1, visitors[0].BounceRate, 0.01)
}

func TestAnalyzer_ReferrerUnknown(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Referrer: "ref1", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp1", Time: time.Now().Add(time.Minute), Path: "/foo", Referrer: "ref1", PageViews: 2, IsBounce: false},
		{Fingerprint: "fp1", Time: time.Now().Add(time.Minute * 2), Path: "/", PageViews: 3, IsBounce: false},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp2", Time: time.Now().Add(time.Minute), Path: "/bar", Referrer: "ref3", PageViews: 2, IsBounce: false},
		{Fingerprint: "fp3", Time: time.Now(), Path: "/", Referrer: "ref1", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp4", Time: time.Now(), Path: "/", Referrer: "ref1", PageViews: 1, IsBounce: true},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Referrer(&Filter{Referrer: Unknown})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Empty(t, visitors[0].Referrer)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.Equal(t, 1, visitors[0].Bounces)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)
}

func TestAnalyzer_Platform(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Desktop: true},
		{Fingerprint: "fp1", Time: time.Now(), Desktop: true},
		{Fingerprint: "fp1", Time: time.Now(), Mobile: true},
		{Fingerprint: "fp2", Time: time.Now(), Mobile: true},
		{Fingerprint: "fp2", Time: time.Now()},
		{Fingerprint: "fp3", Time: time.Now(), Desktop: true},
		{Fingerprint: "fp4", Time: time.Now(), Desktop: true},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	platform, err := analyzer.Platform(&Filter{From: pastDay(5), To: Today()})
	assert.NoError(t, err)
	assert.Equal(t, 3, platform.PlatformDesktop)
	assert.Equal(t, 2, platform.PlatformMobile)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0.5, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0.1666, platform.RelativePlatformUnknown, 0.01)
	_, err = analyzer.Platform(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Languages(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Language: "en"},
		{Fingerprint: "fp1", Time: time.Now(), Language: "en"},
		{Fingerprint: "fp2", Time: time.Now(), Language: "de"},
		{Fingerprint: "fp3", Time: time.Now(), Language: "de"},
		{Fingerprint: "fp4", Time: time.Now(), Language: "jp"},
		{Fingerprint: "fp5", Time: time.Now(), Language: "en"},
		{Fingerprint: "fp6", Time: time.Now(), Language: "en"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.Languages(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Countries(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: "en"},
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: "en"},
		{Fingerprint: "fp2", Time: time.Now(), CountryCode: "de"},
		{Fingerprint: "fp3", Time: time.Now(), CountryCode: "de"},
		{Fingerprint: "fp4", Time: time.Now(), CountryCode: "jp"},
		{Fingerprint: "fp5", Time: time.Now(), CountryCode: "en"},
		{Fingerprint: "fp6", Time: time.Now(), CountryCode: "en"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.Countries(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Cities(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), City: "London"},
		{Fingerprint: "fp1", Time: time.Now(), City: "London"},
		{Fingerprint: "fp2", Time: time.Now(), City: "Berlin"},
		{Fingerprint: "fp3", Time: time.Now(), City: "Berlin"},
		{Fingerprint: "fp4", Time: time.Now(), City: "Tokyo"},
		{Fingerprint: "fp5", Time: time.Now(), City: "London"},
		{Fingerprint: "fp6", Time: time.Now(), City: "London"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Cities(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "London", visitors[0].City)
	assert.Equal(t, "Berlin", visitors[1].City)
	assert.Equal(t, "Tokyo", visitors[2].City)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Cities(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Browser(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Browser: BrowserChrome},
		{Fingerprint: "fp1", Time: time.Now(), Browser: BrowserChrome},
		{Fingerprint: "fp2", Time: time.Now(), Browser: BrowserFirefox},
		{Fingerprint: "fp3", Time: time.Now(), Browser: BrowserFirefox},
		{Fingerprint: "fp4", Time: time.Now(), Browser: BrowserSafari},
		{Fingerprint: "fp5", Time: time.Now(), Browser: BrowserChrome},
		{Fingerprint: "fp6", Time: time.Now(), Browser: BrowserChrome},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.Browser(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_BrowserVersion(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "85.1"},
		{Fingerprint: "fp2", Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "85.1"},
		{Fingerprint: "fp2", Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "85.1"},
		{Fingerprint: "fp3", Time: time.Now(), Browser: BrowserFirefox, BrowserVersion: "89.0.0"},
		{Fingerprint: "fp4", Time: time.Now(), Browser: BrowserFirefox, BrowserVersion: "89.0.1"},
		{Fingerprint: "fp5", Time: time.Now(), Browser: BrowserSafari, BrowserVersion: "14.1.2"},
		{Fingerprint: "fp6", Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "87.2"},
		{Fingerprint: "fp7", Time: time.Now(), Browser: BrowserChrome, BrowserVersion: "86.0"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.BrowserVersion(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_OS(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), OS: OSWindows},
		{Fingerprint: "fp1", Time: time.Now(), OS: OSWindows},
		{Fingerprint: "fp2", Time: time.Now(), OS: OSMac},
		{Fingerprint: "fp3", Time: time.Now(), OS: OSMac},
		{Fingerprint: "fp4", Time: time.Now(), OS: OSLinux},
		{Fingerprint: "fp5", Time: time.Now(), OS: OSWindows},
		{Fingerprint: "fp6", Time: time.Now(), OS: OSWindows},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.OS(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, OSWindows, visitors[0].OS)
	assert.Equal(t, OSMac, visitors[1].OS)
	assert.Equal(t, OSLinux, visitors[2].OS)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.OS(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_OSVersion(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), OS: OSWindows, OSVersion: "10"},
		{Fingerprint: "fp2", Time: time.Now(), OS: OSWindows, OSVersion: "10"},
		{Fingerprint: "fp2", Time: time.Now(), OS: OSWindows, OSVersion: "10"},
		{Fingerprint: "fp3", Time: time.Now(), OS: OSMac, OSVersion: "14.0.0"},
		{Fingerprint: "fp4", Time: time.Now(), OS: OSMac, OSVersion: "13.1.0"},
		{Fingerprint: "fp5", Time: time.Now(), OS: OSLinux},
		{Fingerprint: "fp6", Time: time.Now(), OS: OSWindows, OSVersion: "9"},
		{Fingerprint: "fp7", Time: time.Now(), OS: OSWindows, OSVersion: "8"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.OSVersion(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_ScreenClass(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: "XXL"},
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: "XXL"},
		{Fingerprint: "fp2", Time: time.Now(), ScreenClass: "XL"},
		{Fingerprint: "fp3", Time: time.Now(), ScreenClass: "XL"},
		{Fingerprint: "fp4", Time: time.Now(), ScreenClass: "L"},
		{Fingerprint: "fp5", Time: time.Now(), ScreenClass: "XXL"},
		{Fingerprint: "fp6", Time: time.Now(), ScreenClass: "XXL"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.ScreenClass(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_UTM(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
		{Fingerprint: "fp1", Time: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
		{Fingerprint: "fp2", Time: time.Now(), UTMSource: "source2", UTMMedium: "medium2", UTMCampaign: "campaign2", UTMContent: "content2", UTMTerm: "term2"},
		{Fingerprint: "fp3", Time: time.Now(), UTMSource: "source2", UTMMedium: "medium2", UTMCampaign: "campaign2", UTMContent: "content2", UTMTerm: "term2"},
		{Fingerprint: "fp4", Time: time.Now(), UTMSource: "source3", UTMMedium: "medium3", UTMCampaign: "campaign3", UTMContent: "content3", UTMTerm: "term3"},
		{Fingerprint: "fp5", Time: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
		{Fingerprint: "fp6", Time: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	_, err = analyzer.UTMSource(getMaxFilter())
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
	_, err = analyzer.UTMMedium(getMaxFilter())
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
	_, err = analyzer.UTMCampaign(getMaxFilter())
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
	_, err = analyzer.UTMContent(getMaxFilter())
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
	_, err = analyzer.UTMTerm(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_AvgTimeOnPage(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(3), SessionID: 3, Path: "/", Title: "Home"},
		{Fingerprint: "fp1", Time: pastDay(3).Add(time.Second * 9), SessionID: 3, Path: "/foo", DurationSeconds: 9, Title: "Foo"},
		{Fingerprint: "fp2", Time: pastDay(3), SessionID: 3, Path: "/", Title: "Home"},
		{Fingerprint: "fp2", Time: pastDay(3).Add(time.Second * 7), SessionID: 3, Path: "/foo", DurationSeconds: 7, Title: "Foo"},
		{Fingerprint: "fp3", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{Fingerprint: "fp3", Time: pastDay(2).Add(time.Second * 5), SessionID: 2, Path: "/foo", DurationSeconds: 5, Title: "Foo"},
		{Fingerprint: "fp4", Time: pastDay(2), SessionID: 2, Path: "/", Title: "Home"},
		{Fingerprint: "fp4", Time: pastDay(2).Add(time.Second * 4), SessionID: 2, Path: "/foo", DurationSeconds: 4, Title: "Foo"},
		{Fingerprint: "fp5", Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{Fingerprint: "fp5", Time: pastDay(1).Add(time.Second * 8), SessionID: 1, Path: "/foo", DurationSeconds: 8, Title: "Foo"},
		{Fingerprint: "fp6", Time: pastDay(1), SessionID: 1, Path: "/", Title: "Home"},
		{Fingerprint: "fp6", Time: pastDay(1).Add(time.Second * 6), SessionID: 1, Path: "/foo", DurationSeconds: 6, Title: "Foo"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	byDay, err := analyzer.AvgTimeOnPage(&Filter{Path: "/", From: pastDay(3), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, byDay, 4)
	assert.Equal(t, 8, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 7, byDay[2].AverageTimeSpentSeconds)
	assert.Equal(t, 0, byDay[3].AverageTimeSpentSeconds)
	byDay, err = analyzer.AvgTimeOnPage(&Filter{MaxTimeOnPageSeconds: 5})
	assert.NoError(t, err)
	assert.Len(t, byDay, 3)
	assert.Equal(t, 5, byDay[0].AverageTimeSpentSeconds)
	assert.Equal(t, 4, byDay[1].AverageTimeSpentSeconds)
	assert.Equal(t, 5, byDay[2].AverageTimeSpentSeconds)
	byDay, err = analyzer.AvgTimeOnPage(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_CalculateGrowth(t *testing.T) {
	analyzer := NewAnalyzer(dbClient)
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
	analyzer := NewAnalyzer(dbClient)
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
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(3).Add(time.Hour * 18), Path: "/"}, // 18:00 UTC -> 03:00 Asia/Tokyo
		{Fingerprint: "fp2", Time: pastDay(2), Path: "/"},                     // 00:00 UTC -> 09:00 Asia/Tokyo
		{Fingerprint: "fp3", Time: pastDay(1).Add(time.Hour * 19), Path: "/"}, // 19:00 UTC -> 04:00 Asia/Tokyo
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: Today(), Path: "/"},
		{Fingerprint: "fp2", Time: Today(), Path: "/simple/page"},
		{Fingerprint: "fp3", Time: Today(), Path: "/siMple/page/"},
		{Fingerprint: "fp4", Time: Today(), Path: "/simple/page/with/many/slashes"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", SessionID: 1, Time: Today(), DurationSeconds: 0, Path: "/", EntryPath: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp1", SessionID: 1, Time: Today().Add(time.Second * 3), DurationSeconds: 3, Path: "/account/billing/", EntryPath: "/", PageViews: 2, IsBounce: false},
		{Fingerprint: "fp1", SessionID: 1, Time: Today().Add(time.Second * 5), DurationSeconds: 2, Path: "/settings/general/", EntryPath: "/", PageViews: 3, IsBounce: false},
		{Fingerprint: "fp1", SessionID: 1, Time: Today().Add(time.Second * 7), DurationSeconds: 2, Path: "/integrations/wordpress/", EntryPath: "/", PageViews: 4, IsBounce: false},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
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
	assert.NoError(t, dbClient.SaveHits([]Hit{
		// / -> /foo -> /bar -> /exit
		{Fingerprint: "fp1", SessionID: 1, Time: Today(), Path: "/", EntryPath: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp1", SessionID: 1, Time: Today().Add(time.Second * 10), Path: "/foo", EntryPath: "/", PageViews: 2, IsBounce: false},
		{Fingerprint: "fp1", SessionID: 1, Time: Today().Add(time.Second * 20), Path: "/bar", EntryPath: "/", PageViews: 3, IsBounce: false},
		{Fingerprint: "fp1", SessionID: 1, Time: Today().Add(time.Second * 30), Path: "/exit", EntryPath: "/", PageViews: 4, IsBounce: false},

		// / -> /bar -> /
		{Fingerprint: "fp2", SessionID: 2, Time: Today(), Path: "/", EntryPath: "/", PageViews: 1, IsBounce: true},
		{Fingerprint: "fp2", SessionID: 2, Time: Today().Add(time.Second * 10), Path: "/bar", EntryPath: "/", PageViews: 2, IsBounce: false},
		{Fingerprint: "fp2", SessionID: 2, Time: Today().Add(time.Second * 20), Path: "/", EntryPath: "/", PageViews: 3, IsBounce: false},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)

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

func getMaxFilter() *Filter {
	return &Filter{
		ClientID:       42,
		From:           pastDay(5),
		To:             pastDay(2),
		Day:            pastDay(1),
		Start:          time.Now().UTC(),
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
		Limit:          42,
	}
}
