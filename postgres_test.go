package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestPostgresStore_SaveVisitorStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveVisitorStats(nil, &VisitorStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Path:     "/",
			Visitors: 42,
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
		stats.PlatformDesktop != 123+5 ||
		stats.PlatformMobile != 89+3 ||
		stats.PlatformUnknown != 52+1 {
		t.Fatalf("Entity not as expected: %v", stats)
	}
}

func TestPostgresStore_SaveVisitorTimeStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveVisitorTimeStats(nil, &VisitorTimeStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Path:     "/",
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
			Path:     "/",
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
}

func TestPostgresStore_SaveReferrerStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveReferrerStats(nil, &ReferrerStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Path:     "/",
			Visitors: 42,
		},
		Referrer: sql.NullString{String: "ref", Valid: true},
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
		stats.Referrer.String != "ref" {
		t.Fatalf("Entity not as expected: %v", stats)
	}
}

func TestPostgresStore_SaveOSStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveOSStats(nil, &OSStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Path:     "/",
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
}

func TestPostgresStore_SaveBrowserStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveBrowserStats(nil, &BrowserStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Path:     "/",
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
}

func TestPostgresStore_HitDays(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), "", "", "", "", false, false)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 11), "", "", "", "", false, false)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 22, 7), "", "", "", "", false, false)
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
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), "", "", "", "", false, false)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), "", "", "", "", false, false)
	createHit(t, store, 0, "fp", "/path", "en", "ua", "", day(2020, 6, 21, 7), "", "", "", "", false, false)
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
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), "", "", "", "", false, false)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), "", "", "", "", false, false)
	createHit(t, store, 0, "fp", "/path", "en", "ua", "", day(2020, 6, 21, 7), "", "", "", "", false, false)
	stats := &VisitorStats{
		Stats: Stats{
			Day:  day(2020, 6, 20, 7),
			Path: "/stats",
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
