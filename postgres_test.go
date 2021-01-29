package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"
)

func TestPostgresStore_SaveVisitorStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveVisitorStats(nil, &VisitorStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
			Sessions: 59,
			Bounces:  11,
		},
		PlatformDesktop: 123,
		PlatformMobile:  89,
		PlatformUnknown: 52,
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(VisitorStats)

	if err := db.Get(stats, `SELECT * FROM "visitor_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	stats.Sessions = 17
	stats.Bounces = 1
	stats.PlatformDesktop = 5
	stats.PlatformMobile = 3
	stats.PlatformUnknown = 1
	err = store.SaveVisitorStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "visitor_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 ||
		stats.Sessions != 59+17 ||
		stats.Bounces != 11+1 ||
		stats.PlatformDesktop != 123+5 ||
		stats.PlatformMobile != 89+3 ||
		stats.PlatformUnknown != 52+1 {
		t.Fatalf("Entity not as expected: %v", stats)
	}

	stats.Path = sql.NullString{String: "/path", Valid: true}
	err = store.SaveVisitorStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	var entries []VisitorStats

	if err := db.Select(&entries, `SELECT * FROM "visitor_stats"`); err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatal("New entry must have been created for path")
	}
}

func TestPostgresStore_SaveVisitorTimeStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveVisitorTimeStats(nil, &VisitorTimeStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		Hour: 5,
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(VisitorTimeStats)

	if err := db.Get(stats, `SELECT * FROM "visitor_time_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	err = store.SaveVisitorTimeStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "visitor_time_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 {
		t.Fatalf("Entity not as expected: %v", stats)
	}
}

func TestPostgresStore_SaveLanguageStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveLanguageStats(nil, &LanguageStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		Language: sql.NullString{String: "en", Valid: true},
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(LanguageStats)

	if err := db.Get(stats, `SELECT * FROM "language_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	err = store.SaveLanguageStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "language_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 ||
		stats.Language.String != "en" {
		t.Fatalf("Entity not as expected: %v", stats)
	}

	stats.Path = sql.NullString{String: "/path", Valid: true}
	err = store.SaveLanguageStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	var entries []LanguageStats

	if err := db.Select(&entries, `SELECT * FROM "language_stats"`); err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatal("New entry must have been created for path")
	}
}

func TestPostgresStore_SaveReferrerStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveReferrerStats(nil, &ReferrerStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		Referrer:     sql.NullString{String: "ref", Valid: true},
		ReferrerName: sql.NullString{String: "Ref", Valid: true},
		ReferrerIcon: sql.NullString{String: "icon", Valid: true},
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(ReferrerStats)

	if err := db.Get(stats, `SELECT * FROM "referrer_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	err = store.SaveReferrerStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "referrer_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 ||
		stats.Referrer.String != "ref" ||
		stats.ReferrerName.String != "Ref" ||
		stats.ReferrerIcon.String != "icon" {
		t.Fatalf("Entity not as expected: %v", stats)
	}

	stats.Path = sql.NullString{String: "/path", Valid: true}
	err = store.SaveReferrerStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	var entries []ReferrerStats

	if err := db.Select(&entries, `SELECT * FROM "referrer_stats"`); err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatal("New entry must have been created for path")
	}
}

func TestPostgresStore_SaveOSStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveOSStats(nil, &OSStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		OS:        sql.NullString{String: OSWindows, Valid: true},
		OSVersion: sql.NullString{String: "10", Valid: true},
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(OSStats)

	if err := db.Get(stats, `SELECT * FROM "os_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	err = store.SaveOSStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "os_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 ||
		stats.OS.String != OSWindows ||
		stats.OSVersion.String != "10" {
		t.Fatalf("Entity not as expected: %v", stats)
	}

	stats.Path = sql.NullString{String: "/path", Valid: true}
	err = store.SaveOSStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	var entries []OSStats

	if err := db.Select(&entries, `SELECT * FROM "os_stats"`); err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatal("New entry must have been created for path")
	}
}

func TestPostgresStore_SaveBrowserStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveBrowserStats(nil, &BrowserStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		Browser:        sql.NullString{String: BrowserChrome, Valid: true},
		BrowserVersion: sql.NullString{String: "84.0", Valid: true},
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(BrowserStats)

	if err := db.Get(stats, `SELECT * FROM "browser_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	err = store.SaveBrowserStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "browser_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 ||
		stats.Browser.String != BrowserChrome ||
		stats.BrowserVersion.String != "84.0" {
		t.Fatalf("Entity not as expected: %v", stats)
	}

	stats.Path = sql.NullString{String: "/path", Valid: true}
	err = store.SaveBrowserStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	var entries []BrowserStats

	if err := db.Select(&entries, `SELECT * FROM "browser_stats"`); err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatal("New entry must have been created for path")
	}
}

