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

func testProcess(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		createTestdata(t, store, tenantID)
		processor := NewProcessor(store, nil)

		if tenantID == 0 {
			if err := processor.Process(); err != nil {
				t.Fatalf("Data must have been processed, but was: %v", err)
			}
		} else {
			if err := processor.ProcessTenant(sql.NullInt64{Int64: tenantID, Valid: true}); err != nil {
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
	}
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
		if err := db.Select(&stats, `SELECT * FROM "visitor_stats" WHERE tenant_id = $1 ORDER BY "day", "path"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "visitor_stats" ORDER BY "day", "path"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 4 {
		t.Fatalf("Four stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path != "/" || stats[0].Visitors != 2 || stats[0].PlatformDesktop != 2 || stats[0].PlatformMobile != 0 || stats[0].PlatformUnknown != 0 ||
		stats[1].Path != "/page" || stats[1].Visitors != 1 || stats[1].PlatformDesktop != 1 || stats[1].PlatformMobile != 0 || stats[1].PlatformUnknown != 0 ||
		stats[2].Path != "/" || stats[2].Visitors != 2 || stats[2].PlatformDesktop != 1 || stats[2].PlatformMobile != 0 || stats[2].PlatformUnknown != 1 ||
		stats[3].Path != "/different-page" || stats[3].Visitors != 1 || stats[3].PlatformDesktop != 0 || stats[3].PlatformMobile != 1 || stats[3].PlatformUnknown != 0 {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func checkVisitorTimeStats(t *testing.T, tenantID int64) {
	db := sqlx.NewDb(postgresDB, "postgres")
	var stats []VisitorTimeStats

	if tenantID != 0 {
		if err := db.Select(&stats, `SELECT * FROM "visitor_time_stats" WHERE tenant_id = $1 ORDER BY "day", "path", "hour"`, tenantID); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Select(&stats, `SELECT * FROM "visitor_time_stats" ORDER BY "day", "path", "hour"`); err != nil {
			t.Fatal(err)
		}
	}

	if len(stats) != 96 {
		t.Fatalf("96 stats must have been created, but was: %v", len(stats))
	}

	if stats[7].Path != "/" || stats[7].Visitors != 2 || stats[7].Hour != 7 ||
		stats[32].Path != "/page" || stats[32].Visitors != 1 || stats[32].Hour != 8 ||
		stats[57].Path != "/" || stats[57].Visitors != 2 || stats[57].Hour != 9 ||
		stats[82].Path != "/different-page" || stats[82].Visitors != 1 || stats[82].Hour != 10 {
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

	if len(stats) != 4 {
		t.Fatalf("Four stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path != "/" || stats[0].Visitors != 2 || stats[0].Language.String != "en" ||
		stats[1].Path != "/page" || stats[1].Visitors != 1 || stats[1].Language.String != "de" ||
		stats[2].Path != "/" || stats[2].Visitors != 2 || stats[2].Language.String != "en" ||
		stats[3].Path != "/different-page" || stats[3].Visitors != 1 || stats[3].Language.String != "jp" {
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

	if len(stats) != 4 {
		t.Fatalf("Four stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path != "/" || stats[0].Visitors != 2 || stats[0].Referrer.String != "ref1" ||
		stats[1].Path != "/page" || stats[1].Visitors != 1 || stats[1].Referrer.String != "ref1" ||
		stats[2].Path != "/" || stats[2].Visitors != 2 || stats[2].Referrer.String != "ref2" ||
		stats[3].Path != "/different-page" || stats[3].Visitors != 1 || stats[3].Referrer.String != "ref3" {
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

	if len(stats) != 5 {
		t.Fatalf("Five stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path != "/" || stats[0].Visitors != 2 || stats[0].OS.String != OSWindows || stats[0].OSVersion.String != "10" ||
		stats[1].Path != "/page" || stats[1].Visitors != 1 || stats[1].OS.String != OSMac || stats[1].OSVersion.String != "10.15.3" ||
		stats[2].Path != "/" || stats[2].Visitors != 1 || stats[2].OS.String != OSLinux || stats[2].OSVersion.String != "" ||
		stats[3].Path != "/" || stats[3].Visitors != 1 || stats[3].OS.String != OSWindows || stats[3].OSVersion.String != "10" ||
		stats[4].Path != "/different-page" || stats[4].Visitors != 1 || stats[4].OS.String != OSAndroid || stats[4].OSVersion.String != "8.0" {
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

	if len(stats) != 5 {
		t.Fatalf("Five stats must have been created, but was: %v", len(stats))
	}

	if stats[0].Path != "/" || stats[0].Visitors != 2 || stats[0].Browser.String != BrowserChrome || stats[0].BrowserVersion.String != "84.0" ||
		stats[1].Path != "/page" || stats[1].Visitors != 1 || stats[1].Browser.String != BrowserChrome || stats[1].BrowserVersion.String != "84.0" ||
		stats[2].Path != "/" || stats[2].Visitors != 1 || stats[2].Browser.String != BrowserFirefox || stats[2].BrowserVersion.String != "53.0" ||
		stats[3].Path != "/" || stats[3].Visitors != 1 || stats[3].Browser.String != BrowserFirefox || stats[3].BrowserVersion.String != "54.0" ||
		stats[4].Path != "/different-page" || stats[4].Visitors != 1 || stats[4].Browser.String != BrowserChrome || stats[4].BrowserVersion.String != "84.0" {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func createTestdata(t *testing.T, store Store, tenantID int64) {
	cleanupDB(t)
	createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", day(2020, 6, 21, 7), OSWindows, "10", BrowserChrome, "84.0", true, false)
	createHit(t, store, tenantID, "fp2", "/", "en", "ua2", "ref1", day(2020, 6, 21, 7), OSWindows, "10", BrowserChrome, "84.0", true, false)
	createHit(t, store, tenantID, "fp3", "/page", "de", "ua3", "ref1", day(2020, 6, 21, 8), OSMac, "10.15.3", BrowserChrome, "84.0", true, false)
	createHit(t, store, tenantID, "fp4", "/", "en", "ua4", "ref2", day(2020, 6, 22, 9), OSWindows, "10", BrowserFirefox, "53.0", true, false)
	createHit(t, store, tenantID, "fp5", "/", "en", "ua5", "ref2", day(2020, 6, 22, 9), OSLinux, "", BrowserFirefox, "54.0", false, false)
	createHit(t, store, tenantID, "fp6", "/different-page", "jp", "ua6", "ref3", day(2020, 6, 22, 10), OSAndroid, "8.0", BrowserChrome, "84.0", false, true)
}

func createHit(t *testing.T, store Store, tenantID int64, fingerprint, path, lang, userAgent, ref string, time time.Time, os, osVersion, browser, browserVersion string, desktop, mobile bool) {
	hit := Hit{
		BaseEntity:     BaseEntity{TenantID: NewTenantID(tenantID)},
		Fingerprint:    fingerprint,
		Path:           sql.NullString{String: path, Valid: path != ""},
		Language:       sql.NullString{String: lang, Valid: path != ""},
		UserAgent:      sql.NullString{String: userAgent, Valid: path != ""},
		Referrer:       sql.NullString{String: ref, Valid: path != ""},
		OS:             sql.NullString{String: os, Valid: os != ""},
		OSVersion:      sql.NullString{String: osVersion, Valid: osVersion != ""},
		Browser:        sql.NullString{String: browser, Valid: browser != ""},
		BrowserVersion: sql.NullString{String: browserVersion, Valid: browserVersion != ""},
		Desktop:        desktop,
		Mobile:         mobile,
		Time:           time,
	}

	if err := store.SaveHits([]Hit{hit}); err != nil {
		t.Fatal(err)
	}
}

func day(year, month, day, hour int) time.Time {
	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC)
}
