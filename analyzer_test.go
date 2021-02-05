package pirsch

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_ActiveVisitors(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", time.Now().UTC().Add(-time.Second*10), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", time.Now().UTC().Add(-time.Second*11), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp2", "/", "en", "ua2", "", time.Now().UTC().Add(-time.Second*31), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp3", "/", "en", "ua3", "", time.Now().UTC().Add(-time.Second*20), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp3", "/path", "en", "ua3", "", time.Now().UTC().Add(-time.Second*28), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		analyzer := NewAnalyzer(store, nil)
		visitors, total, err := analyzer.ActiveVisitors(&Filter{
			TenantID: NewTenantID(tenantID),
		}, time.Second*30)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Path.String, "/")
		assert.Equal(t, visitors[1].Path.String, "/path")
		assert.Equal(t, visitors[0].Visitors, 2)
		assert.Equal(t, visitors[1].Visitors, 1)
	}
}

func TestAnalyzer_Visitors(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp2", "/path", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &VisitorStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
				Sessions:   67,
				Bounces:    30,
			},
		}
		assert.NoError(t, store.SaveVisitorStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Visitors(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 4)
		assert.Equal(t, visitors[0].Day, pastDay(3))
		assert.Equal(t, visitors[1].Day, pastDay(2))
		assert.Equal(t, visitors[2].Day, pastDay(1))
		assert.Equal(t, visitors[3].Day, today())
		assert.Equal(t, visitors[0].Visitors, 0)
		assert.Equal(t, visitors[1].Visitors, 42)
		assert.Equal(t, visitors[2].Visitors, 0)
		assert.Equal(t, visitors[3].Visitors, 2)
		assert.Equal(t, visitors[0].Sessions, 0)
		assert.Equal(t, visitors[1].Sessions, 67)
		assert.Equal(t, visitors[2].Sessions, 0)
		assert.Equal(t, visitors[3].Sessions, 2)
		assert.Equal(t, visitors[0].Bounces, 0)
		assert.Equal(t, visitors[1].Bounces, 30)
		assert.Equal(t, visitors[2].Bounces, 0)
		assert.Equal(t, visitors[3].Bounces, 2)
		assert.InDelta(t, visitors[0].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[1].BounceRate, 0.71, 0.01)
		assert.InDelta(t, visitors[2].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[3].BounceRate, 1, 0.01)
	}
}

func TestAnalyzer_VisitorHours(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today().Add(time.Hour*5), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp2", "/path", "en", "ua1", "", today().Add(time.Hour*12), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &VisitorTimeStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
			},
			Hour: 5,
		}
		assert.NoError(t, store.SaveVisitorTimeStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.VisitorHours(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 24)
		assert.Equal(t, visitors[5].Hour, 5)
		assert.Equal(t, visitors[12].Hour, 12)
		assert.Equal(t, visitors[5].Visitors, 43)
		assert.Equal(t, visitors[12].Visitors, 1)
	}
}

func TestAnalyzer_Languages(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &LanguageStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
			},
			Language: sql.NullString{String: "de", Valid: true},
		}
		assert.NoError(t, store.SaveLanguageStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Languages(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Language.String, "de")
		assert.Equal(t, visitors[1].Language.String, "en")
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.001)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.001)
	}
}

func TestAnalyzer_Referrer(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "ref2", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &ReferrerStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
				Bounces:    11,
			},
			Referrer: sql.NullString{String: "ref2", Valid: true},
		}
		assert.NoError(t, store.SaveReferrerStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Referrer(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Referrer.String, "ref2")
		assert.Equal(t, visitors[1].Referrer.String, "ref1")
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.Equal(t, visitors[0].Bounces, 11)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
		assert.InDelta(t, visitors[0].BounceRate, 0.2619, 0.01)
		assert.InDelta(t, visitors[1].BounceRate, 0, 0.01)
	}
}

