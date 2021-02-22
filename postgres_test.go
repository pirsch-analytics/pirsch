package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPostgresStore_SaveVisitorStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveVisitorStats(nil, &VisitorStats{
		Stats: Stats{
			Day:                           day(2020, 9, 3, 0),
			Visitors:                      42,
			Sessions:                      59,
			Bounces:                       11,
			Views:                         103,
			AverageSessionDurationSeconds: 59,
		},
		PlatformDesktop: 123,
		PlatformMobile:  89,
		PlatformUnknown: 52,
	})
	assert.NoError(t, err)
	stats := new(VisitorStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "visitor_stats"`))
	stats.Visitors = 11
	stats.Sessions = 17
	stats.Bounces = 1
	stats.Views = 2
	stats.PlatformDesktop = 5
	stats.PlatformMobile = 3
	stats.PlatformUnknown = 1
	stats.AverageSessionDurationSeconds = 226
	assert.NoError(t, store.SaveVisitorStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "visitor_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
	assert.Equal(t, 59+17, stats.Sessions)
	assert.Equal(t, 11+1, stats.Bounces)
	assert.Equal(t, 103+2, stats.Views)
	assert.Equal(t, 123+5, stats.PlatformDesktop)
	assert.Equal(t, 89+3, stats.PlatformMobile)
	assert.Equal(t, 52+1, stats.PlatformUnknown)
	assert.Equal(t, 61, stats.AverageSessionDurationSeconds)
	stats.Path = sql.NullString{String: "/path", Valid: true}
	assert.NoError(t, store.SaveVisitorStats(nil, stats))
	var entries []VisitorStats
	assert.NoError(t, db.Select(&entries, `SELECT * FROM "visitor_stats"`))
	assert.Len(t, entries, 2)
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
	assert.NoError(t, err)
	stats := new(VisitorTimeStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "visitor_time_stats"`))
	stats.Visitors = 11
	assert.NoError(t, store.SaveVisitorTimeStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "visitor_time_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
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
	assert.NoError(t, err)
	stats := new(LanguageStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "language_stats"`))
	stats.Visitors = 11
	assert.NoError(t, store.SaveLanguageStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "language_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
	assert.Equal(t, "en", stats.Language.String)
	stats.Path = sql.NullString{String: "/path", Valid: true}
	assert.NoError(t, store.SaveLanguageStats(nil, stats))
	var entries []LanguageStats
	assert.NoError(t, db.Select(&entries, `SELECT * FROM "language_stats"`))
	assert.Len(t, entries, 2)
}

func TestPostgresStore_SaveReferrerStats(t *testing.T) {
	cleanupDB(t)
	db := sqlx.NewDb(postgresDB, "postgres")
	store := NewPostgresStore(postgresDB, nil)
	err := store.SaveReferrerStats(nil, &ReferrerStats{
		Stats: Stats{
			Day:      day(2020, 9, 3, 0),
			Visitors: 42,
			Bounces:  31,
		},
		Referrer:     sql.NullString{String: "ref", Valid: true},
		ReferrerName: sql.NullString{String: "Ref", Valid: true},
		ReferrerIcon: sql.NullString{String: "icon", Valid: true},
	})
	assert.NoError(t, err)
	stats := new(ReferrerStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "referrer_stats"`))
	stats.Visitors = 11
	stats.Bounces = 3
	assert.NoError(t, store.SaveReferrerStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "referrer_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
	assert.Equal(t, 31+3, stats.Bounces)
	assert.Equal(t, "ref", stats.Referrer.String)
	assert.Equal(t, "Ref", stats.ReferrerName.String)
	assert.Equal(t, "icon", stats.ReferrerIcon.String)
	stats.Path = sql.NullString{String: "/path", Valid: true}
	assert.NoError(t, store.SaveReferrerStats(nil, stats))
	var entries []ReferrerStats
	assert.NoError(t, db.Select(&entries, `SELECT * FROM "referrer_stats"`))
	assert.Len(t, entries, 2)
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
	assert.NoError(t, err)
	stats := new(OSStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "os_stats"`))
	stats.Visitors = 11
	assert.NoError(t, store.SaveOSStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "os_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
	assert.Equal(t, OSWindows, stats.OS.String)
	assert.Equal(t, "10", stats.OSVersion.String)
	stats.Path = sql.NullString{String: "/path", Valid: true}
	assert.NoError(t, store.SaveOSStats(nil, stats))
	var entries []OSStats
	assert.NoError(t, db.Select(&entries, `SELECT * FROM "os_stats"`))
	assert.Len(t, entries, 2)
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
	assert.NoError(t, err)
	stats := new(BrowserStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "browser_stats"`))
	stats.Visitors = 11
	assert.NoError(t, store.SaveBrowserStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "browser_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
	assert.Equal(t, BrowserChrome, stats.Browser.String)
	assert.Equal(t, "84.0", stats.BrowserVersion.String)
	stats.Path = sql.NullString{String: "/path", Valid: true}
	assert.NoError(t, store.SaveBrowserStats(nil, stats))
	var entries []BrowserStats
	assert.NoError(t, db.Select(&entries, `SELECT * FROM "browser_stats"`))
	assert.Len(t, entries, 2)
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
	assert.NoError(t, err)
	stats := new(ScreenStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "screen_stats"`))
	stats.Visitors = 11
	assert.NoError(t, store.SaveScreenStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "screen_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
	assert.Equal(t, 1920, stats.Width)
	assert.Equal(t, 1080, stats.Height)
	assert.False(t, stats.Class.Valid)
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
	assert.NoError(t, err)
	stats := new(ScreenStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "screen_stats"`))
	assert.Equal(t, "XXL", stats.Class.String)
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
	assert.NoError(t, err)
	stats := new(CountryStats)
	assert.NoError(t, db.Get(stats, `SELECT * FROM "country_stats"`))
	stats.Visitors = 11
	assert.NoError(t, store.SaveCountryStats(nil, stats))
	assert.NoError(t, db.Get(stats, `SELECT * FROM "country_stats"`))
	assert.Equal(t, 42+11, stats.Visitors)
	assert.Equal(t, "gb", stats.CountryCode.String)
}

