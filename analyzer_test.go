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
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Languages(&Filter{Day: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].Language.String)
	assert.Equal(t, "de", visitors[1].Language.String)
	assert.Equal(t, "jp", visitors[2].Language.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
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
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Countries(&Filter{Day: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].CountryCode.String)
	assert.Equal(t, "de", visitors[1].CountryCode.String)
	assert.Equal(t, "jp", visitors[2].CountryCode.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
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
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Browser(&Filter{Day: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, BrowserChrome, visitors[0].Browser.String)
	assert.Equal(t, BrowserFirefox, visitors[1].Browser.String)
	assert.Equal(t, BrowserSafari, visitors[2].Browser.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
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
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.OS(&Filter{Day: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, OSWindows, visitors[0].OS.String)
	assert.Equal(t, OSMac, visitors[1].OS.String)
	assert.Equal(t, OSLinux, visitors[2].OS.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
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

func TestAnalyzer_ScreenSize(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now(), ScreenWidth: 1920, ScreenHeight: 1080},
		{Fingerprint: "fp1", Time: time.Now(), ScreenWidth: 1920, ScreenHeight: 1080},
		{Fingerprint: "fp1", Time: time.Now(), ScreenWidth: 1280, ScreenHeight: 720},
		{Fingerprint: "fp2", Time: time.Now(), ScreenWidth: 1280, ScreenHeight: 720},
		{Fingerprint: "fp2", Time: time.Now(), ScreenWidth: 640, ScreenHeight: 720},
		{Fingerprint: "fp3", Time: time.Now(), ScreenWidth: 1920, ScreenHeight: 1080},
		{Fingerprint: "fp4", Time: time.Now(), ScreenWidth: 1920, ScreenHeight: 1080},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.ScreenSize(&Filter{Day: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, 1920, visitors[0].ScreenWidth)
	assert.Equal(t, 1080, visitors[0].ScreenHeight)
	assert.Equal(t, 1280, visitors[1].ScreenWidth)
	assert.Equal(t, 720, visitors[1].ScreenHeight)
	assert.Equal(t, 640, visitors[2].ScreenWidth)
	assert.Equal(t, 720, visitors[2].ScreenHeight)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
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
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.ScreenClass(&Filter{Day: Today()})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "XXL", visitors[0].ScreenClass.String)
	assert.Equal(t, "XL", visitors[1].ScreenClass.String)
	assert.Equal(t, "L", visitors[2].ScreenClass.String)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
}