func TestAnalyzer_OS(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, OSWindows, "10", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, OSMac, "10.15.3", "", "", "", false, false, 0, 0)
		stats := &OSStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
			},
			OS:        sql.NullString{String: OSMac, Valid: true},
			OSVersion: sql.NullString{String: "10.14.1", Valid: true},
		}
		assert.NoError(t, store.SaveOSStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.OS(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].OS.String, OSMac)
		assert.Equal(t, visitors[1].OS.String, OSWindows)
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_Browser(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", BrowserChrome, "84.0", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", BrowserFirefox, "54.0", "", false, false, 0, 0)
		stats := &BrowserStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
			},
			Browser:        sql.NullString{String: BrowserChrome, Valid: true},
			BrowserVersion: sql.NullString{String: "83.1", Valid: true},
		}
		assert.NoError(t, store.SaveBrowserStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Browser(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Browser.String, BrowserChrome)
		assert.Equal(t, visitors[1].Browser.String, BrowserFirefox)
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_Platform(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, true, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &VisitorStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
			},
			PlatformDesktop: 42,
			PlatformMobile:  43,
			PlatformUnknown: 44,
		}
		assert.NoError(t, store.SaveVisitorStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors := analyzer.Platform(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})
		assert.Equal(t, visitors.PlatformDesktop, 43)
		assert.Equal(t, visitors.PlatformMobile, 44)
		assert.Equal(t, visitors.PlatformUnknown, 45)
		assert.InDelta(t, visitors.RelativePlatformDesktop, 0.325, 0.01)
		assert.InDelta(t, visitors.RelativePlatformMobile, 0.33, 0.01)
		assert.InDelta(t, visitors.RelativePlatformUnknown, 0.34, 0.01)
	}
}

func TestAnalyzer_PlatformNoData(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		analyzer := NewAnalyzer(store, nil)
		visitors := analyzer.Platform(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})
		assert.Equal(t, visitors.PlatformDesktop, 0)
		assert.Equal(t, visitors.PlatformMobile, 0)
		assert.Equal(t, visitors.PlatformUnknown, 0)
		assert.InDelta(t, visitors.RelativePlatformDesktop, 0, 0.01)
		assert.InDelta(t, visitors.RelativePlatformMobile, 0, 0.01)
		assert.InDelta(t, visitors.RelativePlatformUnknown, 0, 0.01)
	}
}

func TestAnalyzer_Screen(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 1920, 1080)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 640, 1080)
		stats := &ScreenStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
			},
			Width:  1920,
			Height: 1080,
		}
		assert.NoError(t, store.SaveScreenStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Screen(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Width, 1920)
		assert.Equal(t, visitors[1].Width, 640)
		assert.Equal(t, visitors[0].Height, 1080)
		assert.Equal(t, visitors[1].Height, 1080)
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_ScreenClass(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 1920, 1080)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 640, 1080)
		stats := &ScreenStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
			},
			Width:  1920,
			Height: 1080,
			Class:  sql.NullString{String: "XXL", Valid: true},
		}
		assert.NoError(t, store.SaveScreenStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.ScreenClass(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Class.String, "XXL")
		assert.Equal(t, visitors[1].Class.String, "M")
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_Country(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "de", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", "", "", "gb", false, false, 0, 0)
		stats := &CountryStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Visitors:   42,
			},
			CountryCode: sql.NullString{String: "gb", Valid: true},
		}
		assert.NoError(t, store.SaveCountryStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Country(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].CountryCode.String, "gb")
		assert.Equal(t, visitors[1].CountryCode.String, "de")
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_TimeOfDay(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", pastDay(1).Add(time.Hour*9), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", pastDay(2).Add(time.Hour*10), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today().Add(time.Hour*17), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today().Add(time.Hour*18), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := []VisitorTimeStats{
			{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Visitors:   7,
					Sessions:   8,
				},
				Hour: 9,
			},
			{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Visitors:   11,
					Sessions:   12,
				},
				Hour: 18,
			},
			{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(1),
					Visitors:   6,
					Sessions:   7,
				},
				Hour: 9,
			},
			{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(1),
					Visitors:   9,
					Sessions:   10,
				},
				Hour: 18,
			},
			{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        today(),
					Visitors:   10,
					Sessions:   11,
				},
				Hour: 9,
			},
			{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        today(),
					Visitors:   14,
					Sessions:   15,
				},
				Hour: 18,
			},
		}

		for _, s := range stats {
			assert.NoError(t, store.SaveVisitorTimeStats(nil, &s))
		}

		analyzer := NewAnalyzer(store, nil)
		days, err := analyzer.TimeOfDay(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(2),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, days, 3)
		assert.Equal(t, days[0].Day, pastDay(2))
		assert.Equal(t, days[1].Day, pastDay(1))
		assert.Equal(t, days[2].Day, today())
		assert.Equal(t, days[0].Stats[9].Visitors, 7)
		assert.Equal(t, days[0].Stats[10].Visitors, 1)
		assert.Equal(t, days[0].Stats[18].Visitors, 11)
		assert.Equal(t, days[1].Stats[9].Visitors, 7)
		assert.Equal(t, days[1].Stats[18].Visitors, 9)
		assert.Equal(t, days[2].Stats[9].Visitors, 10)
		assert.Equal(t, days[2].Stats[17].Visitors, 1)
		assert.Equal(t, days[2].Stats[18].Visitors, 15)
	}
}