func TestPostgresStore_Session(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", pastDay(2), time.Now(), "", "", "", "", "", false, false, 0, 0)
	session := store.Session(NullTenant, "fp", pastDay(1))
	assert.True(t, session.IsZero())
	session = store.Session(NullTenant, "fp", pastDay(3))
	assert.False(t, session.IsZero())
}

func TestPostgresStore_HitDays(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 11), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 22, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	days, err := store.HitDays(NullTenant)
	assert.NoError(t, err)
	assert.Len(t, days, 2)
	assert.Equal(t, day(2020, 6, 21, 0), days[0].UTC())
	assert.Equal(t, day(2020, 6, 22, 0), days[1].UTC())
}

func TestPostgresStore_HitPaths(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/path", "en", "ua", "", day(2020, 6, 21, 7), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	paths, err := store.HitPaths(NullTenant, day(2020, 6, 20, 0))
	assert.NoError(t, err)
	assert.Len(t, paths, 0)
	paths, err = store.HitPaths(NullTenant, day(2020, 6, 21, 0))
	assert.NoError(t, err)
	assert.Len(t, paths, 2)
	assert.Equal(t, "/", paths[0])
	assert.Equal(t, "/path", paths[1])
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
	assert.NoError(t, store.SaveVisitorStats(nil, stats))
	paths, err := store.Paths(NullTenant, day(2020, 6, 15, 0), day(2020, 6, 19, 0))
	assert.NoError(t, err)
	assert.Len(t, paths, 0)
	paths, err = store.Paths(NullTenant, day(2020, 6, 20, 0), day(2020, 6, 25, 0))
	assert.NoError(t, err)
	assert.Len(t, paths, 3)
	assert.Equal(t, "/", paths[0])
	assert.Equal(t, "/path", paths[1])
	assert.Equal(t, "/stats", paths[2])
}

