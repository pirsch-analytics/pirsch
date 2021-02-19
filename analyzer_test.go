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
		assert.Equal(t, "/", visitors[0].Path.String)
		assert.Equal(t, "/path", visitors[1].Path.String)
		assert.Equal(t, 2, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
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
				Views:      71,
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
		assert.Equal(t, pastDay(3), visitors[0].Day)
		assert.Equal(t, pastDay(2), visitors[1].Day)
		assert.Equal(t, pastDay(1), visitors[2].Day)
		assert.Equal(t, today(), visitors[3].Day)
		assert.Equal(t, 0, visitors[0].Visitors)
		assert.Equal(t, 42, visitors[1].Visitors)
		assert.Equal(t, 0, visitors[2].Visitors)
		assert.Equal(t, 2, visitors[3].Visitors)
		assert.Equal(t, 0, visitors[0].Sessions)
		assert.Equal(t, 67, visitors[1].Sessions)
		assert.Equal(t, 0, visitors[2].Sessions)
		assert.Equal(t, 2, visitors[3].Sessions)
		assert.Equal(t, 0, visitors[0].Bounces)
		assert.Equal(t, 30, visitors[1].Bounces)
		assert.Equal(t, 0, visitors[2].Bounces)
		assert.Equal(t, 2, visitors[3].Bounces)
		assert.Equal(t, 0, visitors[0].Views)
		assert.Equal(t, 71, visitors[1].Views)
		assert.Equal(t, 0, visitors[2].Views)
		assert.Equal(t, 2, visitors[3].Views)
		assert.InDelta(t, 0, visitors[0].BounceRate, 0.01)
		assert.InDelta(t, 0.71, visitors[1].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[2].BounceRate, 0.01)
		assert.InDelta(t, 1, visitors[3].BounceRate, 0.01)
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
		assert.Equal(t, 5, visitors[5].Hour)
		assert.Equal(t, 12, visitors[12].Hour)
		assert.Equal(t, 43, visitors[5].Visitors)
		assert.Equal(t, 1, visitors[12].Visitors)
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
		assert.Equal(t, "de", visitors[0].Language.String)
		assert.Equal(t, "en", visitors[1].Language.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.001)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.001)
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
		assert.Equal(t, "ref2", visitors[0].Referrer.String)
		assert.Equal(t, "ref1", visitors[1].Referrer.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.Equal(t, 11, visitors[0].Bounces)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.2619, visitors[0].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
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
		assert.Equal(t, OSMac, visitors[0].OS.String)
		assert.Equal(t, OSWindows, visitors[1].OS.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, BrowserChrome, visitors[0].Browser.String)
		assert.Equal(t, BrowserFirefox, visitors[1].Browser.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, 43, visitors.PlatformDesktop)
		assert.Equal(t, 44, visitors.PlatformMobile)
		assert.Equal(t, 45, visitors.PlatformUnknown)
		assert.InDelta(t, 0.325, visitors.RelativePlatformDesktop, 0.01)
		assert.InDelta(t, 0.33, visitors.RelativePlatformMobile, 0.01)
		assert.InDelta(t, 0.34, visitors.RelativePlatformUnknown, 0.01)
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
		assert.Equal(t, 0, visitors.PlatformDesktop)
		assert.Equal(t, 0, visitors.PlatformMobile)
		assert.Equal(t, 0, visitors.PlatformUnknown)
		assert.InDelta(t, 0, visitors.RelativePlatformDesktop, 0.01)
		assert.InDelta(t, 0, visitors.RelativePlatformMobile, 0.01)
		assert.InDelta(t, 0, visitors.RelativePlatformUnknown, 0.01)
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
		assert.Equal(t, 1920, visitors[0].Width)
		assert.Equal(t, 640, visitors[1].Width)
		assert.Equal(t, 1080, visitors[0].Height)
		assert.Equal(t, 1080, visitors[1].Height)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, "XXL", visitors[0].Class.String)
		assert.Equal(t, "M", visitors[1].Class.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, "gb", visitors[0].CountryCode.String)
		assert.Equal(t, "de", visitors[1].CountryCode.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, pastDay(2), days[0].Day)
		assert.Equal(t, pastDay(1), days[1].Day)
		assert.Equal(t, today(), days[2].Day)
		assert.Equal(t, 7, days[0].Stats[9].Visitors)
		assert.Equal(t, 1, days[0].Stats[10].Visitors)
		assert.Equal(t, 11, days[0].Stats[18].Visitors)
		assert.Equal(t, 7, days[1].Stats[9].Visitors)
		assert.Equal(t, 9, days[1].Stats[18].Visitors)
		assert.Equal(t, 10, days[2].Stats[9].Visitors)
		assert.Equal(t, 1, days[2].Stats[17].Visitors)
		assert.Equal(t, 15, days[2].Stats[18].Visitors)
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
				Views:      71,
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
		assert.Equal(t, "/path", visitors[0].Path)
		assert.Equal(t, "/", visitors[1].Path)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.Equal(t, 30, visitors[0].Bounces)
		assert.Equal(t, 0, visitors[1].Bounces)
		assert.Equal(t, 72, visitors[0].Views)
		assert.Equal(t, 1, visitors[1].Views)
		assert.InDelta(t, 0.9772, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.0227, visitors[1].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.6976, visitors[0].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
		assert.InDelta(t, 0.9863, visitors[0].RelativeViews, 0.01)
		assert.InDelta(t, 0.0136, visitors[1].RelativeViews, 0.01)
		assert.Equal(t, pastDay(3), visitors[0].Stats[0].Day)
		assert.Equal(t, pastDay(2), visitors[0].Stats[1].Day)
		assert.Equal(t, pastDay(1), visitors[0].Stats[2].Day)
		assert.Equal(t, today(), visitors[0].Stats[3].Day)
		assert.Equal(t, pastDay(3), visitors[1].Stats[0].Day)
		assert.Equal(t, pastDay(2), visitors[1].Stats[1].Day)
		assert.Equal(t, pastDay(1), visitors[1].Stats[2].Day)
		assert.Equal(t, today(), visitors[1].Stats[3].Day)
		assert.Equal(t, 0, visitors[0].Stats[0].Visitors)
		assert.Equal(t, 42, visitors[0].Stats[1].Visitors)
		assert.Equal(t, 0, visitors[0].Stats[2].Visitors)
		assert.Equal(t, 1, visitors[0].Stats[3].Visitors)
		assert.Equal(t, 0, visitors[1].Stats[0].Visitors)
		assert.Equal(t, 0, visitors[1].Stats[1].Visitors)
		assert.Equal(t, 0, visitors[1].Stats[2].Visitors)
		assert.Equal(t, 1, visitors[1].Stats[3].Visitors)
		assert.Equal(t, 0, visitors[0].Stats[0].Sessions)
		assert.Equal(t, 67, visitors[0].Stats[1].Sessions)
		assert.Equal(t, 0, visitors[0].Stats[2].Sessions)
		assert.Equal(t, 1, visitors[0].Stats[3].Sessions)
		assert.Equal(t, 0, visitors[1].Stats[0].Sessions)
		assert.Equal(t, 0, visitors[1].Stats[1].Sessions)
		assert.Equal(t, 0, visitors[1].Stats[2].Sessions)
		assert.Equal(t, 1, visitors[1].Stats[3].Sessions)
		assert.Equal(t, 0, visitors[0].Stats[0].Bounces)
		assert.Equal(t, 30, visitors[0].Stats[1].Bounces)
		assert.Equal(t, 0, visitors[0].Stats[2].Bounces)
		assert.Equal(t, 0, visitors[0].Stats[3].Bounces)
		assert.Equal(t, 0, visitors[1].Stats[0].Bounces)
		assert.Equal(t, 0, visitors[1].Stats[1].Bounces)
		assert.Equal(t, 0, visitors[1].Stats[2].Bounces)
		assert.Equal(t, 0, visitors[1].Stats[3].Bounces)
		assert.Equal(t, 0, visitors[0].Stats[0].Views)
		assert.Equal(t, 71, visitors[0].Stats[1].Views)
		assert.Equal(t, 0, visitors[0].Stats[2].Views)
		assert.Equal(t, 1, visitors[0].Stats[3].Views)
		assert.Equal(t, 0, visitors[1].Stats[0].Views)
		assert.Equal(t, 0, visitors[1].Stats[1].Views)
		assert.Equal(t, 0, visitors[1].Stats[2].Views)
		assert.Equal(t, 1, visitors[1].Stats[3].Views)
		assert.InDelta(t, 0, visitors[0].Stats[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.9767, visitors[0].Stats[1].RelativeVisitors, 0.01)
		assert.InDelta(t, 0, visitors[0].Stats[2].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.0232, visitors[0].Stats[3].RelativeVisitors, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[1].RelativeVisitors, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[2].RelativeVisitors, 0.01)
		assert.InDelta(t, 1, visitors[1].Stats[3].RelativeVisitors, 0.01)
		assert.InDelta(t, 0, visitors[0].Stats[0].BounceRate, 0.01)
		assert.InDelta(t, 0.71, visitors[0].Stats[1].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[0].Stats[2].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[0].Stats[3].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[0].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[1].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[2].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[3].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[0].Stats[0].RelativeViews, 0.01)
		assert.InDelta(t, 0.9861, visitors[0].Stats[1].RelativeViews, 0.01)
		assert.InDelta(t, 0, visitors[0].Stats[2].RelativeViews, 0.01)
		assert.InDelta(t, 0.0138, visitors[0].Stats[3].RelativeViews, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[0].RelativeViews, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[1].RelativeViews, 0.01)
		assert.InDelta(t, 0, visitors[1].Stats[2].RelativeViews, 0.01)
		assert.InDelta(t, 1, visitors[1].Stats[3].RelativeViews, 0.01)
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
		assert.Equal(t, "de", visitors[0].Language.String)
		assert.Equal(t, "en", visitors[1].Language.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, "ref2", visitors[0].Referrer.String)
		assert.Equal(t, "ref1", visitors[1].Referrer.String)
		assert.Equal(t, "ref2Name", visitors[0].ReferrerName.String)
		assert.Equal(t, "ref2Icon", visitors[0].ReferrerIcon.String)
		assert.False(t, visitors[1].ReferrerName.Valid)
		assert.False(t, visitors[1].ReferrerIcon.Valid)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.2619, visitors[0].BounceRate, 0.01)
		assert.InDelta(t, 0, visitors[1].BounceRate, 0.01)
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
		assert.Equal(t, OSWindows, visitors[0].OS.String)
		assert.Equal(t, OSMac, visitors[1].OS.String)
		assert.Equal(t, 43, visitors[0].Visitors)
		assert.Equal(t, 1, visitors[1].Visitors)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, BrowserChrome, visitors[0].Browser.String)
		assert.Equal(t, BrowserFirefox, visitors[1].Browser.String)
		assert.Equal(t, 43, visitors[0].Visitors, 43)
		assert.Equal(t, 1, visitors[1].Visitors, 1)
		assert.InDelta(t, 0.977, visitors[0].RelativeVisitors, 0.01)
		assert.InDelta(t, 0.022, visitors[1].RelativeVisitors, 0.01)
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
		assert.Equal(t, 43, visitors.PlatformDesktop)
		assert.Equal(t, 44, visitors.PlatformMobile)
		assert.Equal(t, 45, visitors.PlatformUnknown)
		assert.InDelta(t, 0.325, visitors.RelativePlatformDesktop, 0.01)
		assert.InDelta(t, 0.33, visitors.RelativePlatformMobile, 0.01)
		assert.InDelta(t, 0.34, visitors.RelativePlatformUnknown, 0.01)
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
		assert.Equal(t, 0, visitors.PlatformDesktop)
		assert.Equal(t, 0, visitors.PlatformMobile)
		assert.Equal(t, 0, visitors.PlatformUnknown)
		assert.InDelta(t, 0, visitors.RelativePlatformDesktop, 0.01)
		assert.InDelta(t, 0, visitors.RelativePlatformMobile, 0.01)
		assert.InDelta(t, 0, visitors.RelativePlatformUnknown, 0.01)
	}
}