func TestAnalyzer_PageVisitors(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &VisitorStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Path:       sql.NullString{String: "/path", Valid: true},
				Visitors:   42,
				Sessions:   67,
				Bounces:    30,
			},
		}
		assert.NoError(t, store.SaveVisitorStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageVisitors(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Len(t, visitors[0].Stats, 4)
		assert.Len(t, visitors[1].Stats, 4)
		assert.Equal(t, visitors[0].Path, "/")
		assert.Equal(t, visitors[1].Path, "/path")
		assert.Equal(t, visitors[0].Stats[0].Day, pastDay(3))
		assert.Equal(t, visitors[0].Stats[1].Day, pastDay(2))
		assert.Equal(t, visitors[0].Stats[2].Day, pastDay(1))
		assert.Equal(t, visitors[0].Stats[3].Day, today())
		assert.Equal(t, visitors[1].Stats[0].Day, pastDay(3))
		assert.Equal(t, visitors[1].Stats[1].Day, pastDay(2))
		assert.Equal(t, visitors[1].Stats[2].Day, pastDay(1))
		assert.Equal(t, visitors[1].Stats[3].Day, today())
		assert.Equal(t, visitors[0].Stats[0].Visitors, 0)
		assert.Equal(t, visitors[0].Stats[1].Visitors, 0)
		assert.Equal(t, visitors[0].Stats[2].Visitors, 0)
		assert.Equal(t, visitors[0].Stats[3].Visitors, 1)
		assert.Equal(t, visitors[1].Stats[0].Visitors, 0)
		assert.Equal(t, visitors[1].Stats[1].Visitors, 42)
		assert.Equal(t, visitors[1].Stats[2].Visitors, 0)
		assert.Equal(t, visitors[1].Stats[3].Visitors, 1)
		assert.Equal(t, visitors[0].Stats[0].Sessions, 0)
		assert.Equal(t, visitors[0].Stats[1].Sessions, 0)
		assert.Equal(t, visitors[0].Stats[2].Sessions, 0)
		assert.Equal(t, visitors[0].Stats[3].Sessions, 1)
		assert.Equal(t, visitors[1].Stats[0].Sessions, 0)
		assert.Equal(t, visitors[1].Stats[1].Sessions, 67)
		assert.Equal(t, visitors[1].Stats[2].Sessions, 0)
		assert.Equal(t, visitors[1].Stats[3].Sessions, 1)
		assert.Equal(t, visitors[0].Stats[0].Bounces, 0)
		assert.Equal(t, visitors[0].Stats[1].Bounces, 0)
		assert.Equal(t, visitors[0].Stats[2].Bounces, 0)
		assert.Equal(t, visitors[0].Stats[3].Bounces, 0)
		assert.Equal(t, visitors[1].Stats[0].Bounces, 0)
		assert.Equal(t, visitors[1].Stats[1].Bounces, 30)
		assert.Equal(t, visitors[1].Stats[2].Bounces, 0)
		assert.Equal(t, visitors[1].Stats[3].Bounces, 0)
		assert.InDelta(t, visitors[0].Stats[0].RelativeVisitors, 0, 0.01)
		assert.InDelta(t, visitors[0].Stats[1].RelativeVisitors, 0, 0.01)
		assert.InDelta(t, visitors[0].Stats[2].RelativeVisitors, 0, 0.01)
		assert.InDelta(t, visitors[0].Stats[3].RelativeVisitors, 1, 0.01)
		assert.InDelta(t, visitors[1].Stats[0].RelativeVisitors, 0, 0.01)
		assert.InDelta(t, visitors[1].Stats[1].RelativeVisitors, 0.9767, 0.01)
		assert.InDelta(t, visitors[1].Stats[2].RelativeVisitors, 0, 0.01)
		assert.InDelta(t, visitors[1].Stats[3].RelativeVisitors, 0.0232, 0.01)
		assert.InDelta(t, visitors[0].Stats[0].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[0].Stats[1].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[0].Stats[2].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[0].Stats[3].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[1].Stats[0].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[1].Stats[1].BounceRate, 0.71, 0.01)
		assert.InDelta(t, visitors[1].Stats[2].BounceRate, 0, 0.01)
		assert.InDelta(t, visitors[1].Stats[3].BounceRate, 0, 0.01)
	}
}