func TestPostgresStore_SaveScreenStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveScreenStats(nil, &ScreenStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		Width:  1920,
		Height: 1080,
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(ScreenStats)

	if err := db.Get(stats, `SELECT * FROM "screen_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	err = store.SaveScreenStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "screen_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 ||
		stats.Width != 1920 ||
		stats.Height != 1080 ||
		stats.Class.Valid {
		t.Fatalf("Entity not as expected: %v", stats)
	}
}

func TestPostgresStore_SaveScreenStatsClass(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveScreenStats(nil, &ScreenStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		Width:  1920,
		Height: 1080,
		Class:  sql.NullString{String: "XXL", Valid: true},
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(ScreenStats)

	if err := db.Get(stats, `SELECT * FROM "screen_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Class.String != "XXL" {
		t.Fatalf("Screen class must have been saved, but was: %v", stats)
	}
}

func TestPostgresStore_SaveCountryStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveCountryStats(nil, &CountryStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
		},
		CountryCode: sql.NullString{String: "gb", Valid: true},
	})

	if err != nil {
		t.Fatalf("Entity must have been saved, but was: %v", err)
	}

	stats := new(CountryStats)

	if err := db.Get(stats, `SELECT * FROM "country_stats"`); err != nil {
		t.Fatal(err)
	}

	stats.Visitors = 11
	err = store.SaveCountryStats(nil, stats)

	if err != nil {
		t.Fatalf("Entity must have been updated, but was: %v", err)
	}

	if err := db.Get(stats, `SELECT * FROM "country_stats"`); err != nil {
		t.Fatal(err)
	}

	if stats.Visitors != 42+11 ||
		stats.CountryCode.String != "gb" {
		t.Fatalf("Entity not as expected: %v", stats)
	}
}

func TestPostgresStore_Session(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", pastDay(2), time.Now(), "", "", "", "", "", false, false, 0, 0)
	session := store.Session(NullTenant, "fp", pastDay(1))

	if !session.IsZero() {
		t.Fatal("No session timestamp must have been found")
	}

	session = store.Session(NullTenant, "fp", pastDay(3))

	if session.IsZero() {
		t.Fatal("Session timestamp must have been found")
	}
}

func TestPostgresStore_HitDays(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 11), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 22, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	days, err := store.HitDays(NullTenant)

	if err != nil {
		t.Fatalf("Days must have been returned, but was: %v", err)
	}

	if len(days) != 2 ||
		!equalDay(days[0], day(2020, 6, 21, 0)) ||
		!equalDay(days[1], day(2020, 6, 22, 0)) {
		t.Fatalf("Days not as expected: %v", days)
	}
}