func TestAnalyzer_TimeOfDayTimezone(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)
	d := day(2020, 9, 24, 15)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", d, time.Time{}, "", "", "", "", "", false, false, 0, 0)
	tz := time.FixedZone("test", 3600*3)
	targetDate := d.In(tz)
	assert.Equal(t, 18, targetDate.Hour())
	analyzer := NewAnalyzer(store, &AnalyzerConfig{
		Timezone: tz,
	})
	visitors, _ := analyzer.TimeOfDay(&Filter{
		From: day(2020, 9, 24, 0),
		To:   day(2020, 9, 24, 0),
	})
	assert.Len(t, visitors, 1)
	assert.Len(t, visitors[0].Stats, 24)
	assert.Equal(t, day(2020, 9, 24, 0), visitors[0].Day)
	assert.Equal(t, 1, visitors[0].Stats[18].Visitors)
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
					Views:    21,
				},
			},
			{
				Stats: Stats{
					Day:      pastDay(3),
					Path:     sql.NullString{String: "/about", Valid: true},
					Visitors: 6,
					Sessions: 7,
					Bounces:  4,
					Views:    22,
				},
			},
			{
				Stats: Stats{
					Day:      pastDay(4),
					Path:     sql.NullString{String: "/home", Valid: true},
					Visitors: 2,
					Sessions: 3,
					Bounces:  1,
					Views:    23,
				},
			},
			{
				Stats: Stats{
					Day:      pastDay(5),
					Path:     sql.NullString{String: "/about", Valid: true},
					Visitors: 8,
					Sessions: 9,
					Bounces:  6,
					Views:    24,
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
		assert.Equal(t, 11, growth.Current.Visitors)
		assert.Equal(t, 13, growth.Current.Sessions)
		assert.Equal(t, 7, growth.Current.Bounces)
		assert.Equal(t, 43, growth.Current.Views)
		assert.Equal(t, 10, growth.Previous.Visitors)
		assert.Equal(t, 12, growth.Previous.Sessions)
		assert.Equal(t, 7, growth.Previous.Bounces)
		assert.Equal(t, 47, growth.Previous.Views)
		assert.InDelta(t, 0.1, growth.VisitorsGrowth, 0.01)
		assert.InDelta(t, 0.08333, growth.SessionsGrowth, 0.01)
		assert.InDelta(t, -0.0909, growth.BouncesGrowth, 0.01)
		assert.InDelta(t, -0.0851, growth.ViewsGrowth, 0.01)
		growth, err = analyzer.Growth(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(1),
			Path:     "/home",
		})
		assert.NoError(t, err)
		assert.InDelta(t, 1.5, growth.VisitorsGrowth, 0.01)
		assert.InDelta(t, 1, growth.SessionsGrowth, 0.01)
		assert.InDelta(t, 0.1999, growth.BouncesGrowth, 0.01)
		assert.InDelta(t, -0.0869, growth.ViewsGrowth, 0.01)
	}
}