func TestAnalyzer_PageLanguages(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &LanguageStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Path:       sql.NullString{String: "/path", Valid: true},
				Visitors:   42,
			},
			Language: sql.NullString{String: "de", Valid: true},
		}
		assert.NoError(t, store.SaveLanguageStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageLanguages(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Language.String, "de")
		assert.Equal(t, visitors[1].Language.String, "en")
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_PageReferrer(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "ref1", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		hit := Hit{
			BaseEntity:   BaseEntity{TenantID: NewTenantID(tenantID)},
			Fingerprint:  "fp1",
			Path:         "/path",
			Language:     sql.NullString{String: "en", Valid: true},
			UserAgent:    sql.NullString{String: "ua1", Valid: true},
			Referrer:     sql.NullString{String: "ref2", Valid: true},
			ReferrerName: sql.NullString{String: "ref2Name", Valid: true},
			ReferrerIcon: sql.NullString{String: "ref2Icon", Valid: true},
			Time:         today(),
		}
		assert.NoError(t, store.SaveHits([]Hit{hit}))
		stats := &ReferrerStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Path:       sql.NullString{String: "/path", Valid: true},
				Visitors:   42,
				Bounces:    11,
			},
			Referrer:     sql.NullString{String: "ref2", Valid: true},
			ReferrerName: sql.NullString{String: "ref2Name", Valid: true},
			ReferrerIcon: sql.NullString{String: "ref2Icon", Valid: true},
		}
		assert.NoError(t, store.SaveReferrerStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageReferrer(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Referrer.String, "ref2")
		assert.Equal(t, visitors[1].Referrer.String, "ref1")
		assert.Equal(t, visitors[0].ReferrerName.String, "ref2Name")
		assert.Equal(t, visitors[0].ReferrerIcon.String, "ref2Icon")
		assert.False(t, visitors[1].ReferrerName.Valid)
		assert.False(t, visitors[1].ReferrerIcon.Valid)
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
		assert.InDelta(t, visitors[0].BounceRate, 0.2619, 0.01)
		assert.InDelta(t, visitors[1].BounceRate, 0, 0.01)
	}
}

