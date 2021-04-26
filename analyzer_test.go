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

func TestAnalyzer_Visitors(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/"},
		{Fingerprint: "fp1", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/foo"},
		{Fingerprint: "fp1", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/bar"},
		{Fingerprint: "fp3", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp4", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2).Add(time.Minute * 30), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp6", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp7", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp8", Time: Today(), Path: "/"},
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
	assert.Equal(t, 3, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 5, visitors[0].Sessions)
	assert.Equal(t, 4, visitors[1].Sessions)
	assert.Equal(t, 1, visitors[2].Sessions)
	assert.Equal(t, 7, visitors[0].Views)
	assert.Equal(t, 4, visitors[1].Views)
	assert.Equal(t, 1, visitors[2].Views)
	assert.Equal(t, 2, visitors[0].Bounces)
	assert.Equal(t, 2, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.InDelta(t, 0.5, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 0.6666, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[2].BounceRate, 0.01)
	visitors, err = analyzer.Visitors(&Filter{From: pastDay(4), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, pastDay(4), visitors[0].Day)
	assert.Equal(t, pastDay(2), visitors[1].Day)
}

func TestAnalyzer_Pages(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/"},
		{Fingerprint: "fp1", Time: pastDay(4), Session: sql.NullTime{Time: pastDay(4), Valid: true}, Path: "/foo"},
		{Fingerprint: "fp1", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp2", Time: pastDay(4), Path: "/bar"},
		{Fingerprint: "fp3", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp4", Time: pastDay(4), Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2), Valid: true}, Path: "/"},
		{Fingerprint: "fp5", Time: pastDay(2), Session: sql.NullTime{Time: pastDay(2).Add(time.Minute * 30), Valid: true}, Path: "/bar"},
		{Fingerprint: "fp6", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp7", Time: pastDay(2), Path: "/"},
		{Fingerprint: "fp8", Time: Today(), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Pages(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path.String)
	assert.Equal(t, "/bar", visitors[1].Path.String)
	assert.Equal(t, "/foo", visitors[2].Path.String)
	assert.Equal(t, 8, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 9, visitors[0].Sessions)
	assert.Equal(t, 2, visitors[1].Sessions)
	assert.Equal(t, 1, visitors[2].Sessions)
	assert.Equal(t, 9, visitors[0].Views)
	assert.Equal(t, 2, visitors[1].Views)
	assert.Equal(t, 1, visitors[2].Views)
	assert.Equal(t, 7, visitors[0].Bounces)
	assert.Equal(t, 2, visitors[1].Bounces)
	assert.Equal(t, 1, visitors[2].Bounces)
	assert.InDelta(t, 0.875, visitors[0].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[1].BounceRate, 0.01)
	assert.InDelta(t, 1, visitors[2].BounceRate, 0.01)
	visitors, err = analyzer.Pages(&Filter{From: pastDay(3), To: pastDay(1)})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "/", visitors[0].Path.String)
	assert.Equal(t, "/bar", visitors[1].Path.String)
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
		{Fingerprint: "fp1", Time: time.Now(), Language: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Language: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Language: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Language: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Language: sql.NullString{String: "jp", Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), Language: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), Language: sql.NullString{String: "en", Valid: true}},
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
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), CountryCode: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), CountryCode: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), CountryCode: sql.NullString{String: "jp", Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), CountryCode: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), CountryCode: sql.NullString{String: "en", Valid: true}},
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
		{Fingerprint: "fp1", Time: time.Now(), Browser: sql.NullString{String: BrowserChrome, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Browser: sql.NullString{String: BrowserChrome, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Browser: sql.NullString{String: BrowserFirefox, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Browser: sql.NullString{String: BrowserFirefox, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Browser: sql.NullString{String: BrowserSafari, Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), Browser: sql.NullString{String: BrowserChrome, Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), Browser: sql.NullString{String: BrowserChrome, Valid: true}},
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
		{Fingerprint: "fp1", Time: time.Now(), OS: sql.NullString{String: OSWindows, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), OS: sql.NullString{String: OSWindows, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), OS: sql.NullString{String: OSMac, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), OS: sql.NullString{String: OSMac, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), OS: sql.NullString{String: OSLinux, Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), OS: sql.NullString{String: OSWindows, Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), OS: sql.NullString{String: OSWindows, Valid: true}},
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
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: sql.NullString{String: "XXL", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: sql.NullString{String: "XXL", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), ScreenClass: sql.NullString{String: "XL", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), ScreenClass: sql.NullString{String: "XL", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), ScreenClass: sql.NullString{String: "L", Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), ScreenClass: sql.NullString{String: "XXL", Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), ScreenClass: sql.NullString{String: "XXL", Valid: true}},
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
