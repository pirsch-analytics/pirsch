package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
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

	if err := processor.Process(); err != nil {
		t.Fatalf("Data must have been processed, but was: %v", err)
	}

	checkHits(t, 0)
	db := sqlx.NewDb(postgresDB, "postgres")
	var visitorStats []VisitorStats
	var timeStats []VisitorTimeStats

	if err := db.Select(&visitorStats, `SELECT * FROM "visitor_stats" ORDER BY "day", "path"`); err != nil {
		t.Fatal(err)
	}

	if err := db.Select(&timeStats, `SELECT * FROM "visitor_time_stats" ORDER BY "day"`); err != nil {
		t.Fatal(err)
	}

	if len(visitorStats) != 2 {
		t.Fatalf("Two visitor stats must have been created, but was: %v", len(visitorStats))
	}

	if len(timeStats) != 24 {
		t.Fatalf("24 visitor time stats must have been created, but was: %v", len(visitorStats))
	}

	if visitorStats[0].Visitors != 2 ||
		visitorStats[0].Sessions != 3 {
		t.Fatalf("Visitor stats must have two visitors and three sessions, but was: %v %v", visitorStats[0].Visitors, visitorStats[0].Sessions)
	}

	if timeStats[4].Visitors != 1 || timeStats[5].Visitors != 1 {
		t.Fatalf("Visitor time stats must have two visitors, but was: %v", timeStats)
	}
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

	if err := processor.Process(); err != nil {
		t.Fatalf("Data must have been processed, but was: %v", err)
	}

	analyzer := NewAnalyzer(store, nil)
	pageVisitors, err := analyzer.PageVisitors(&Filter{
		From: day(2020, 12, 25, 0),
		To:   day(2020, 12, 28, 0),
	})

	if err != nil {
		t.Fatalf("Page visitors must have been returned, but was: %v", err)
	}

	if len(pageVisitors) != 2 ||
		pageVisitors[0].Path != "/" ||
		sumUpVisitors(pageVisitors[0].Stats) != 1 ||
		pageVisitors[1].Path != "/path" ||
		sumUpVisitors(pageVisitors[1].Stats) != 2 {
		t.Fatalf("Page visitor statistics not as expected: %v", pageVisitors)
	}

	visitors, err := analyzer.Visitors(nil)

	if err != nil {
		t.Fatalf("Visitors must have been returned, but was: %v", err)
	}

	if sumUpVisitors(visitors) != 2 {
		t.Fatalf("Visitor statistics not as expected: %v", visitors)
	}
}

func testProcess(t *testing.T, tenantID int64) {
	store := NewPostgresStore(postgresDB, nil)
	createTestdata(t, store, tenantID)
	processor := NewProcessor(store)

	if tenantID == 0 {
		if err := processor.Process(); err != nil {
			t.Fatalf("Data must have been processed, but was: %v", err)
		}
	} else {
		if err := processor.ProcessTenant(NewTenantID(tenantID)); err != nil {
			t.Fatalf("Data must have been processed, but was: %v", err)
		}
	}

	checkHits(t, tenantID)
	checkVisitorStats(t, tenantID)
	checkVisitorTimeStats(t, tenantID)
	checkLanguageStats(t, tenantID)
	checkReferrerStats(t, tenantID)
	checkOSStats(t, tenantID)
	checkBrowserStats(t, tenantID)
	checkScreenStats(t, tenantID)
	checkCountryStats(t, tenantID)
}