func TestAnalyzer_PageOS(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, OSMac, "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, OSMac, "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, OSWindows, "", "", "", "", false, false, 0, 0)
		stats := &OSStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Path:       sql.NullString{String: "/path", Valid: true},
				Visitors:   42,
			},
			OS: sql.NullString{String: OSWindows, Valid: true},
		}
		assert.NoError(t, store.SaveOSStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageOS(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].OS.String, OSWindows)
		assert.Equal(t, visitors[1].OS.String, OSMac)
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_PageBrowser(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", BrowserFirefox, "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, "", "", BrowserFirefox, "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, "", "", BrowserChrome, "", "", false, false, 0, 0)
		stats := &BrowserStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Path:       sql.NullString{String: "/path", Valid: true},
				Visitors:   42,
			},
			Browser: sql.NullString{String: BrowserChrome, Valid: true},
		}
		assert.NoError(t, store.SaveBrowserStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageBrowser(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Len(t, visitors, 2)
		assert.Equal(t, visitors[0].Browser.String, BrowserChrome)
		assert.Equal(t, visitors[1].Browser.String, BrowserFirefox)
		assert.Equal(t, visitors[0].Visitors, 43)
		assert.Equal(t, visitors[1].Visitors, 1)
		assert.InDelta(t, visitors[0].RelativeVisitors, 0.977, 0.01)
		assert.InDelta(t, visitors[1].RelativeVisitors, 0.022, 0.01)
	}
}

func TestAnalyzer_PagePlatform(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, true, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &VisitorStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Path:       sql.NullString{String: "/path", Valid: true},
			},
			PlatformDesktop: 42,
			PlatformMobile:  43,
			PlatformUnknown: 44,
		}
		assert.NoError(t, store.SaveVisitorStats(nil, stats))
		analyzer := NewAnalyzer(store, nil)
		visitors := analyzer.PagePlatform(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})
		assert.Equal(t, visitors.PlatformDesktop, 43)
		assert.Equal(t, visitors.PlatformMobile, 44)
		assert.Equal(t, visitors.PlatformUnknown, 45)
		assert.InDelta(t, visitors.RelativePlatformDesktop, 0.325, 0.01)
		assert.InDelta(t, visitors.RelativePlatformMobile, 0.33, 0.01)
		assert.InDelta(t, visitors.RelativePlatformUnknown, 0.34, 0.01)
	}
}

func TestAnalyzer_PagePlatformNoData(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		analyzer := NewAnalyzer(store, nil)
		visitors := analyzer.PagePlatform(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})
		assert.Equal(t, visitors.PlatformDesktop, 0)
		assert.Equal(t, visitors.PlatformMobile, 0)
		assert.Equal(t, visitors.PlatformUnknown, 0)
		assert.InDelta(t, visitors.RelativePlatformDesktop, 0, 0.01)
		assert.InDelta(t, visitors.RelativePlatformMobile, 0, 0.01)
		assert.InDelta(t, visitors.RelativePlatformUnknown, 0, 0.01)
	}
}

func TestAnalyzer_TimeOfDayTimezone(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)
	d := day(2020, 9, 24, 15)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", d, time.Time{}, "", "", "", "", "", false, false, 0, 0)
	tz := time.FixedZone("test", 3600*3)
	targetDate := d.In(tz)
	assert.Equal(t, targetDate.Hour(), 18)
	analyzer := NewAnalyzer(store, &AnalyzerConfig{
		Timezone: tz,
	})
	visitors, _ := analyzer.TimeOfDay(&Filter{
		From: day(2020, 9, 24, 0),
		To:   day(2020, 9, 24, 0),
	})
	assert.Len(t, visitors, 1)
	assert.Len(t, visitors[0].Stats, 24)
	assert.Equal(t, visitors[0].Day, day(2020, 9, 24, 0))
	assert.Equal(t, visitors[0].Stats[18].Visitors, 1)
}