func TestPostgresStore_CountVisitorsByPath(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
	visitors, err := store.CountVisitorsByPath(nil, NullTenant, today(), "/", true)
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[0].PlatformDesktop)
	assert.Equal(t, 0, visitors[0].PlatformMobile)
	assert.Equal(t, 0, visitors[0].PlatformUnknown)
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
	assert.Equal(t, 1, platforms.PlatformDesktop)
	assert.Equal(t, 1, platforms.PlatformMobile)
	assert.Equal(t, 1, platforms.PlatformUnknown)
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
	assert.Equal(t, 2, visitors)
}

func TestPostgresStore_CountVisitorsByReferrer(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "ref1", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "ref1", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/", "en", "ua", "ref2", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/page", "en", "ua", "ref2", pastDay(5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	stats, err := store.CountVisitorsByReferrer(nil, NullTenant, pastDay(5))
	assert.NoError(t, err)
	assert.Len(t, stats, 2)

	// ignore order...
	if stats[0].Referrer.String == "ref1" && stats[0].Bounces != 1 ||
		stats[0].Referrer.String == "ref2" && stats[0].Bounces != 0 ||
		stats[1].Referrer.String == "ref1" && stats[1].Bounces != 1 ||
		stats[1].Referrer.String == "ref2" && stats[1].Bounces != 0 {
		t.Fatalf("Stats not as expected: %v", stats)
	}
}

func TestPostgresStore_ActiveVisitors(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", time.Now().Add(-time.Second*2), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/page", "en", "ua", "", time.Now().Add(-time.Second*3), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	total := store.ActiveVisitors(NullTenant, time.Now().Add(-time.Second*10))
	assert.Equal(t, 1, total)
}

func TestPostgresStore_ActivePageVisitors(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	createHit(t, store, 0, "fp1", "/", "en", "ua", "", time.Now().Add(-time.Second*2), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/page", "en", "ua", "", time.Now().Add(-time.Second*3), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/page", "en", "ua", "", time.Now().Add(-time.Second*4), time.Time{}, "", "", "", "", "", false, false, 0, 0)
	stats, err := store.ActivePageVisitors(NullTenant, time.Now().Add(-time.Second*10))
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "/page", stats[0].Path.String)
	assert.Equal(t, "/", stats[1].Path.String)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
}

func TestPostgresStore_AverageSessionDuration(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	day := pastDay(3)
	createHit(t, store, 0, "fp", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*100), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*95), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/p3", "en", "ua", "", day.Add(time.Hour-time.Second*90), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*10), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*5), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
	assert.Equal(t, 8, store.AverageSessionDuration(nil, NullTenant, day))
}

func TestPostgresStore_AverageSessionDurationByPath(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	day := pastDay(3)
	createHit(t, store, 0, "fp1", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*100), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*95), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/p3", "en", "ua", "", day.Add(time.Hour-time.Second*90), day, "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*10), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp1", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*5), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*10), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp2", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*5), day.Add(time.Second), "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp3", "/p1", "en", "ua", "", day.Add(time.Hour-time.Second*20), day.Add(time.Second*2), "", "", "", "", "", false, false, 0, 0)
	createHit(t, store, 0, "fp3", "/p2", "en", "ua", "", day.Add(time.Hour-time.Second*5), day.Add(time.Second*2), "", "", "", "", "", false, false, 0, 0)
	assert.Equal(t, 8, store.AverageSessionDurationByPath(nil, NullTenant, day, "/p1"))
	assert.Equal(t, 5, store.AverageSessionDurationByPath(nil, NullTenant, day, "/p2"))
	assert.Equal(t, 0, store.AverageSessionDurationByPath(nil, NullTenant, day, "/p3"))
}