func checkHits(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	count := 1

	if tenantID != 0 {
		if err := db.Get(&count, `SELECT COUNT(1) FROM "hit" WHERE tenant_id = $1`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Get(&count, `SELECT COUNT(1) FROM "hit"`); err != nil {
			t.Fatal(err)
		}
	}

	if count != 0 {
		t.Fatalf("Hits must have been cleaned up, but was: %v", count)
	}
}

func checkVisitorStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []VisitorStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "visitor_stats" WHERE tenant_id = $1 AND "path" IS NOT NULL ORDER BY "day", "path"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "visitor_stats" WHERE "path" IS NOT NULL ORDER BY "day", "path"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 4 {
		t.Fatalf("Four stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path.String != "/" || stats[0].Visitors != 2 || stats[0].PlatformDesktop != 2 || stats[0].PlatformMobile != 0 || stats[0].PlatformUnknown != 0 || stats[0].Bounces != 2 ||
		stats[1].Path.String != "/page" || stats[1].Visitors != 1 || stats[1].PlatformDesktop != 1 || stats[1].PlatformMobile != 0 || stats[1].PlatformUnknown != 0 || stats[1].Bounces != 1 ||
		stats[2].Path.String != "/" || stats[2].Visitors != 2 || stats[2].PlatformDesktop != 1 || stats[2].PlatformMobile != 0 || stats[2].PlatformUnknown != 1 || stats[2].Bounces != 2 ||
		stats[3].Path.String != "/different-page" || stats[3].Visitors != 1 || stats[3].PlatformDesktop != 0 || stats[3].PlatformMobile != 1 || stats[3].PlatformUnknown != 0 || stats[3].Bounces != 1 {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkVisitorTimeStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []VisitorTimeStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "visitor_time_stats" WHERE tenant_id = $1 ORDER BY "day", "hour"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "visitor_time_stats" ORDER BY "day", "hour"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 48 {
		t.Fatalf("48 stats must have been created, but was: %v", len(stats))
	}

	if stats[7].Visitors != 2 || stats[7].Hour != 7 ||
		stats[8].Visitors != 1 || stats[8].Hour != 8 ||
		stats[24+9].Visitors != 2 || stats[24+9].Hour != 9 ||
		stats[24+10].Visitors != 1 || stats[24+10].Hour != 10 {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkLanguageStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []LanguageStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "language_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "language"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "language_stats" ORDER BY "day", "path", "language"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 8 {
		t.Fatalf("Four stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path.String != "/" || stats[0].Visitors != 2 || stats[0].Language.String != "en" ||
		stats[1].Path.String != "/page" || stats[1].Visitors != 1 || stats[1].Language.String != "de" ||
		stats[2].Path.Valid || stats[2].Visitors != 1 || stats[2].Language.String != "de" ||
		stats[3].Path.Valid || stats[3].Visitors != 2 || stats[3].Language.String != "en" ||
		stats[4].Path.String != "/" || stats[4].Visitors != 2 || stats[4].Language.String != "en" ||
		stats[5].Path.String != "/different-page" || stats[5].Visitors != 1 || stats[5].Language.String != "jp" ||
		stats[6].Path.Valid || stats[6].Visitors != 2 || stats[6].Language.String != "en" ||
		stats[7].Path.Valid || stats[7].Visitors != 1 || stats[7].Language.String != "jp" {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkReferrerStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []ReferrerStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "referrer_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "referrer"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "referrer_stats" ORDER BY "day", "path", "referrer"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 7 {
		t.Fatalf("Seven stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path.String != "/" || stats[0].Visitors != 2 || stats[0].Referrer.String != "ref1" ||
		stats[1].Path.String != "/page" || stats[1].Visitors != 1 || stats[1].Referrer.String != "ref1" ||
		stats[2].Path.Valid || stats[2].Visitors != 3 || stats[2].Referrer.String != "ref1" ||
		stats[3].Path.String != "/" || stats[3].Visitors != 2 || stats[3].Referrer.String != "ref2" ||
		stats[4].Path.String != "/different-page" || stats[4].Visitors != 1 || stats[4].Referrer.String != "ref3" ||
		stats[5].Path.Valid || stats[5].Visitors != 2 || stats[5].Referrer.String != "ref2" ||
		stats[6].Path.Valid || stats[6].Visitors != 1 || stats[6].Referrer.String != "ref3" {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkOSStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []OSStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "os_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "os", "os_version"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "os_stats" ORDER BY "day", "path", "os", "os_version"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 10 {
		t.Fatalf("Ten stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path.String != "/" || stats[0].Visitors != 2 || stats[0].OS.String != OSWindows || stats[0].OSVersion.String != "10" ||
		stats[1].Path.String != "/page" || stats[1].Visitors != 1 || stats[1].OS.String != OSMac || stats[1].OSVersion.String != "10.15.3" ||
		stats[2].Path.Valid || stats[2].Visitors != 1 || stats[2].OS.String != OSMac || stats[2].OSVersion.Valid ||
		stats[3].Path.Valid || stats[3].Visitors != 2 || stats[3].OS.String != OSWindows || stats[3].OSVersion.Valid ||
		stats[4].Path.String != "/" || stats[4].Visitors != 1 || stats[4].OS.String != OSLinux || stats[4].OSVersion.String != "" ||
		stats[5].Path.String != "/" || stats[5].Visitors != 1 || stats[5].OS.String != OSWindows || stats[5].OSVersion.String != "10" ||
		stats[6].Path.String != "/different-page" || stats[6].Visitors != 1 || stats[6].OS.String != OSAndroid || stats[6].OSVersion.String != "8.0" ||
		stats[7].Path.Valid || stats[7].Visitors != 1 || stats[7].OS.String != OSAndroid || stats[7].OSVersion.Valid ||
		stats[8].Path.Valid || stats[8].Visitors != 1 || stats[8].OS.String != OSLinux || stats[8].OSVersion.Valid ||
		stats[9].Path.Valid || stats[9].Visitors != 1 || stats[9].OS.String != OSWindows || stats[9].OSVersion.Valid {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkBrowserStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []BrowserStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "browser_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "browser", "browser_version"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "browser_stats" ORDER BY "day", "path", "browser", "browser_version"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 8 {
		t.Fatalf("Eight stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path.String != "/" || stats[0].Visitors != 2 || stats[0].Browser.String != BrowserChrome || stats[0].BrowserVersion.String != "84.0" ||
		stats[1].Path.String != "/page" || stats[1].Visitors != 1 || stats[1].Browser.String != BrowserChrome || stats[1].BrowserVersion.String != "84.0" ||
		stats[2].Path.Valid || stats[2].Visitors != 3 || stats[2].Browser.String != BrowserChrome || stats[2].BrowserVersion.Valid ||
		stats[3].Path.String != "/" || stats[3].Visitors != 1 || stats[3].Browser.String != BrowserFirefox || stats[3].BrowserVersion.String != "53.0" ||
		stats[4].Path.String != "/" || stats[4].Visitors != 1 || stats[4].Browser.String != BrowserFirefox || stats[4].BrowserVersion.String != "54.0" ||
		stats[5].Path.String != "/different-page" || stats[5].Visitors != 1 || stats[5].Browser.String != BrowserChrome || stats[5].BrowserVersion.String != "84.0" ||
		stats[6].Path.Valid || stats[6].Visitors != 1 || stats[6].Browser.String != BrowserChrome || stats[6].BrowserVersion.Valid ||
		stats[7].Path.Valid || stats[7].Visitors != 2 || stats[7].Browser.String != BrowserFirefox || stats[7].BrowserVersion.Valid {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkScreenStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []ScreenStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "screen_stats" WHERE tenant_id = $1 ORDER BY "day", "width", "height"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "screen_stats" ORDER BY "day", "width", "height"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 5 {
		t.Fatalf("Five stats must have been created, but was: %v", len(stats))
	}

	if !stats[0].Day.Equal(day(2020, 6, 21, 0)) || stats[0].Width != 0 || stats[0].Height != 0 || stats[0].Visitors != 1 ||
		!stats[1].Day.Equal(day(2020, 6, 21, 0)) || stats[1].Width != 640 || stats[1].Height != 1080 || stats[1].Visitors != 2 ||
		!stats[2].Day.Equal(day(2020, 6, 22, 0)) || stats[2].Width != 0 || stats[2].Height != 0 || stats[2].Visitors != 1 ||
		!stats[3].Day.Equal(day(2020, 6, 22, 0)) || stats[3].Width != 640 || stats[3].Height != 1024 || stats[3].Visitors != 1 ||
		!stats[4].Day.Equal(day(2020, 6, 22, 0)) || stats[4].Width != 1920 || stats[4].Height != 1080 || stats[4].Visitors != 1 {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkCountryStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []CountryStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "country_stats" WHERE tenant_id = $1 ORDER BY "day", "country_code"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "country_stats" ORDER BY "day", "country_code"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 5 {
		t.Fatalf("Five stats must have been created, but was: %v", len(stats))
	}

	if !stats[0].Day.Equal(day(2020, 6, 21, 0)) || stats[0].CountryCode.String != "de" || stats[0].Visitors != 1 ||
		!stats[1].Day.Equal(day(2020, 6, 21, 0)) || stats[1].CountryCode.String != "gb" || stats[1].Visitors != 2 ||
		!stats[2].Day.Equal(day(2020, 6, 22, 0)) || stats[2].CountryCode.String != "de" || stats[2].Visitors != 1 ||
		!stats[3].Day.Equal(day(2020, 6, 22, 0)) || stats[3].CountryCode.String != "gb" || stats[3].Visitors != 1 ||
		!stats[4].Day.Equal(day(2020, 6, 22, 0)) || stats[4].CountryCode.String != "jp" || stats[4].Visitors != 1 {
		t.Fatalf("Stats not as expected: %v", stats)
	}
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

func createHit(t *testing.T, store Store, tenantID int64, fingerprint, path, lang, userAgent, ref string, time, session time.Time, os, osVersion, browser, browserVersion, countryCode string, desktop, mobile bool, w, h int) {
	hit := Hit{
		BaseEntity:     BaseEntity{TenantID: NewTenantID(tenantID)},
		Fingerprint:    fingerprint,
		Session:        sql.NullTime{Time: session, Valid: !session.IsZero()},
		Path:           path,
		Language:       sql.NullString{String: lang, Valid: path != ""},
		UserAgent:      sql.NullString{String: userAgent, Valid: path != ""},
		Referrer:       sql.NullString{String: ref, Valid: path != ""},
		OS:             sql.NullString{String: os, Valid: os != ""},
		OSVersion:      sql.NullString{String: osVersion, Valid: osVersion != ""},
		Browser:        sql.NullString{String: browser, Valid: browser != ""},
		BrowserVersion: sql.NullString{String: browserVersion, Valid: browserVersion != ""},
		CountryCode:    sql.NullString{String: countryCode, Valid: countryCode != ""},
		Desktop:        desktop,
		Mobile:         mobile,
		ScreenWidth:    w,
		ScreenHeight:   h,
		Time:           time,
	}

	if err := store.SaveHits([]Hit{hit}); err != nil {
		t.Fatal(err)
	}
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