func TestAnalyzer_Growth(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		stats := []VisitorStats{
			{
				Stats: Stats{
					Day:      pastDay(2),
					Path:     sql.NullString{String: "/home", Valid: true},
					Visitors: 5,
					Sessions: 6,
					Bounces:  3,
				},
			},
			{
				Stats: Stats{
					Day:      pastDay(3),
					Path:     sql.NullString{String: "/about", Valid: true},
					Visitors: 6,
					Sessions: 7,
					Bounces:  4,
				},
			},
			{
				Stats: Stats{
					Day:      pastDay(4),
					Path:     sql.NullString{String: "/home", Valid: true},
					Visitors: 2,
					Sessions: 3,
					Bounces:  1,
				},
			},
			{
				Stats: Stats{
					Day:      pastDay(5),
					Path:     sql.NullString{String: "/about", Valid: true},
					Visitors: 8,
					Sessions: 9,
					Bounces:  6,
				},
			},
		}

		for _, s := range stats {
			s.BaseEntity.TenantID = NewTenantID(tenantID)
			assert.NoError(t, store.SaveVisitorStats(nil, &s))
		}

		// save again without paths for processed statistics
		for _, s := range stats {
			s.BaseEntity.TenantID = NewTenantID(tenantID)
			s.Path.Valid = false
			assert.NoError(t, store.SaveVisitorStats(nil, &s))
		}

		analyzer := NewAnalyzer(store, nil)
		growth, err := analyzer.Growth(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(1),
		})
		assert.NoError(t, err)
		assert.Equal(t, growth.Current.Visitors, 11)
		assert.Equal(t, growth.Current.Sessions, 13)
		assert.Equal(t, growth.Current.Bounces, 7)
		assert.Equal(t, growth.Previous.Visitors, 10)
		assert.Equal(t, growth.Previous.Sessions, 12)
		assert.Equal(t, growth.Previous.Bounces, 7)
		assert.InDelta(t, growth.VisitorsGrowth, 0.1, 0.01)
		assert.InDelta(t, growth.SessionsGrowth, 0.08333, 0.01)
		assert.InDelta(t, growth.BouncesGrowth, -0.0909, 0.01)
		growth, err = analyzer.Growth(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(1),
			Path:     "/home",
		})
		assert.NoError(t, err)
		assert.InDelta(t, growth.VisitorsGrowth, 1.5, 0.01)
		assert.InDelta(t, growth.SessionsGrowth, 1, 0.01)
		assert.InDelta(t, growth.BouncesGrowth, 0.1999, 0.01)
	}
}

func TestAnalyzer_GrowthNoData(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)
	analyzer := NewAnalyzer(store, nil)
	growth, err := analyzer.Growth(&Filter{
		From: pastDay(3),
		To:   pastDay(1),
	})
	assert.NoError(t, err)
	assert.Equal(t, growth.Current.Visitors, 0)
	assert.Equal(t, growth.Current.Sessions, 0)
	assert.Equal(t, growth.Current.Bounces, 0)
	assert.Equal(t, growth.Previous.Visitors, 0)
	assert.Equal(t, growth.Previous.Sessions, 0)
	assert.Equal(t, growth.Previous.Bounces, 0)
	assert.InDelta(t, growth.VisitorsGrowth, 0, 0.01)
	assert.InDelta(t, growth.SessionsGrowth, 0, 0.01)
	assert.InDelta(t, growth.BouncesGrowth, 0, 0.01)
}

func TestAnalyzer_CalculateGrowth(t *testing.T) {
	analyzer := NewAnalyzer(newTestStore(), nil)
	growth := analyzer.calculateGrowth(0, 0)
	assert.InDelta(t, growth, 0, 0.001)
	growth = analyzer.calculateGrowth(1000, 0)
	assert.InDelta(t, growth, 1, 0.001)
	growth = analyzer.calculateGrowth(0, 1000)
	assert.InDelta(t, growth, -1, 0.001)
	growth = analyzer.calculateGrowth(100, 50)
	assert.InDelta(t, growth, 1, 0.001)
	growth = analyzer.calculateGrowth(50, 100)
	assert.InDelta(t, growth, -0.5, 0.001)
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}

func equalDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
