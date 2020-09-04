package pirsch

import (
	"database/sql"
	"testing"
	"time"
)

func TestAnalyzer_ActiveVisitors(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", time.Now().UTC().Add(-time.Second*10), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", time.Now().UTC().Add(-time.Second*11), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp2", "/", "en", "ua2", "", time.Now().UTC().Add(-time.Second*31), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp3", "/", "en", "ua3", "", time.Now().UTC().Add(-time.Second*20), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp3", "/path", "en", "ua3", "", time.Now().UTC().Add(-time.Second*28), "", "", "", "", false, false)
			analyzer := NewAnalyzer(store)
			visitors, total, err := analyzer.ActiveVisitors(&Filter{
				TenantID: NewTenantID(tenantID),
			}, time.Second*30)

			if err != nil {
				t.Fatalf("Visitors must be returned, but was:  %v", err)
			}

			if total != 3 {
				t.Fatalf("Three active visitors must have been returned, but was: %v", total)
			}

			if len(visitors) != 2 ||
				visitors[0].Path != "/" || visitors[0].Visitors != 2 ||
				visitors[1].Path != "/path" || visitors[1].Visitors != 1 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_ActiveVisitorsPathFilter(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", time.Now().UTC().Add(-time.Second*10), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", time.Now().UTC().Add(-time.Second*11), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp2", "/", "en", "ua2", "", time.Now().UTC().Add(-time.Second*31), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp3", "/", "en", "ua3", "", time.Now().UTC().Add(-time.Second*20), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp3", "/path", "en", "ua3", "", time.Now().UTC().Add(-time.Second*28), "", "", "", "", false, false)
			analyzer := NewAnalyzer(store)
			visitors, total, err := analyzer.ActiveVisitors(&Filter{
				TenantID: NewTenantID(tenantID),
				Path:     "/PAth",
			}, time.Second*30)

			if err != nil {
				t.Fatalf("Visitors must be returned, but was:  %v", err)
			}

			if total != 1 {
				t.Fatalf("One active visitor must have been returned, but was: %v", total)
			}

			if len(visitors) != 1 ||
				visitors[0].Path != "/path" || visitors[0].Visitors != 1 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_Visitors(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp2", "/path", "en", "ua1", "", today(), "", "", "", "", false, false)
			stats := &VisitorStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Path:       "/path",
					Visitors:   42,
				},
			}

			if err := store.SaveVisitorStats(nil, stats); err != nil {
				t.Fatal(err)
			}

			analyzer := NewAnalyzer(store)
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

			if !visitors[0].Day.Equal(pastDay(3)) || visitors[0].Visitors != 0 ||
				!visitors[1].Day.Equal(pastDay(2)) || visitors[1].Visitors != 42 ||
				!visitors[2].Day.Equal(pastDay(1)) || visitors[2].Visitors != 0 ||
				!visitors[3].Day.Equal(today()) || visitors[3].Visitors != 2 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_PageVisitors(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", today(), "", "", "", "", false, false)
			stats := &VisitorStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Path:       "/path",
					Visitors:   42,
				},
			}

			if err := store.SaveVisitorStats(nil, stats); err != nil {
				t.Fatal(err)
			}

			analyzer := NewAnalyzer(store)
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
				!visitors[0].Stats[0].Day.Equal(pastDay(3)) || visitors[0].Stats[0].Visitors != 0 ||
				!visitors[0].Stats[1].Day.Equal(pastDay(2)) || visitors[0].Stats[1].Visitors != 0 ||
				!visitors[0].Stats[2].Day.Equal(pastDay(1)) || visitors[0].Stats[2].Visitors != 0 ||
				!visitors[0].Stats[3].Day.Equal(today()) || visitors[0].Stats[3].Visitors != 1 {
				t.Fatalf("First path not as expected: %v", visitors)
			}

			if len(visitors[1].Stats) != 4 || visitors[1].Path != "/path" ||
				!visitors[1].Stats[0].Day.Equal(pastDay(3)) || visitors[1].Stats[0].Visitors != 0 ||
				!visitors[1].Stats[1].Day.Equal(pastDay(2)) || visitors[1].Stats[1].Visitors != 42 ||
				!visitors[1].Stats[2].Day.Equal(pastDay(1)) || visitors[1].Stats[2].Visitors != 0 ||
				!visitors[1].Stats[3].Day.Equal(today()) || visitors[1].Stats[3].Visitors != 1 {
				t.Fatalf("Second path not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_Languages(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), "", "", "", "", false, false)
			stats := &LanguageStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Path:       "/path",
					Visitors:   42,
				},
				Language: sql.NullString{String: "de", Valid: true},
			}

			if err := store.SaveLanguageStats(nil, stats); err != nil {
				t.Fatal(err)
			}

			analyzer := NewAnalyzer(store)
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

			if visitors[0].Language.String != "de" || visitors[0].Visitors != 43 ||
				visitors[1].Language.String != "en" || visitors[1].Visitors != 1 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_Referrer(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", today(), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "ref2", today(), "", "", "", "", false, false)
			stats := &ReferrerStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Path:       "/path",
					Visitors:   42,
				},
				Referrer: sql.NullString{String: "ref2", Valid: true},
			}

			if err := store.SaveReferrerStats(nil, stats); err != nil {
				t.Fatal(err)
			}

			analyzer := NewAnalyzer(store)
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

			if visitors[0].Referrer.String != "ref2" || visitors[0].Visitors != 43 ||
				visitors[1].Referrer.String != "ref1" || visitors[1].Visitors != 1 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_OS(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), OSWindows, "10", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), OSMac, "10.15.3", "", "", false, false)
			stats := &OSStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Path:       "/path",
					Visitors:   42,
				},
				OS:        sql.NullString{String: OSMac, Valid: true},
				OSVersion: sql.NullString{String: "10.14.1", Valid: true},
			}

			if err := store.SaveOSStats(nil, stats); err != nil {
				t.Fatal(err)
			}

			analyzer := NewAnalyzer(store)
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

			if visitors[0].OS.String != OSMac || visitors[0].Visitors != 43 ||
				visitors[1].OS.String != OSWindows || visitors[1].Visitors != 1 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_Browser(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), "", "", BrowserChrome, "84.0", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), "", "", BrowserFirefox, "54.0", false, false)
			stats := &BrowserStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Path:       "/path",
					Visitors:   42,
				},
				Browser:        sql.NullString{String: BrowserChrome, Valid: true},
				BrowserVersion: sql.NullString{String: "83.1", Valid: true},
			}

			if err := store.SaveBrowserStats(nil, stats); err != nil {
				t.Fatal(err)
			}

			analyzer := NewAnalyzer(store)
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

			if visitors[0].Browser.String != BrowserChrome || visitors[0].Visitors != 43 ||
				visitors[1].Browser.String != BrowserFirefox || visitors[1].Visitors != 1 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
	}
}

func TestAnalyzer_Platform(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", today(), "", "", "", "", true, false)
			createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), "", "", "", "", false, true)
			createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", today(), "", "", "", "", false, false)
			stats := &VisitorStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        pastDay(2),
					Path:       "/path",
				},
				PlatformDesktop: 42,
				PlatformMobile:  43,
				PlatformUnknown: 44,
			}

			if err := store.SaveVisitorStats(nil, stats); err != nil {
				t.Fatal(err)
			}

			analyzer := NewAnalyzer(store)
			visitors := analyzer.Platform(&Filter{
				TenantID: NewTenantID(tenantID),
				From:     pastDay(3),
				To:       today(),
			})

			if visitors.PlatformDesktop != 43 || visitors.PlatformMobile != 44 || visitors.PlatformUnknown != 45 {
				t.Fatalf("Visitors not as expected: %v", visitors)
			}
		}
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
