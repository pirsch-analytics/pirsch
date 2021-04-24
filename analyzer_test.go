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
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Language: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Language: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/bar", Language: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/bar", Language: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/foo", Language: sql.NullString{String: "jp", Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), Path: "/", Language: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), Path: "/", Language: sql.NullString{String: "en", Valid: true}},
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
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", CountryCode: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", CountryCode: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/bar", CountryCode: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/bar", CountryCode: sql.NullString{String: "de", Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/foo", CountryCode: sql.NullString{String: "jp", Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), Path: "/", CountryCode: sql.NullString{String: "en", Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), Path: "/", CountryCode: sql.NullString{String: "en", Valid: true}},
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
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Browser: sql.NullString{String: BrowserChrome, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", Browser: sql.NullString{String: BrowserChrome, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/bar", Browser: sql.NullString{String: BrowserFirefox, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/bar", Browser: sql.NullString{String: BrowserFirefox, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/foo", Browser: sql.NullString{String: BrowserSafari, Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), Path: "/", Browser: sql.NullString{String: BrowserChrome, Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), Path: "/", Browser: sql.NullString{String: BrowserChrome, Valid: true}},
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
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", OS: sql.NullString{String: OSWindows, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/", OS: sql.NullString{String: OSWindows, Valid: true}},
		{Fingerprint: "fp1", Time: time.Now(), Path: "/bar", OS: sql.NullString{String: OSMac, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/bar", OS: sql.NullString{String: OSMac, Valid: true}},
		{Fingerprint: "fp2", Time: time.Now(), Path: "/foo", OS: sql.NullString{String: OSLinux, Valid: true}},
		{Fingerprint: "fp3", Time: time.Now(), Path: "/", OS: sql.NullString{String: OSWindows, Valid: true}},
		{Fingerprint: "fp4", Time: time.Now(), Path: "/", OS: sql.NullString{String: OSWindows, Valid: true}},
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
