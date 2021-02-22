package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProcessor_Process(t *testing.T) {
	testProcess(t, 0)
}

func TestProcessor_ProcessTenant(t *testing.T) {
	testProcess(t, 1)
}

func TestProcessor_ProcessSessions(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)

	// create hits for two visitors and three sessions
	now := time.Now()
	createHit(t, store, 0, "fp1", "/", "en", "", "", day(2020, 9, 7, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "", "", day(2020, 9, 7, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp2", "/", "en", "", "", day(2020, 9, 7, 5), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp2", "/", "en", "", "", day(2020, 9, 7, 5), now.Add(time.Second*1), OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	processor := NewProcessor(store)
	assert.NoError(t, processor.Process())
	checkHits(t, 0)
	db := sqlx.NewDb(postgresDB, "postgres")
	var visitorStats []VisitorStats
	var timeStats []VisitorTimeStats
	assert.NoError(t, db.Select(&visitorStats, `SELECT * FROM "visitor_stats" ORDER BY "day", "path"`))
	assert.NoError(t, db.Select(&timeStats, `SELECT * FROM "visitor_time_stats" ORDER BY "day", "hour"`))
	assert.Len(t, visitorStats, 2)
	assert.Len(t, timeStats, 24)
	assert.Equal(t, 2, visitorStats[0].Visitors)
	assert.Equal(t, 3, visitorStats[0].Sessions)
	assert.Equal(t, 1, timeStats[4].Visitors)
	assert.Equal(t, 1, timeStats[5].Visitors)
}

func TestProcessor_ProcessPaths(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)
	now := time.Now()
	createHit(t, store, 0, "fp1", "/", "en", "", "", day(2020, 12, 27, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "", "", day(2020, 12, 27, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "", "", day(2020, 12, 27, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "", "", day(2020, 12, 27, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/path", "en", "", "", day(2020, 12, 27, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/path", "en", "", "", day(2020, 12, 27, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/path", "en", "", "", day(2020, 12, 27, 4), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp2", "/path", "en", "", "", day(2020, 12, 27, 5), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp2", "/path", "en", "", "", day(2020, 12, 27, 5), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	createHit(t, store, 0, "fp2", "/path", "en", "", "", day(2020, 12, 27, 5), now, OSWindows, "10", BrowserChrome, "84.0", "", true, false, 0, 0)
	processor := NewProcessor(store)
	assert.NoError(t, processor.Process())
	analyzer := NewAnalyzer(store, nil)
	pageVisitors, err := analyzer.PageVisitors(&Filter{
		From: day(2020, 12, 25, 0),
		To:   day(2020, 12, 28, 0),
	})
	assert.NoError(t, err)
	assert.Len(t, pageVisitors, 2)
	assert.Equal(t, "/path", pageVisitors[0].Path)
	assert.Equal(t, 2, pageVisitors[0].Visitors)
	assert.Equal(t, "/", pageVisitors[1].Path)
	assert.Equal(t, 1, pageVisitors[1].Visitors)
	visitors, err := analyzer.Visitors(&Filter{
		From: day(2020, 12, 24, 0),
		To:   day(2020, 21, 31, 0),
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, sumUpVisitors(visitors))
}

func TestProcessor_ProcessReferrerBounces(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)
	createHit(t, store, 0, "fp1", "/", "en", "ua1", "ref1", day(2020, 12, 21, 7), time.Time{}, OSWindows, "10", BrowserChrome, "84.0", "de", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/second-page", "en", "ua1", "ref1", day(2020, 12, 21, 7), time.Time{}, OSWindows, "10", BrowserChrome, "84.0", "de", true, false, 0, 0)
	createHit(t, store, 0, "fp2", "/second-page", "en", "ua1", "ref1", day(2020, 12, 21, 7), time.Time{}, OSWindows, "10", BrowserChrome, "84.0", "de", true, false, 0, 0)
	processor := NewProcessor(store)
	assert.NoError(t, processor.Process())
	analyzer := NewAnalyzer(store, nil)
	pageVisitors, err := analyzer.PageVisitors(&Filter{
		From: day(2020, 12, 21, 0),
		To:   day(2020, 12, 21, 0),
	})
	assert.NoError(t, err)
	assert.Len(t, pageVisitors, 2)
	assert.Equal(t, "/second-page", pageVisitors[0].Path)
	assert.Equal(t, "/", pageVisitors[1].Path)
	assert.Equal(t, 2, pageVisitors[0].Stats[0].Visitors)
	assert.Equal(t, 1, pageVisitors[1].Stats[0].Visitors)
	assert.Equal(t, 1, pageVisitors[0].Stats[0].Bounces)
	assert.Equal(t, 0, pageVisitors[1].Stats[0].Bounces)
	visitors, err := analyzer.Visitors(&Filter{
		From: day(2020, 12, 21, 0),
		To:   day(2020, 12, 21, 0),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[0].Bounces)
	pageReferrer, err := analyzer.PageReferrer(&Filter{
		From: day(2020, 12, 21, 0),
		To:   day(2020, 12, 21, 0),
		Path: "/second-page",
	})
	assert.NoError(t, err)
	assert.Len(t, pageReferrer, 1)
	assert.Equal(t, 2, pageReferrer[0].Visitors)
	assert.Equal(t, 1, pageReferrer[0].Bounces)
	referrer, err := analyzer.Referrer(&Filter{
		From: day(2020, 12, 21, 0),
		To:   day(2020, 12, 21, 0),
	})
	assert.NoError(t, err)
	assert.Len(t, referrer, 1)
	assert.Equal(t, 2, referrer[0].Visitors)
	assert.Equal(t, 1, referrer[0].Bounces)
}

func TestProcessor_ProcessAverageSessionDuration(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)
	day := pastDay(3)
	createSessions(t, store, 0, day)
	processor := NewProcessor(store)
	assert.NoError(t, processor.Process())
	analyzer := NewAnalyzer(store, nil)
	visitors, err := analyzer.Visitors(&Filter{
		From: day,
		To:   day,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, 8, visitors[0].AverageSessionDurationSeconds)
	createSessions(t, store, 0, day)
	assert.NoError(t, processor.Process())
	visitors, err = analyzer.Visitors(&Filter{
		From: day,
		To:   day,
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, addAverage(8, (10+5)/2, 4), visitors[0].AverageSessionDurationSeconds)
}

func testProcess(t *testing.T, tenantID int64) {
	store := NewPostgresStore(postgresDB, nil)
	createTestdata(t, store, tenantID)
	processor := NewProcessor(store)

	if tenantID == 0 {
		assert.NoError(t, processor.Process())
	} else {
		assert.NoError(t, processor.ProcessTenant(NewTenantID(tenantID)))
	}

	checkHits(t, tenantID)
	checkVisitorStats(t, tenantID)
	checkVisitorTimeStats(t, tenantID)
	checkLanguageStats(t, tenantID)
	checkReferrerStats(t, tenantID)
	checkOSStats(t, tenantID)
	checkBrowserStats(t, tenantID)
	checkScreenStats(t, tenantID)
	checkScreenStatsClasses(t, tenantID)
	checkCountryStats(t, tenantID)
}

func checkHits(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	count := 1

	if tenantID != 0 {
		assert.NoError(t, db.Get(&count, `SELECT COUNT(1) FROM "hit" WHERE tenant_id = $1`, tenantID))
	} else {
		assert.NoError(t, db.Get(&count, `SELECT COUNT(1) FROM "hit"`))
	}

	assert.Equal(t, 0, count)
}

func checkVisitorStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []VisitorStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "visitor_stats" WHERE tenant_id = $1 AND "path" IS NOT NULL ORDER BY "day", "path"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "visitor_stats" WHERE "path" IS NOT NULL ORDER BY "day", "path"`))
	}

	assert.Len(t, stats, 4)
	assert.Equal(t, "/", stats[0].Path.String)
	assert.Equal(t, "/page", stats[1].Path.String)
	assert.Equal(t, "/", stats[2].Path.String)
	assert.Equal(t, "/different-page", stats[3].Path.String)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 2, stats[2].Visitors)
	assert.Equal(t, 1, stats[3].Visitors)
	assert.Equal(t, 2, stats[0].PlatformDesktop)
	assert.Equal(t, 1, stats[1].PlatformDesktop)
	assert.Equal(t, 1, stats[2].PlatformDesktop)
	assert.Equal(t, 0, stats[3].PlatformDesktop)
	assert.Equal(t, 0, stats[0].PlatformMobile)
	assert.Equal(t, 0, stats[1].PlatformMobile)
	assert.Equal(t, 0, stats[2].PlatformMobile)
	assert.Equal(t, 1, stats[3].PlatformMobile)
	assert.Equal(t, 0, stats[0].PlatformUnknown)
	assert.Equal(t, 0, stats[1].PlatformUnknown)
	assert.Equal(t, 1, stats[2].PlatformUnknown)
	assert.Equal(t, 0, stats[3].PlatformUnknown)
	assert.Equal(t, 2, stats[0].Bounces)
	assert.Equal(t, 1, stats[1].Bounces)
	assert.Equal(t, 2, stats[2].Bounces)
	assert.Equal(t, 1, stats[3].Bounces)
	assert.Equal(t, 2, stats[0].Views)
	assert.Equal(t, 1, stats[1].Views)
	assert.Equal(t, 2, stats[2].Views)
	assert.Equal(t, 1, stats[3].Views)
}

func checkVisitorTimeStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []VisitorTimeStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "visitor_time_stats" WHERE tenant_id = $1 ORDER BY "day", "hour"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "visitor_time_stats" ORDER BY "day", "hour"`))
	}

	assert.Len(t, stats, 48)
	assert.Equal(t, 2, stats[7].Visitors)
	assert.Equal(t, 1, stats[8].Visitors)
	assert.Equal(t, 2, stats[24+9].Visitors)
	assert.Equal(t, 1, stats[24+10].Visitors)
	assert.Equal(t, 7, stats[7].Hour)
	assert.Equal(t, 8, stats[8].Hour)
	assert.Equal(t, 9, stats[24+9].Hour)
	assert.Equal(t, 10, stats[24+10].Hour)
}

func checkLanguageStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []LanguageStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "language_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "language"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "language_stats" ORDER BY "day", "path", "language"`))
	}

	assert.Len(t, stats, 8)
	assert.Equal(t, "/", stats[0].Path.String)
	assert.Equal(t, "/page", stats[1].Path.String)
	assert.False(t, stats[2].Path.Valid)
	assert.False(t, stats[3].Path.Valid)
	assert.Equal(t, "/", stats[4].Path.String)
	assert.Equal(t, "/different-page", stats[5].Path.String)
	assert.False(t, stats[6].Path.Valid)
	assert.False(t, stats[7].Path.Valid)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
	assert.Equal(t, 2, stats[3].Visitors)
	assert.Equal(t, 2, stats[4].Visitors)
	assert.Equal(t, 1, stats[5].Visitors)
	assert.Equal(t, 2, stats[6].Visitors)
	assert.Equal(t, 1, stats[7].Visitors)
	assert.Equal(t, "en", stats[0].Language.String)
	assert.Equal(t, "de", stats[1].Language.String)
	assert.Equal(t, "de", stats[2].Language.String)
	assert.Equal(t, "en", stats[3].Language.String)
	assert.Equal(t, "en", stats[4].Language.String)
	assert.Equal(t, "jp", stats[5].Language.String)
	assert.Equal(t, "en", stats[6].Language.String)
	assert.Equal(t, "jp", stats[7].Language.String)
}

func checkReferrerStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []ReferrerStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "referrer_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "referrer"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "referrer_stats" ORDER BY "day", "path", "referrer"`))
	}

	assert.Len(t, stats, 7)
	assert.Equal(t, "/", stats[0].Path.String)
	assert.Equal(t, "/page", stats[1].Path.String)
	assert.False(t, stats[2].Path.Valid)
	assert.Equal(t, "/", stats[3].Path.String)
	assert.Equal(t, "/different-page", stats[4].Path.String)
	assert.False(t, stats[5].Path.Valid)
	assert.False(t, stats[6].Path.Valid)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 3, stats[2].Visitors)
	assert.Equal(t, 2, stats[3].Visitors)
	assert.Equal(t, 1, stats[4].Visitors)
	assert.Equal(t, 2, stats[5].Visitors)
	assert.Equal(t, 1, stats[6].Visitors)
	assert.Equal(t, 2, stats[0].Bounces)
	assert.Equal(t, 1, stats[1].Bounces)
	assert.Equal(t, 3, stats[2].Bounces)
	assert.Equal(t, 2, stats[3].Bounces)
	assert.Equal(t, 1, stats[4].Bounces)
	assert.Equal(t, 2, stats[5].Bounces)
	assert.Equal(t, 1, stats[6].Bounces)
	assert.Equal(t, "ref1", stats[0].Referrer.String)
	assert.Equal(t, "ref1", stats[1].Referrer.String)
	assert.Equal(t, "ref1", stats[2].Referrer.String)
	assert.Equal(t, "ref2", stats[3].Referrer.String)
	assert.Equal(t, "ref3", stats[4].Referrer.String)
	assert.Equal(t, "ref2", stats[5].Referrer.String)
	assert.Equal(t, "ref3", stats[6].Referrer.String)
}

func checkOSStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []OSStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "os_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "os", "os_version"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "os_stats" ORDER BY "day", "path", "os", "os_version"`))
	}

	assert.Len(t, stats, 10)
	assert.Equal(t, "/", stats[0].Path.String)
	assert.Equal(t, "/page", stats[1].Path.String)
	assert.False(t, stats[2].Path.Valid)
	assert.False(t, stats[3].Path.Valid)
	assert.Equal(t, "/", stats[4].Path.String)
	assert.Equal(t, "/", stats[5].Path.String)
	assert.Equal(t, "/different-page", stats[6].Path.String)
	assert.False(t, stats[7].Path.Valid)
	assert.False(t, stats[8].Path.Valid)
	assert.False(t, stats[9].Path.Valid)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
	assert.Equal(t, 2, stats[3].Visitors)
	assert.Equal(t, 1, stats[4].Visitors)
	assert.Equal(t, 1, stats[5].Visitors)
	assert.Equal(t, 1, stats[6].Visitors)
	assert.Equal(t, 1, stats[7].Visitors)
	assert.Equal(t, 1, stats[8].Visitors)
	assert.Equal(t, 1, stats[9].Visitors)
	assert.Equal(t, OSWindows, stats[0].OS.String)
	assert.Equal(t, OSMac, stats[1].OS.String)
	assert.Equal(t, OSMac, stats[2].OS.String)
	assert.Equal(t, OSWindows, stats[3].OS.String)
	assert.Equal(t, OSLinux, stats[4].OS.String)
	assert.Equal(t, OSWindows, stats[5].OS.String)
	assert.Equal(t, OSAndroid, stats[6].OS.String)
	assert.Equal(t, OSAndroid, stats[7].OS.String)
	assert.Equal(t, OSLinux, stats[8].OS.String)
	assert.Equal(t, OSWindows, stats[9].OS.String)
	assert.Equal(t, "10", stats[0].OSVersion.String)
	assert.Equal(t, "10.15.3", stats[1].OSVersion.String)
	assert.False(t, stats[2].OSVersion.Valid)
	assert.False(t, stats[3].OSVersion.Valid)
	assert.False(t, stats[4].OSVersion.Valid)
	assert.Equal(t, "10", stats[5].OSVersion.String)
	assert.Equal(t, "8.0", stats[6].OSVersion.String)
	assert.False(t, stats[7].OSVersion.Valid)
	assert.False(t, stats[8].OSVersion.Valid)
	assert.False(t, stats[9].OSVersion.Valid)
}

func checkBrowserStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []BrowserStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "browser_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "browser", "browser_version"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "browser_stats" ORDER BY "day", "path", "browser", "browser_version"`))
	}

	assert.Len(t, stats, 8)
	assert.Equal(t, "/", stats[0].Path.String)
	assert.Equal(t, "/page", stats[1].Path.String)
	assert.False(t, stats[2].Path.Valid)
	assert.Equal(t, "/", stats[3].Path.String)
	assert.Equal(t, "/", stats[4].Path.String)
	assert.Equal(t, "/different-page", stats[5].Path.String)
	assert.False(t, stats[6].Path.Valid)
	assert.False(t, stats[7].Path.Valid)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 3, stats[2].Visitors)
	assert.Equal(t, 1, stats[3].Visitors)
	assert.Equal(t, 1, stats[4].Visitors)
	assert.Equal(t, 1, stats[5].Visitors)
	assert.Equal(t, 1, stats[6].Visitors)
	assert.Equal(t, 2, stats[7].Visitors)
	assert.Equal(t, BrowserChrome, stats[0].Browser.String)
	assert.Equal(t, BrowserChrome, stats[1].Browser.String)
	assert.Equal(t, BrowserChrome, stats[2].Browser.String)
	assert.Equal(t, BrowserFirefox, stats[3].Browser.String)
	assert.Equal(t, BrowserFirefox, stats[4].Browser.String)
	assert.Equal(t, BrowserChrome, stats[5].Browser.String)
	assert.Equal(t, BrowserChrome, stats[6].Browser.String)
	assert.Equal(t, BrowserFirefox, stats[7].Browser.String)
	assert.Equal(t, "84.0", stats[0].BrowserVersion.String)
	assert.Equal(t, "84.0", stats[1].BrowserVersion.String)
	assert.False(t, stats[2].BrowserVersion.Valid)
	assert.Equal(t, "53.0", stats[3].BrowserVersion.String)
	assert.Equal(t, "54.0", stats[4].BrowserVersion.String)
	assert.Equal(t, "84.0", stats[5].BrowserVersion.String)
	assert.False(t, stats[6].BrowserVersion.Valid)
	assert.False(t, stats[7].BrowserVersion.Valid)
}

func checkScreenStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []ScreenStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "screen_stats" WHERE tenant_id = $1 AND "class" IS NULL ORDER BY "day", "width", "height"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "screen_stats" WHERE "class" IS NULL ORDER BY "day", "width", "height"`))
	}

	assert.Len(t, stats, 3)
	assert.Equal(t, day(2020, 6, 21, 0), stats[0].Day.UTC())
	assert.Equal(t, day(2020, 6, 22, 0), stats[1].Day.UTC())
	assert.Equal(t, day(2020, 6, 22, 0), stats[2].Day.UTC())
	assert.Equal(t, 640, stats[0].Width)
	assert.Equal(t, 640, stats[1].Width)
	assert.Equal(t, 1920, stats[2].Width)
	assert.Equal(t, 1080, stats[0].Height)
	assert.Equal(t, 1024, stats[1].Height)
	assert.Equal(t, 1080, stats[2].Height)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
}

func checkScreenStatsClasses(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []ScreenStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "screen_stats" WHERE tenant_id = $1 AND "class" IS NOT NULL ORDER BY "day", "class"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "screen_stats" WHERE "class" IS NOT NULL ORDER BY "day", "class"`))
	}

	assert.Len(t, stats, 3)
	assert.Equal(t, day(2020, 6, 21, 0), stats[0].Day.UTC())
	assert.Equal(t, day(2020, 6, 22, 0), stats[1].Day.UTC())
	assert.Equal(t, day(2020, 6, 22, 0), stats[2].Day.UTC())
	assert.Equal(t, "M", stats[0].Class.String)
	assert.Equal(t, "M", stats[1].Class.String)
	assert.Equal(t, "XXL", stats[2].Class.String)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
}

func checkCountryStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []CountryStats

	if tenantID != 0 {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "country_stats" WHERE tenant_id = $1 ORDER BY "day", "country_code"`, tenantID))
	} else {
		assert.NoError(t, db.Select(&stats, `SELECT * FROM "country_stats" ORDER BY "day", "country_code"`))
	}

	assert.Len(t, stats, 5)
	assert.Equal(t, day(2020, 6, 21, 0), stats[0].Day.UTC())
	assert.Equal(t, day(2020, 6, 21, 0), stats[1].Day.UTC())
	assert.Equal(t, day(2020, 6, 22, 0), stats[2].Day.UTC())
	assert.Equal(t, day(2020, 6, 22, 0), stats[3].Day.UTC())
	assert.Equal(t, day(2020, 6, 22, 0), stats[4].Day.UTC())
	assert.Equal(t, "de", stats[0].CountryCode.String)
	assert.Equal(t, "gb", stats[1].CountryCode.String)
	assert.Equal(t, "de", stats[2].CountryCode.String)
	assert.Equal(t, "gb", stats[3].CountryCode.String)
	assert.Equal(t, "jp", stats[4].CountryCode.String)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 2, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
	assert.Equal(t, 1, stats[3].Visitors)
	assert.Equal(t, 1, stats[4].Visitors)
}

func createTestdata(t *testing.T, store Store, tenantID int64) {
	cleanupDB(t)
	createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", day(2020, 6, 21, 7), time.Time{}, OSWindows, "10", BrowserChrome, "84.0", "de", true, false, 0, 0)
	createHit(t, store, tenantID, "fp2", "/", "en", "ua2", "ref1", day(2020, 6, 21, 7), time.Time{}, OSWindows, "10", BrowserChrome, "84.0", "gb", true, false, 640, 1080)
	createHit(t, store, tenantID, "fp3", "/page", "de", "ua3", "ref1", day(2020, 6, 21, 8), time.Time{}, OSMac, "10.15.3", BrowserChrome, "84.0", "gb", true, false, 640, 1080)
	createHit(t, store, tenantID, "fp4", "/", "en", "ua4", "ref2", day(2020, 6, 22, 9), time.Time{}, OSWindows, "10", BrowserFirefox, "53.0", "gb", true, false, 640, 1024)
	createHit(t, store, tenantID, "fp5", "/", "en", "ua5", "ref2", day(2020, 6, 22, 9), time.Time{}, OSLinux, "", BrowserFirefox, "54.0", "de", false, false, 1920, 1080)
	createHit(t, store, tenantID, "fp6", "/different-page", "jp", "ua6", "ref3", day(2020, 6, 22, 10), time.Time{}, OSAndroid, "8.0", BrowserChrome, "84.0", "jp", false, true, 0, 0)
}

func createSessions(t *testing.T, store Store, tenantID int64, day time.Time) {
	createHit(t, store, tenantID, "fp", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*100), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, tenantID, "fp", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*95), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, tenantID, "fp", "/p3", "en", "ua", "", day.Add(time.Hour-time.Second*90), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, tenantID, "fp", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*10), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, tenantID, "fp", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*5), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
}

func createHit(t *testing.T, store Store, tenantID int64, fingerprint, path, lang, userAgent, ref string, time, session time.Time, os, osVersion, browser, browserVersion, countryCode string, desktop, mobile bool, w, h int) {
	screenClass := GetScreenClass(w)
	hit := Hit{
		BaseEntity:     BaseEntity{TenantID: NewTenantID(tenantID)},
		Fingerprint:    fingerprint,
		Session:        sql.NullTime{Time: session, Valid: !session.IsZero()},
		Path:           path,
		Language:       sql.NullString{String: lang, Valid: lang != ""},
		UserAgent:      sql.NullString{String: userAgent, Valid: userAgent != ""},
		Referrer:       sql.NullString{String: ref, Valid: ref != ""},
		OS:             sql.NullString{String: os, Valid: os != ""},
		OSVersion:      sql.NullString{String: osVersion, Valid: osVersion != ""},
		Browser:        sql.NullString{String: browser, Valid: browser != ""},
		BrowserVersion: sql.NullString{String: browserVersion, Valid: browserVersion != ""},
		CountryCode:    sql.NullString{String: countryCode, Valid: countryCode != ""},
		Desktop:        desktop,
		Mobile:         mobile,
		ScreenWidth:    w,
		ScreenHeight:   h,
		ScreenClass:    sql.NullString{String: screenClass, Valid: screenClass != ""},
		Time:           time,
	}

	assert.NoError(t, store.SaveHits([]Hit{hit}))
}

func day(year, month, day, hour int) time.Time {
	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC)
}

func sumUpVisitors(stats []Stats) int {
	sum := 0

	for _, s := range stats {
		sum += s.Visitors
	}

	return sum
}