func TestAnalyzer_GrowthToday(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
		createHit(t, store, tenantID, "fp2", "/", "en", "ua1", "", today(), time.Time{}, "", "", "", "", "", true, false, 0, 0)
		stats := []VisitorStats{
			{
				Stats: Stats{
					Day:      pastDay(1),
					Path:     sql.NullString{String: "/home", Valid: true},
					Visitors: 3,
					Sessions: 6,
					Bounces:  1,
					Views:    10,
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
			From:     today(),
			To:       today(),
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, growth.Current.Visitors)
		assert.Equal(t, 2, growth.Current.Sessions)
		assert.Equal(t, 2, growth.Current.Bounces)
		assert.Equal(t, 2, growth.Current.Views)
		assert.Equal(t, 3, growth.Previous.Visitors)
		assert.Equal(t, 6, growth.Previous.Sessions)
		assert.Equal(t, 1, growth.Previous.Bounces)
		assert.Equal(t, 10, growth.Previous.Views)
		assert.InDelta(t, -0.3333, growth.VisitorsGrowth, 0.01)
		assert.InDelta(t, -0.6666, growth.SessionsGrowth, 0.01)
		assert.InDelta(t, 2, growth.BouncesGrowth, 0.01)
		assert.InDelta(t, -0.8, growth.ViewsGrowth, 0.01)
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
	assert.Equal(t, 0, growth.Current.Visitors)
	assert.Equal(t, 0, growth.Current.Sessions)
	assert.Equal(t, 0, growth.Current.Bounces)
	assert.Equal(t, 0, growth.Current.Views)
	assert.Equal(t, 0, growth.Previous.Visitors)
	assert.Equal(t, 0, growth.Previous.Sessions)
	assert.Equal(t, 0, growth.Previous.Bounces)
	assert.Equal(t, 0, growth.Previous.Views)
	assert.InDelta(t, 0, growth.VisitorsGrowth, 0.01)
	assert.InDelta(t, 0, growth.SessionsGrowth, 0.01)
	assert.InDelta(t, 0, growth.BouncesGrowth, 0.01)
	assert.InDelta(t, 0, growth.ViewsGrowth, 0.01)
}

func TestAnalyzer_CalculateGrowth(t *testing.T) {
	analyzer := NewAnalyzer(newTestStore(), nil)
	growth := analyzer.calculateGrowth(0, 0)
	assert.InDelta(t, 0, growth, 0.001)
	growth = analyzer.calculateGrowth(1000, 0)
	assert.InDelta(t, 1, growth, 0.001)
	growth = analyzer.calculateGrowth(0, 1000)
	assert.InDelta(t, -1, growth, 0.001)
	growth = analyzer.calculateGrowth(100, 50)
	assert.InDelta(t, 1, growth, 0.001)
	growth = analyzer.calculateGrowth(50, 100)
	assert.InDelta(t, -0.5, growth, 0.001)
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}

func equalDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
