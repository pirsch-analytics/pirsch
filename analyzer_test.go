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
	visitors, err := analyzer.Visitors(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, pastDay(4), visitors[0].Day)
	assert.Equal(t, pastDay(2), visitors[1].Day)
	assert.Equal(t, Today(), visitors[2].Day)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 4, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 6, visitors[0].Sessions)
	assert.Equal(t, 4, visitors[1].Sessions)
	assert.Equal(t, 1, visitors[2].Sessions)
	assert.Equal(t, 7, visitors[0].Views)
	assert.Equal(t, 6, visitors[1].Views)
	assert.Equal(t, 1, visitors[2].Views)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 2, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.5, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[2].BounceRate, 0.01)
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
	assert.Len(t, visitors, 2)
	assert.Equal(t, pastDay(4), visitors[0].Day)
	assert.Equal(t, pastDay(2), visitors[1].Day)
	asd, err = analyzer.AvgSessionDuration(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, asd, 1)
	tsd, err = analyzer.TotalSessionDuration(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 900, tsd)
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
	assert.Len(t, visitors, 6)
	assert.Equal(t, 0, visitors[0].Hour)
	assert.Equal(t, 3, visitors[1].Hour)
	assert.Equal(t, 4, visitors[2].Hour)
	assert.Equal(t, 5, visitors[3].Hour)
	assert.Equal(t, 8, visitors[4].Hour)
	assert.Equal(t, 10, visitors[5].Hour)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 3, visitors[3].Visitors)
	assert.Equal(t, 2, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	visitors, err = analyzer.VisitorHours(&Filter{From: pastDay(1), To: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 5)
	assert.Equal(t, 3, visitors[0].Hour)
	assert.Equal(t, 4, visitors[1].Hour)
	assert.Equal(t, 5, visitors[2].Hour)
	assert.Equal(t, 8, visitors[3].Hour)
	assert.Equal(t, 10, visitors[4].Hour)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
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
	atop, err := analyzer.AvgTimeOnPage(nil)
	assert.NoError(t, err)
	assert.Len(t, atop, 2)
	assert.Equal(t, "/", atop[0].Path.String)
	assert.Equal(t, "/bar", atop[1].Path.String)
	assert.Equal(t, 390, atop[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, atop[1].AverageTimeSpentSeconds)
	ttop, err := analyzer.TotalTimeOnPage(nil)
	assert.NoError(t, err)
	assert.Equal(t, 1380, ttop)
	visitors, err = analyzer.Pages(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path.String)
	assert.Equal(t, "/bar", visitors[1].Path.String)
	atop, err = analyzer.AvgTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, atop, 2)
	assert.Equal(t, "/", atop[0].Path.String)
	assert.Equal(t, "/bar", atop[1].Path.String)
	assert.Equal(t, 600, atop[0].AverageTimeSpentSeconds)
	assert.Equal(t, 600, atop[1].AverageTimeSpentSeconds)
	ttop, err = analyzer.TotalTimeOnPage(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Equal(t, 1200, ttop)
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