func TestPostgresStore_HitPaths(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/path", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	paths, err := store.HitPaths(NullTenant, day(2020, 6, 20, 0))

	if err != nil {
		t.Fatalf("Paths must have been returned, but was: %v", err)
	}

	if len(paths) != 0 {
		t.Fatalf("No paths must have been returned, but was: %v", len(paths))
	}

	paths, err = store.HitPaths(NullTenant, day(2020, 6, 21, 0))

	if err != nil {
		t.Fatalf("Paths must have been returned, but was: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("Two paths must have been returned, but was: %v", len(paths))
	}

	if paths[0] != "/" || paths[1] != "/path" {
		t.Fatalf("Paths not as expected: %v", paths)
	}
}

func TestPostgresStore_Paths(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/path", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	stats := &VisitorStats{
		Stats: Stats{
			Day:  day(2020, 6, 20, 7),
			Path: sql.NullString{String: "/stats", Valid: true},
		},
	}

	if err := store.SaveVisitorStats(nil, stats); err != nil {
		t.Fatal(err)
	}

	paths, err := store.Paths(NullTenant, day(2020, 6, 15, 0), day(2020, 6, 19, 0))

	if err != nil {
		t.Fatalf("Paths must have been returned, but was: %v", err)
	}

	if len(paths) != 0 {
		t.Fatalf("No paths must have been returned, but was: %v", len(paths))
	}

	paths, err = store.Paths(NullTenant, day(2020, 6, 20, 0), day(2020, 6, 25, 0))

	if err != nil {
		t.Fatalf("Paths must have been returned, but was: %v", err)
	}

	if len(paths) != 3 {
		t.Fatalf("Three paths must have been returned, but was: %v", len(paths))
	}

	if paths[0] != "/" || paths[1] != "/path" || paths[2] != "/stats" {
		t.Fatalf("Paths not as expected: %v", paths)
	}
}

func TestPostgresStore_CountVisitorsByPath(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	visitors, err := store.CountVisitorsByPath(nil, NullTenant, today(), "/", true)

	if err != nil {
		t.Fatalf("Visitors must have been returned, but was: %v", err)
	}

	if len(visitors) != 1 ||
		visitors[0].Visitors != 1 ||
		visitors[0].PlatformDesktop != 1 ||
		visitors[0].PlatformMobile != 0 ||
		visitors[0].PlatformUnknown != 0 {
		t.Fatalf("Visitors not as expected: %v", visitors)
	}
}

func TestPostgresStore_CountVisitorsByPlatform(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", pastDay(1), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", pastDay(1), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	createHit(t, store, 0, "fp2", "/", "en", "ua", "", pastDay(1), time.Time{}, "", "", "", "", "", false, true, 0, 0)
	createHit(t, store, 0, "fp2", "/", "en", "ua", "", pastDay(1), time.Time{}, "", "", "", "", "", false, true, 0, 0)
	createHit(t, store, 0, "fp3", "/", "en", "ua", "", pastDay(1), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp3", "/", "en", "ua", "", pastDay(1), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	platforms := store.CountVisitorsByPlatform(nil, NullTenant, pastDay(1))

	if platforms.PlatformDesktop != 1 ||
		platforms.PlatformMobile != 1 ||
		platforms.PlatformUnknown != 1 {
		t.Fatalf("Platforms not as expected: %v", platforms)
	}
}

func TestPostgresStore_CountVisitorsByPathAndMaxOneHit(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/", "en", "ua", "", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/page", "en", "ua", "", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp3", "/", "en", "ua", "", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	visitors := store.CountVisitorsByPathAndMaxOneHit(nil, NullTenant, pastDay(5), "/")

	if visitors != 2 {
		t.Fatalf("Two visitors must have bounced, but was: %v", visitors)
	}
}

func TestPostgresStore_ActiveVisitors(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", time.Now().Add(-time.Second*2), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/page", "en", "ua", "", time.Now().Add(-time.Second*3), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	total := store.ActiveVisitors(NullTenant, time.Now().Add(-time.Second*10))

	if total != 1 {
		t.Fatalf("One active visitor must have been returned, but was: %v", total)
	}
}

func TestPostgresStore_ActivePageVisitors(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", time.Now().Add(-time.Second*2), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/page", "en", "ua", "", time.Now().Add(-time.Second*3), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/page", "en", "ua", "", time.Now().Add(-time.Second*4), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	stats, err := store.ActivePageVisitors(NullTenant, time.Now().Add(-time.Second*10))

	if err != nil {
		t.Fatalf("Active page visitors must have been returned, but was: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("Two active page visitors must have been returned, but was: %v", len(stats))
	}

	if stats[0].Path.String != "/page" || stats[0].Visitors != 2 ||
		stats[1].Path.String != "/" || stats[1].Visitors != 1 {
		t.Fatalf("Visitor count not as expected: %v", stats)
	}
}
