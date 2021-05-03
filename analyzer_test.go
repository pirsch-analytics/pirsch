package pirsch

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_ActiveVisitors(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 30), Path: "/"},
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 15), Path: "/"},
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 5), Path: "/bar"},
		{Fingerprint: "fp2", Time: time.Now().Add(-time.Minute * 4), Path: "/bar"},
		{Fingerprint: "fp2", Time: time.Now().Add(-time.Minute * 3), Path: "/foo"},
		{Fingerprint: "fp3", Time: time.Now().Add(-time.Minute * 3), Path: "/"},
		{Fingerprint: "fp4", Time: time.Now().Add(-time.Minute), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, count, err := analyzer.ActiveVisitors(nil, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path.String)
	assert.Equal(t, "/bar", visitors[1].Path.String)
	assert.Equal(t, "/foo", visitors[2].Path.String)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	visitors, count, err = analyzer.ActiveVisitors(&Filter{Path: "/bar"}, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "/bar", visitors[0].Path.String)
	assert.Equal(t, 2, visitors[0].Visitors)
	_, _, err = analyzer.ActiveVisitors(getMaxFilter(), time.Minute*10)
	assert.NoError(t, err)
}

func TestAnalyzer_VisitorsAndAvgSessionDuration(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/"},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 5), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/foo"},
		{Fingerprint: "fp1", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4).Add(time.Minute * 10), Session: sql.NullTime{Time: pastDay(4).Add(time.Minute * 30), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp3", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp4", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2).Add(time.Minute * 5), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp6", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/"},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 10), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp7", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp8", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp9", Time: Today(), Path: "/"},
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
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 0, visitors[1].Bounces)
	assert.Equal(t, 2, visitors[2].Bounces)
	assert.Equal(t, 0, visitors[3].Bounces)
	assert.Equal(t, 1, visitors[4].Bounces)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)
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
	tsd, err := analyzer.TotalSessionDuration(nil)
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
	tsd, err = analyzer.TotalSessionDuration(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 900, tsd)
	_, err = analyzer.Visitors(getMaxFilter())
	assert.NoError(t, err)
	_, err = analyzer.AvgSessionDuration(getMaxFilter())
	assert.NoError(t, err)
	_, err = analyzer.TotalSessionDuration(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Growth(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/"},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 15), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/bar", PreviousTimeOnPageSeconds: 900},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp3", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp4", Time: pastDay(3), Session: sql.NullTime{Time: pastDay(3), Valid: true}, Path: "/"},
		{Fingerprint: "fp4", Time: pastDay(3).Add(time.Minute * 5), Session: sql.NullTime{Time: pastDay(3), Valid: true}, Path: "/foo", PreviousTimeOnPageSeconds: 300},
		{Fingerprint: "fp4", Time: pastDay(3), Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(3), Session: sql.NullTime{Time: pastDay(3), Valid: true}, Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(3).Add(time.Minute * 10), Session: sql.NullTime{Time: pastDay(3).Add(time.Minute * 30), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp6", Time: pastDay(3), Path: "/"},
		{Fingerprint: "fp7", Time: pastDay(3), Path: "/"},
		{Fingerprint: "fp8", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/"},
		{Fingerprint: "fp8", Time: pastDay(2).Add(time.Minute * 5), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/bar", PreviousTimeOnPageSeconds: 300},
		{Fingerprint: "fp9", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp10", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp11", Time: Today(), Path: "/"},
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
	assert.InDelta(t, 0, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, 0, growth.TimeSpentGrowth, 0.001)
	growth, err = analyzer.Growth(&Filter{From: pastDay(3), To: pastDay(2)})
	assert.NoError(t, err)
	assert.NotNil(t, growth)
	assert.InDelta(t, 1.3333, growth.VisitorsGrowth, 0.001)
	assert.InDelta(t, 1.75, growth.ViewsGrowth, 0.001)
	assert.InDelta(t, 2, growth.SessionsGrowth, 0.001)
	assert.InDelta(t, 1, growth.BouncesGrowth, 0.001)
	assert.InDelta(t, -0.3333, growth.TimeSpentGrowth, 0.001)
	_, err = analyzer.Growth(getMaxFilter())
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
		{Fingerprint: "fp1", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/"},
		{Fingerprint: "fp1", Time: pastDay(4).Add(time.Minute * 3), Session: sql.NullTime{Time: pastDay(4), Valid: true}, PreviousTimeOnPageSeconds: 180, Path: "/foo"},
		{Fingerprint: "fp1", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/bar"},
		{Fingerprint: "fp3", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp4", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2).Add(time.Minute * 5), Session: sql.NullTime{Time: pastDay(2).Add(time.Minute * 30), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp6", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/"},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 10), Session: sql.NullTime{Time: pastDay(2), Valid: true}, PreviousTimeOnPageSeconds: 600, Path: "/bar"},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 11), Session: sql.NullTime{Time: pastDay(2).Add(time.Hour), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp6", Time: pastDay(2).Add(time.Minute * 21), Session: sql.NullTime{Time: pastDay(2).Add(time.Hour), Valid: true}, PreviousTimeOnPageSeconds: 600, Path: "/foo"},
		{Fingerprint: "fp7", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp8", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp9", Time: Today(), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Pages(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path.String)
	assert.Equal(t, "/bar", visitors[1].Path.String)
	assert.Equal(t, "/foo", visitors[2].Path.String)
	assert.Equal(t, 9, visitors[0].Visitors)
	assert.Equal(t, 3, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.InDelta(t, 0.6428, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2142, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1428, visitors[2].RelativeVisitors, 0.01)
	assert.Equal(t, 10, visitors[0].Sessions)
	assert.Equal(t, 4, visitors[1].Sessions)
	assert.Equal(t, 2, visitors[2].Sessions)
	assert.Equal(t, 10, visitors[0].Views)
	assert.Equal(t, 4, visitors[1].Views)
	assert.Equal(t, 2, visitors[2].Views)
	assert.InDelta(t, 0.625, visitors[0].RelativeViews, 0.01)
	assert.InDelta(t, 0.25, visitors[1].RelativeViews, 0.01)
	assert.InDelta(t, 0.125, visitors[2].RelativeViews, 0.01)
	assert.Equal(t, 8, visitors[0].Bounces)
	assert.Equal(t, 2, visitors[1].Bounces)
	assert.Equal(t, 2, visitors[2].Bounces)
	assert.InDelta(t, 0.8888, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.6666, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[2].BounceRate, 0.01)
	atop, err := analyzer.AvgTimeOnPages(nil)
	assert.NoError(t, err)
	assert.Len(t, atop, 2)
	assert.Equal(t, "/", atop[0].Path.String)
	assert.Equal(t, "/bar", atop[1].Path.String)
	assert.Equal(t, 390, atop[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, atop[1].AverageTimeSpentSeconds)
	top, err := analyzer.AvgTimeOnPage(nil)
	assert.NoError(t, err)
	assert.Len(t, top, 2)
	assert.Equal(t, pastDay(4), top[0].Day)
	assert.Equal(t, pastDay(2), top[1].Day)
	assert.Equal(t, 180, top[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, top[1].AverageTimeSpentSeconds)
	ttop, err := analyzer.TotalTimeOnPage(nil)
	assert.NoError(t, err)
	assert.Equal(t, 1380, ttop)
	visitors, err = analyzer.Pages(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path.String)
	assert.Equal(t, "/bar", visitors[1].Path.String)
	atop, err = analyzer.AvgTimeOnPages(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, atop, 2)
	assert.Equal(t, "/", atop[0].Path.String)
	assert.Equal(t, "/bar", atop[1].Path.String)
	assert.Equal(t, 600, atop[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, atop[1].AverageTimeSpentSeconds)
	top, err = analyzer.AvgTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, top, 3)
	assert.Equal(t, pastDay(3), top[0].Day)
	assert.Equal(t, pastDay(2), top[1].Day)
	assert.Equal(t, pastDay(1), top[2].Day)
	assert.Equal(t, 0, top[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, top[1].AverageTimeSpentSeconds)
	assert.Equal(t, 0, top[2].AverageTimeSpentSeconds)
	ttop, err = analyzer.TotalTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 1200, ttop)
	_, err = analyzer.Pages(getMaxFilter())
	assert.NoError(t, err)
	_, err = analyzer.AvgTimeOnPages(getMaxFilter())
	assert.NoError(t, err)
	_, err = analyzer.TotalTimeOnPage(getMaxFilter())
	assert.NoError(t, err)
	visitors, err = analyzer.Pages(&Filter{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
}

func TestAnalyzer_Referrer(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Referrer: sql.NullString{String: "ref1", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/foo", Referrer: sql.NullString{String: "ref1", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Referrer: sql.NullString{String: "ref2", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/", Referrer: sql.NullString{String: "ref2", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/bar", Referrer: sql.NullString{String: "ref3", Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), Path: "/", Referrer: sql.NullString{String: "ref1", Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), Path: "/", Referrer: sql.NullString{String: "ref1", Valid: true}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Referrer(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "ref1", visitors[0].Referrer.String)
	assert.Equal(t, "ref2", visitors[1].Referrer.String)
	assert.Equal(t, "ref3", visitors[2].Referrer.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 2, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.InDelta(t, 0.6666, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[2].BounceRate, 0.01)
	_, err = analyzer.Referrer(getMaxFilter())
	assert.NoError(t, err)
	visitors, err = analyzer.Referrer(&Filter{Limit: 1})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
}

func TestAnalyzer_Languages(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Language: "en"},
		{Fingerprint: "fp1", Time: time.Now(), Language: "en"},
		{Fingerprint: "fp1", Time: time.Now(), Language: "de"},
		{Fingerprint: "fp2", Time: time.Now(), Language: "de"},
		{Fingerprint: "fp2", Time: time.Now(), Language: "jp"},
		{Fingerprint: "fp3", Time: time.Now(), Language: "en"},
		{Fingerprint: "fp4", Time: time.Now(), Language: "en"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Languages(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].Language.String)
	assert.Equal(t, "de", visitors[1].Language.String)
	assert.Equal(t, "jp", visitors[2].Language.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Languages(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Countries(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: "en"},
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: "en"},
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: "de"},
		{Fingerprint: "fp2", Time: time.Now(), CountryCode: "de"},
		{Fingerprint: "fp2", Time: time.Now(), CountryCode: "jp"},
		{Fingerprint: "fp3", Time: time.Now(), CountryCode: "en"},
		{Fingerprint: "fp4", Time: time.Now(), CountryCode: "en"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Countries(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].CountryCode.String)
	assert.Equal(t, "de", visitors[1].CountryCode.String)
	assert.Equal(t, "jp", visitors[2].CountryCode.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Countries(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_Browser(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), Browser: BrowserChrome},
		{Fingerprint: "fp1", Time: time.Now(), Browser: BrowserChrome},
		{Fingerprint: "fp1", Time: time.Now(), Browser: BrowserFirefox},
		{Fingerprint: "fp2", Time: time.Now(), Browser: BrowserFirefox},
		{Fingerprint: "fp2", Time: time.Now(), Browser: BrowserSafari},
		{Fingerprint: "fp3", Time: time.Now(), Browser: BrowserChrome},
		{Fingerprint: "fp4", Time: time.Now(), Browser: BrowserChrome},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Browser(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, BrowserChrome, visitors[0].Browser.String)
	assert.Equal(t, BrowserFirefox, visitors[1].Browser.String)
	assert.Equal(t, BrowserSafari, visitors[2].Browser.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Browser(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_OS(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), OS: OSWindows},
		{Fingerprint: "fp1", Time: time.Now(), OS: OSWindows},
		{Fingerprint: "fp1", Time: time.Now(), OS: OSMac},
		{Fingerprint: "fp2", Time: time.Now(), OS: OSMac},
		{Fingerprint: "fp2", Time: time.Now(), OS: OSLinux},
		{Fingerprint: "fp3", Time: time.Now(), OS: OSWindows},
		{Fingerprint: "fp4", Time: time.Now(), OS: OSWindows},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.OS(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, OSWindows, visitors[0].OS.String)
	assert.Equal(t, OSMac, visitors[1].OS.String)
	assert.Equal(t, OSLinux, visitors[2].OS.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.OS(getMaxFilter())
	assert.NoError(t, err)
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

func TestAnalyzer_ScreenClass(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: "XXL"},
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: "XXL"},
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: "XL"},
		{Fingerprint: "fp2", Time: time.Now(), ScreenClass: "XL"},
		{Fingerprint: "fp2", Time: time.Now(), ScreenClass: "L"},
		{Fingerprint: "fp3", Time: time.Now(), ScreenClass: "XXL"},
		{Fingerprint: "fp4", Time: time.Now(), ScreenClass: "XXL"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.ScreenClass(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "XXL", visitors[0].ScreenClass.String)
	assert.Equal(t, "XL", visitors[1].ScreenClass.String)
	assert.Equal(t, "L", visitors[2].ScreenClass.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.ScreenClass(getMaxFilter())
	assert.NoError(t, err)
}

func TestAnalyzer_UTM(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), UTMSource: sql.NullString{String: "source1", Valid: true}, UTMMedium: sql.NullString{String: "medium1", Valid: true}, UTMCampaign: sql.NullString{String: "campaign1", Valid: true}, UTMContent: sql.NullString{String: "content1", Valid: true}, UTMTerm: sql.NullString{String: "term1", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), UTMSource: sql.NullString{String: "source1", Valid: true}, UTMMedium: sql.NullString{String: "medium1", Valid: true}, UTMCampaign: sql.NullString{String: "campaign1", Valid: true}, UTMContent: sql.NullString{String: "content1", Valid: true}, UTMTerm: sql.NullString{String: "term1", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), UTMSource: sql.NullString{String: "source2", Valid: true}, UTMMedium: sql.NullString{String: "medium2", Valid: true}, UTMCampaign: sql.NullString{String: "campaign2", Valid: true}, UTMContent: sql.NullString{String: "content2", Valid: true}, UTMTerm: sql.NullString{String: "term2", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), UTMSource: sql.NullString{String: "source2", Valid: true}, UTMMedium: sql.NullString{String: "medium2", Valid: true}, UTMCampaign: sql.NullString{String: "campaign2", Valid: true}, UTMContent: sql.NullString{String: "content2", Valid: true}, UTMTerm: sql.NullString{String: "term2", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), UTMSource: sql.NullString{String: "source3", Valid: true}, UTMMedium: sql.NullString{String: "medium3", Valid: true}, UTMCampaign: sql.NullString{String: "campaign3", Valid: true}, UTMContent: sql.NullString{String: "content3", Valid: true}, UTMTerm: sql.NullString{String: "term3", Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), UTMSource: sql.NullString{String: "source1", Valid: true}, UTMMedium: sql.NullString{String: "medium1", Valid: true}, UTMCampaign: sql.NullString{String: "campaign1", Valid: true}, UTMContent: sql.NullString{String: "content1", Valid: true}, UTMTerm: sql.NullString{String: "term1", Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), UTMSource: sql.NullString{String: "source1", Valid: true}, UTMMedium: sql.NullString{String: "medium1", Valid: true}, UTMCampaign: sql.NullString{String: "campaign1", Valid: true}, UTMContent: sql.NullString{String: "content1", Valid: true}, UTMTerm: sql.NullString{String: "term1", Valid: true}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	source, err := analyzer.UTMSource(nil)
	assert.NoError(t, err)
	assert.Len(t, source, 3)
	assert.Equal(t, "source1", source[0].UTMSource.String)
	assert.Equal(t, "source2", source[1].UTMSource.String)
	assert.Equal(t, "source3", source[2].UTMSource.String)
	assert.Equal(t, 3, source[0].Visitors)
	assert.Equal(t, 2, source[1].Visitors)
	assert.Equal(t, 1, source[2].Visitors)
	assert.InDelta(t, 0.5, source[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, source[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, source[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMSource(getMaxFilter())
	assert.NoError(t, err)
	medium, err := analyzer.UTMMedium(nil)
	assert.NoError(t, err)
	assert.Len(t, medium, 3)
	assert.Equal(t, "medium1", medium[0].UTMMedium.String)
	assert.Equal(t, "medium2", medium[1].UTMMedium.String)
	assert.Equal(t, "medium3", medium[2].UTMMedium.String)
	assert.Equal(t, 3, medium[0].Visitors)
	assert.Equal(t, 2, medium[1].Visitors)
	assert.Equal(t, 1, medium[2].Visitors)
	assert.InDelta(t, 0.5, medium[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, medium[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, medium[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMMedium(getMaxFilter())
	assert.NoError(t, err)
	campaign, err := analyzer.UTMCampaign(nil)
	assert.NoError(t, err)
	assert.Len(t, campaign, 3)
	assert.Equal(t, "campaign1", campaign[0].UTMCampaign.String)
	assert.Equal(t, "campaign2", campaign[1].UTMCampaign.String)
	assert.Equal(t, "campaign3", campaign[2].UTMCampaign.String)
	assert.Equal(t, 3, campaign[0].Visitors)
	assert.Equal(t, 2, campaign[1].Visitors)
	assert.Equal(t, 1, campaign[2].Visitors)
	assert.InDelta(t, 0.5, campaign[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, campaign[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, campaign[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMCampaign(getMaxFilter())
	assert.NoError(t, err)
	content, err := analyzer.UTMContent(nil)
	assert.NoError(t, err)
	assert.Len(t, content, 3)
	assert.Equal(t, "content1", content[0].UTMContent.String)
	assert.Equal(t, "content2", content[1].UTMContent.String)
	assert.Equal(t, "content3", content[2].UTMContent.String)
	assert.Equal(t, 3, content[0].Visitors)
	assert.Equal(t, 2, content[1].Visitors)
	assert.Equal(t, 1, content[2].Visitors)
	assert.InDelta(t, 0.5, content[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, content[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, content[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMContent(getMaxFilter())
	assert.NoError(t, err)
	term, err := analyzer.UTMTerm(nil)
	assert.NoError(t, err)
	assert.Len(t, term, 3)
	assert.Equal(t, "term1", term[0].UTMTerm.String)
	assert.Equal(t, "term2", term[1].UTMTerm.String)
	assert.Equal(t, "term3", term[2].UTMTerm.String)
	assert.Equal(t, 3, term[0].Visitors)
	assert.Equal(t, 2, term[1].Visitors)
	assert.Equal(t, 1, term[2].Visitors)
	assert.InDelta(t, 0.5, term[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, term[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, term[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTMTerm(getMaxFilter())
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

func getMaxFilter() *Filter {
	return &Filter{
		ClientID:       42,
		From:           pastDay(5),
		To:             pastDay(2),
		Day:            pastDay(1),
		Start:          time.Now().UTC(),
		Path:           "/path",
		Language:       "en",
		Country:        "en",
		Referrer:       "ref",
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
