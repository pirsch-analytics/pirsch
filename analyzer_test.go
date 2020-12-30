package pirsch

import (
	"database/sql"
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

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if total != 2 {
			t.Fatalf("Two active visitors must have been returned, but was: %v", total)
		}

		if len(visitors) != 2 ||
			visitors[0].Path.String != "/" || visitors[0].Visitors != 2 ||
			visitors[1].Path.String != "/path" || visitors[1].Visitors != 1 {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveVisitorStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Visitors(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 4 {
			t.Fatalf("Four visitors must have been returned, but was: %v", len(visitors))
		}

		if !visitors[0].Day.Equal(pastDay(3)) || visitors[0].Visitors != 0 || visitors[0].Sessions != 0 || visitors[0].Bounces != 0 || !inRange(visitors[0].BounceRate, 0) ||
			!visitors[1].Day.Equal(pastDay(2)) || visitors[1].Visitors != 42 || visitors[1].Sessions != 67 || visitors[1].Bounces != 30 || !inRange(visitors[1].BounceRate, 0.71) ||
			!visitors[2].Day.Equal(pastDay(1)) || visitors[2].Visitors != 0 || visitors[2].Sessions != 0 || visitors[2].Bounces != 0 || !inRange(visitors[2].BounceRate, 0) ||
			!visitors[3].Day.Equal(today()) || visitors[3].Visitors != 2 || visitors[3].Sessions != 2 || visitors[3].Bounces != 2 || !inRange(visitors[3].BounceRate, 1) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveVisitorTimeStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.VisitorHours(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 24 {
			t.Fatalf("24 visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[5].Hour != 5 || visitors[5].Visitors != 43 ||
			visitors[12].Hour != 12 || visitors[12].Visitors != 1 {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveLanguageStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Languages(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Language.String != "de" || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Language.String != "en" || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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
			},
			Referrer: sql.NullString{String: "ref2", Valid: true},
		}

		if err := store.SaveReferrerStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Referrer(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Referrer.String != "ref2" || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Referrer.String != "ref1" || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveOSStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.OS(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].OS.String != OSMac || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].OS.String != OSWindows || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveBrowserStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Browser(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Browser.String != BrowserChrome || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Browser.String != BrowserFirefox || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveVisitorStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors := analyzer.Platform(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})

		if visitors.PlatformDesktop != 43 || !inRange(visitors.RelativePlatformDesktop, 0.325) ||
			visitors.PlatformMobile != 44 || !inRange(visitors.RelativePlatformMobile, 0.33) ||
			visitors.PlatformUnknown != 45 || !inRange(visitors.RelativePlatformUnknown, 0.34) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if visitors.PlatformDesktop != 0 || !inRange(visitors.RelativePlatformDesktop, 0.001) ||
			visitors.PlatformMobile != 0 || !inRange(visitors.RelativePlatformMobile, 0.001) ||
			visitors.PlatformUnknown != 0 || !inRange(visitors.RelativePlatformUnknown, 0.001) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveScreenStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Screen(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Width != 1920 || visitors[0].Height != 1080 || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Width != 640 || visitors[1].Height != 1080 || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveScreenStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.ScreenClass(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Class.String != "XXL" || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Class.String != "M" || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveCountryStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.Country(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(4),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].CountryCode.String != "gb" || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].CountryCode.String != "de" || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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
			if err := store.SaveVisitorTimeStats(nil, &s); err != nil {
				t.Fatal(err)
			}
		}

		analyzer := NewAnalyzer(store, nil)
		days, err := analyzer.TimeOfDay(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(2),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(days) != 3 {
			t.Fatalf("Three results must have been returned, but was: %v", len(days))
		}

		if !days[0].Day.Equal(pastDay(2)) ||
			!days[1].Day.Equal(pastDay(1)) ||
			!days[2].Day.Equal(today()) {
			t.Fatalf("Days not as expected: %v", days)
		}

		if days[0].Stats[9].Visitors != 7 || days[0].Stats[10].Visitors != 1 || days[0].Stats[18].Visitors != 11 {
			t.Fatalf("First day not as expected: %v", days[0])
		}

		if days[1].Stats[9].Visitors != 7 || days[1].Stats[18].Visitors != 9 {
			t.Fatalf("Second day not as expected: %v", days[1])
		}

		if days[2].Stats[9].Visitors != 10 || days[2].Stats[17].Visitors != 1 || days[2].Stats[18].Visitors != 15 {
			t.Fatalf("Third day not as expected: %v", days[2])
		}
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

		if err := store.SaveVisitorStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageVisitors(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if len(visitors[0].Stats) != 4 || visitors[0].Path != "/" ||
			!visitors[0].Stats[0].Day.Equal(pastDay(3)) || visitors[0].Stats[0].Visitors != 0 || visitors[0].Stats[0].Sessions != 0 || visitors[0].Stats[0].Bounces != 0 || !inRange(visitors[0].Stats[0].BounceRate, 0) ||
			!visitors[0].Stats[1].Day.Equal(pastDay(2)) || visitors[0].Stats[1].Visitors != 0 || visitors[0].Stats[1].Sessions != 0 || visitors[0].Stats[1].Bounces != 0 || !inRange(visitors[0].Stats[1].BounceRate, 0) ||
			!visitors[0].Stats[2].Day.Equal(pastDay(1)) || visitors[0].Stats[2].Visitors != 0 || visitors[0].Stats[2].Sessions != 0 || visitors[0].Stats[2].Bounces != 0 || !inRange(visitors[0].Stats[2].BounceRate, 0) ||
			!visitors[0].Stats[3].Day.Equal(today()) || visitors[0].Stats[3].Visitors != 1 || visitors[0].Stats[3].Sessions != 1 || visitors[0].Stats[3].Bounces != 0 || !inRange(visitors[0].Stats[3].BounceRate, 0) {
			t.Fatalf("First path not as expected: %v", visitors)
		}

		if len(visitors[1].Stats) != 4 || visitors[1].Path != "/path" ||
			!visitors[1].Stats[0].Day.Equal(pastDay(3)) || visitors[1].Stats[0].Visitors != 0 || visitors[1].Stats[0].Sessions != 0 || visitors[1].Stats[0].Bounces != 0 || !inRange(visitors[1].Stats[0].BounceRate, 0) ||
			!visitors[1].Stats[1].Day.Equal(pastDay(2)) || visitors[1].Stats[1].Visitors != 42 || visitors[1].Stats[1].Sessions != 67 || visitors[1].Stats[1].Bounces != 30 || !inRange(visitors[1].Stats[1].BounceRate, 0.71) ||
			!visitors[1].Stats[2].Day.Equal(pastDay(1)) || visitors[1].Stats[2].Visitors != 0 || visitors[1].Stats[2].Sessions != 0 || visitors[1].Stats[2].Bounces != 0 || !inRange(visitors[1].Stats[2].BounceRate, 0) ||
			!visitors[1].Stats[3].Day.Equal(today()) || visitors[1].Stats[3].Visitors != 1 || visitors[1].Stats[3].Sessions != 1 || visitors[1].Stats[3].Bounces != 0 || !inRange(visitors[1].Stats[3].BounceRate, 0) {
			t.Fatalf("Second path not as expected: %v", visitors)
		}
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

		if err := store.SaveLanguageStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageLanguages(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Language.String != "de" || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Language.String != "en" || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
	}
}

func TestAnalyzer_PageReferrer(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		store := NewPostgresStore(postgresDB, nil)
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "ref1", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "ref2", today(), time.Time{}, "", "", "", "", "", false, false, 0, 0)
		stats := &ReferrerStats{
			Stats: Stats{
				BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
				Day:        pastDay(2),
				Path:       sql.NullString{String: "/path", Valid: true},
				Visitors:   42,
			},
			Referrer: sql.NullString{String: "ref2", Valid: true},
		}

		if err := store.SaveReferrerStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageReferrer(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Referrer.String != "ref2" || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Referrer.String != "ref1" || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveOSStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageOS(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].OS.String != OSWindows || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].OS.String != OSMac || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveBrowserStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors, err := analyzer.PageBrowser(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
		}

		if visitors[0].Browser.String != BrowserChrome || visitors[0].Visitors != 43 || !inRange(visitors[0].RelativeVisitors, 0.977) ||
			visitors[1].Browser.String != BrowserFirefox || visitors[1].Visitors != 1 || !inRange(visitors[1].RelativeVisitors, 0.022) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if err := store.SaveVisitorStats(nil, stats); err != nil {
			t.Fatal(err)
		}

		analyzer := NewAnalyzer(store, nil)
		visitors := analyzer.PagePlatform(&Filter{
			TenantID: NewTenantID(tenantID),
			Path:     "/path",
			From:     pastDay(3),
			To:       today(),
		})

		if visitors.PlatformDesktop != 43 || !inRange(visitors.RelativePlatformDesktop, 0.325) ||
			visitors.PlatformMobile != 44 || !inRange(visitors.RelativePlatformMobile, 0.33) ||
			visitors.PlatformUnknown != 45 || !inRange(visitors.RelativePlatformUnknown, 0.34) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
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

		if visitors.PlatformDesktop != 0 || !inRange(visitors.RelativePlatformDesktop, 0.001) ||
			visitors.PlatformMobile != 0 || !inRange(visitors.RelativePlatformMobile, 0.001) ||
			visitors.PlatformUnknown != 0 || !inRange(visitors.RelativePlatformUnknown, 0.001) {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
	}
}

func TestAnalyzer_TimeOfDayTimezone(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	cleanupDB(t)
	d := day(2020, 9, 24, 15)
	createHit(t, store, 0, "fp", "/", "en", "ua", "", d, time.Time{}, "", "", "", "", "", false, false, 0, 0)
	tz := time.FixedZone("test", 3600*3)
	targetDate := d.In(tz)

	if targetDate.Hour() != 18 {
		t.Fatalf("Fixed time not as expected: %v", targetDate)
	}

	analyzer := NewAnalyzer(store, &AnalyzerConfig{
		Timezone: tz,
	})
	visitors, _ := analyzer.TimeOfDay(&Filter{
		From: day(2020, 9, 24, 0),
		To:   day(2020, 9, 24, 0),
	})

	if len(visitors) != 1 || len(visitors[0].Stats) != 24 ||
		!visitors[0].Day.Equal(day(2020, 9, 24, 0)) ||
		visitors[0].Stats[18].Visitors != 1 {
		t.Fatalf("Visitors not as expected: %v", visitors)
	}
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

			if err := store.SaveVisitorStats(nil, &s); err != nil {
				t.Fatal(err)
			}
		}

		// save again without paths for processed statistics
		for _, s := range stats {
			s.BaseEntity.TenantID = NewTenantID(tenantID)
			s.Path.Valid = false

			if err := store.SaveVisitorStats(nil, &s); err != nil {
				t.Fatal(err)
			}
		}

		analyzer := NewAnalyzer(store, nil)
		growth, err := analyzer.Growth(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(1),
		})

		if err != nil {
			t.Fatalf("Growth must be returned, but was: %v", err)
		}

		if growth.Current.Visitors != 11 ||
			growth.Current.Sessions != 13 ||
			growth.Current.Bounces != 7 {
			t.Fatalf("Current sums not as expected: %v", growth.Current)
		}

		if growth.Previous.Visitors != 10 ||
			growth.Previous.Sessions != 12 ||
			growth.Previous.Bounces != 7 {
			t.Fatalf("Previous sums not as expected: %v", growth.Current)
		}

		if !inRange(growth.VisitorsGrowth, 0.1) ||
			!inRange(growth.SessionsGrowth, 0.08333) ||
			!inRange(growth.BouncesGrowth, -0.0909) {
			t.Fatalf("Growth not as expected: %v %v %v", growth.VisitorsGrowth, growth.SessionsGrowth, growth.BouncesGrowth)
		}

		growth, err = analyzer.Growth(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(1),
			Path:     "/home",
		})

		if err != nil {
			t.Fatalf("Growth for path must be returned, but was: %v", err)
		}

		if !inRange(growth.VisitorsGrowth, 1.5) ||
			!inRange(growth.SessionsGrowth, 1) ||
			!inRange(growth.BouncesGrowth, 0.1999) {
			t.Fatalf("Growth for path not as expected: %v %v %v", growth.VisitorsGrowth, growth.SessionsGrowth, growth.BouncesGrowth)
		}
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

	if err != nil {
		t.Fatalf("Growth must be returned, but was: %v", err)
	}

	if growth.Current.Visitors != 0 ||
		growth.Current.Sessions != 0 ||
		growth.Current.Bounces != 0 {
		t.Fatalf("Current sums not as expected: %v", growth.Current)
	}

	if growth.Previous.Visitors != 0 ||
		growth.Previous.Sessions != 0 ||
		growth.Previous.Bounces != 0 {
		t.Fatalf("Previous sums not as expected: %v", growth.Current)
	}

	if !inRange(growth.VisitorsGrowth, 0) ||
		!inRange(growth.SessionsGrowth, 0) ||
		!inRange(growth.BouncesGrowth, 0) {
		t.Fatalf("Growth not as expected: %v %v %v", growth.VisitorsGrowth, growth.SessionsGrowth, growth.BouncesGrowth)
	}
}

func TestAnalyzer_CalculateGrowth(t *testing.T) {
	analyzer := NewAnalyzer(newTestStore(), nil)

	if growth := analyzer.calculateGrowth(0, 0); !inRange(growth, 0) {
		t.Fatalf("Growth must be zero, but was: %v", growth)
	}

	if growth := analyzer.calculateGrowth(1000, 0); !inRange(growth, 1) {
		t.Fatalf("Growth must be 1, but was: %v", growth)
	}

	if growth := analyzer.calculateGrowth(0, 1000); !inRange(growth, -1) {
		t.Fatalf("Growth must be -1, but was: %v", growth)
	}

	if growth := analyzer.calculateGrowth(100, 50); !inRange(growth, 1) {
		t.Fatalf("Growth must be 1, but was: %v", growth)
	}

	if growth := analyzer.calculateGrowth(50, 100); !inRange(growth, -0.5) {
		t.Fatalf("Growth must be -0.5, but was: %v", growth)
	}
}

func pastDay(n int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}

func equalDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func inRange(f, target float64) bool {
	return f > target-0.01 && f < target+0.01
}
