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
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", day(2020, 9, 4, 0), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp2", "/path", "en", "ua1", "", day(2020, 9, 4, 0), "", "", "", "", false, false)
			stats := &VisitorStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        day(2020, 9, 2, 0),
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
				From:     day(2020, 9, 1, 0),
				To:       day(2020, 9, 4, 0),
			})

			if err != nil {
				t.Fatalf("Visitors must be returned, but was:  %v", err)
			}

			if len(visitors) != 4 {
				t.Fatalf("Four visitors must have been returned, but was: %v", len(visitors))
			}

			if !visitors[0].Day.Equal(day(2020, 9, 1, 0)) || visitors[0].Visitors != 0 ||
				!visitors[1].Day.Equal(day(2020, 9, 2, 0)) || visitors[1].Visitors != 42 ||
				!visitors[2].Day.Equal(day(2020, 9, 3, 0)) || visitors[2].Visitors != 0 ||
				!visitors[3].Day.Equal(day(2020, 9, 4, 0)) || visitors[3].Visitors != 2 {
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
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", day(2020, 9, 4, 0), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "", day(2020, 9, 4, 0), "", "", "", "", false, false)
			stats := &VisitorStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        day(2020, 9, 2, 0),
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
				From:     day(2020, 9, 1, 0),
				To:       day(2020, 9, 4, 0),
			})

			if err != nil {
				t.Fatalf("Visitors must be returned, but was:  %v", err)
			}

			if len(visitors) != 2 {
				t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
			}

			if len(visitors[0].Stats) != 4 || visitors[0].Path != "/" ||
				!visitors[0].Stats[0].Day.Equal(day(2020, 9, 1, 0)) || visitors[0].Stats[0].Visitors != 0 ||
				!visitors[0].Stats[1].Day.Equal(day(2020, 9, 2, 0)) || visitors[0].Stats[1].Visitors != 0 ||
				!visitors[0].Stats[2].Day.Equal(day(2020, 9, 3, 0)) || visitors[0].Stats[2].Visitors != 0 ||
				!visitors[0].Stats[3].Day.Equal(day(2020, 9, 4, 0)) || visitors[0].Stats[3].Visitors != 1 {
				t.Fatalf("First path not as expected: %v", visitors)
			}

			if len(visitors[1].Stats) != 4 || visitors[1].Path != "/path" ||
				!visitors[1].Stats[0].Day.Equal(day(2020, 9, 1, 0)) || visitors[1].Stats[0].Visitors != 0 ||
				!visitors[1].Stats[1].Day.Equal(day(2020, 9, 2, 0)) || visitors[1].Stats[1].Visitors != 42 ||
				!visitors[1].Stats[2].Day.Equal(day(2020, 9, 3, 0)) || visitors[1].Stats[2].Visitors != 0 ||
				!visitors[1].Stats[3].Day.Equal(day(2020, 9, 4, 0)) || visitors[1].Stats[3].Visitors != 1 {
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
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "", day(2020, 9, 4, 0), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "de", "ua1", "", day(2020, 9, 4, 0), "", "", "", "", false, false)
			stats := &LanguageStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        day(2020, 9, 2, 0),
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
				From:     day(2020, 9, 1, 0),
				To:       day(2020, 9, 4, 0),
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

/*func TestAnalyzer_Referrer(t *testing.T) {
	tenantIDs := []int64{0, 1}

	for _, tenantID := range tenantIDs {
		for _, store := range testStorageBackends() {
			cleanupDB(t)
			createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", day(2020, 9, 4, 0), "", "", "", "", false, false)
			createHit(t, store, tenantID, "fp1", "/path", "en", "ua1", "ref2", day(2020, 9, 4, 0), "", "", "", "", false, false)
			stats := &ReferrerStats{
				Stats: Stats{
					BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
					Day:        day(2020, 9, 2, 0),
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
				From:     day(2020, 9, 1, 0),
				To:       day(2020, 9, 4, 0),
			})

			if err != nil {
				t.Fatalf("Visitors must be returned, but was:  %v", err)
			}

			if len(visitors) != 2 {
				t.Fatalf("Two visitors must have been returned, but was: %v", len(visitors))
			}

			if len(visitors[0].Stats) != 4 || visitors[0].Path != "/" ||
				!visitors[0].Stats[0].Day.Equal(day(2020, 9, 1, 0)) || visitors[0].Stats[0].Visitors != 0 || visitors[0].Stats[0].Referrer.Valid ||
				!visitors[0].Stats[1].Day.Equal(day(2020, 9, 2, 0)) || visitors[0].Stats[1].Visitors != 0 || visitors[0].Stats[1].Referrer.Valid ||
				!visitors[0].Stats[2].Day.Equal(day(2020, 9, 3, 0)) || visitors[0].Stats[2].Visitors != 0 || visitors[0].Stats[2].Referrer.Valid ||
				!visitors[0].Stats[3].Day.Equal(day(2020, 9, 4, 0)) || visitors[0].Stats[3].Visitors != 1 || visitors[0].Stats[3].Referrer.String != "ref1" {
				t.Fatalf("First path not as expected: %v", visitors)
			}

			if len(visitors[1].Stats) != 4 || visitors[1].Path != "/path" ||
				!visitors[1].Stats[0].Day.Equal(day(2020, 9, 1, 0)) || visitors[1].Stats[0].Visitors != 0 || visitors[0].Stats[0].Referrer.Valid ||
				!visitors[1].Stats[1].Day.Equal(day(2020, 9, 2, 0)) || visitors[1].Stats[1].Visitors != 42 || visitors[0].Stats[1].Referrer.String != "ref2" ||
				!visitors[1].Stats[2].Day.Equal(day(2020, 9, 3, 0)) || visitors[1].Stats[2].Visitors != 0 || visitors[1].Stats[2].Referrer.Valid ||
				!visitors[1].Stats[3].Day.Equal(day(2020, 9, 4, 0)) || visitors[1].Stats[3].Visitors != 1 || visitors[0].Stats[3].Referrer.String != "ref2" {
				t.Fatalf("Second path not as expected: %v", visitors)
			}
		}
	}
}*/

/*
func testAnalyzerVisitors(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		visitors, err := analyzer.Visitors(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 7 {
			t.Fatalf("Must have returns statistics for seven days, but was: %v", len(visitors))
		}

		if !equalDay(visitors[0].Day, pastDay(6)) ||
			!equalDay(visitors[1].Day, pastDay(5)) ||
			!equalDay(visitors[2].Day, pastDay(4)) ||
			!equalDay(visitors[3].Day, pastDay(3)) ||
			!equalDay(visitors[4].Day, pastDay(2)) ||
			!equalDay(visitors[5].Day, pastDay(1)) ||
			!equalDay(visitors[6].Day, pastDay(0)) {
			t.Fatalf("HitDays not as expected: %v", visitors)
		}

		if visitors[0].Visitors != 0 ||
			visitors[1].Visitors != 0 ||
			visitors[2].Visitors != 0 ||
			visitors[3].Visitors != 26 ||
			visitors[4].Visitors != 39 ||
			visitors[5].Visitors != 42 ||
			visitors[6].Visitors != 3 {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
	}
}

func testAnalyzerVisitorsFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		visitors, err := analyzer.Visitors(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Must have returns statistics for two days, but was: %v", len(visitors))
		}

		if !equalDay(visitors[0].Day, pastDay(3)) ||
			!equalDay(visitors[1].Day, pastDay(2)) {
			t.Fatalf("HitDays not as expected: %v", visitors)
		}

		if visitors[0].Visitors != 26 ||
			visitors[1].Visitors != 39 {
			t.Fatalf("Visitors not as expected: %v", visitors)
		}
	}
}

func testAnalyzerPageVisits(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		visits, err := analyzer.PageVisits(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Visits must be returned, but was:  %v", err)
		}

		if len(visits) != 4 {
			t.Fatalf("Must have returns statistics for four pages, but was: %v", len(visits))
		}

		if visits[0].Path.String != "/" ||
			visits[1].Path.String != "/bar" ||
			visits[2].Path.String != "/foo" ||
			visits[3].Path.String != "/laa" {
			t.Fatal("HitPaths not as expected")
		}

		if len(visits[0].VisitorsPerDay) != 7 ||
			len(visits[1].VisitorsPerDay) != 7 ||
			len(visits[2].VisitorsPerDay) != 7 ||
			len(visits[3].VisitorsPerDay) != 7 {
			t.Fatal("Page visits not as expected")
		}

		if visits[0].VisitorsPerDay[5].Visitors != 45 ||
			visits[0].VisitorsPerDay[6].Visitors != 1 ||
			visits[1].VisitorsPerDay[4].Visitors != 67 ||
			visits[1].VisitorsPerDay[6].Visitors != 1 ||
			visits[2].VisitorsPerDay[6].Visitors != 1 ||
			visits[3].VisitorsPerDay[5].Visitors != 23 ||
			visits[3].VisitorsPerDay[6].Visitors != 0 {
			t.Fatal("Visitors not as expected")
		}
	}
}

func testAnalyzerReferrerVisits(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		visits, err := analyzer.ReferrerVisits(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Visits must be returned, but was:  %v", err)
		}

		if len(visits) != 3 {
			t.Fatalf("Must have returns statistics for three referrer, but was: %v", len(visits))
		}

		if visits[0].Referrer.String != "ref1" ||
			visits[1].Referrer.String != "ref2" ||
			visits[2].Referrer.String != "ref3" {
			t.Fatal("Referrer not as expected")
		}

		if len(visits[0].VisitorsPerReferrer) != 7 ||
			len(visits[1].VisitorsPerReferrer) != 7 ||
			len(visits[2].VisitorsPerReferrer) != 7 {
			t.Fatal("Referrer visits not as expected")
		}

		if visits[0].VisitorsPerReferrer[5].Visitors != 32 ||
			visits[0].VisitorsPerReferrer[6].Visitors != 1 ||
			visits[1].VisitorsPerReferrer[5].Visitors != 43 ||
			visits[1].VisitorsPerReferrer[6].Visitors != 1 ||
			visits[2].VisitorsPerReferrer[4].Visitors != 54 ||
			visits[2].VisitorsPerReferrer[6].Visitors != 1 {
			t.Fatal("Visitors not as expected")
		}
	}
}

func testAnalyzerPageVisitsFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		visits, err := analyzer.PageVisits(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Visits must be returned, but was:  %v", err)
		}

		if len(visits) != 1 {
			t.Fatalf("Must have returns statistics for one page, but was: %v", len(visits))
		}

		if visits[0].Path.String != "/bar" {
			t.Fatal("Path not as expected")
		}

		if len(visits[0].VisitorsPerDay) != 2 {
			t.Fatal("Page visits not as expected")
		}

		if visits[0].VisitorsPerDay[0].Visitors != 0 ||
			visits[0].VisitorsPerDay[1].Visitors != 67 {
			t.Fatal("Visitors not as expected")
		}
	}
}

func testAnalyzerPages(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		pages, err := analyzer.Pages(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Pages must be returned, but was:  %v", err)
		}

		if len(pages) != 4 {
			t.Fatalf("Number of pages not as expected: %v", len(pages))
		}

		if pages[0].Path.String != "/bar" || pages[0].Visitors != 68 ||
			pages[1].Path.String != "/" || pages[1].Visitors != 46 ||
			pages[2].Path.String != "/laa" || pages[2].Visitors != 23 ||
			pages[3].Path.String != "/foo" || pages[3].Visitors != 1 {
			t.Fatalf("Pages not as expected: %v", pages)
		}
	}
}

func testAnalyzerPagesFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		pages, err := analyzer.Pages(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Pages must be returned, but was:  %v", err)
		}

		if len(pages) != 1 {
			t.Fatalf("Number of pages not as expected: %v", len(pages))
		}

		if pages[0].Path.String != "/bar" || pages[0].Visitors != 67 {
			t.Fatalf("Pages not as expected: %v", pages)
		}
	}
}

func testAnalyzerLanguages(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		langs, total, err := analyzer.Languages(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Languages must be returned, but was:  %v", err)
		}

		if total != 50+14+53 {
			t.Fatalf("Total number of visitors not as expected: %v", total)
		}

		if len(langs) != 3 {
			t.Fatalf("Number of languages not as expected: %v", len(langs))
		}

		if langs[0].Language.String != "jp" || langs[0].Visitors != 53 || !inRange(langs[0].RelativeVisitors, 0.45) ||
			langs[1].Language.String != "en" || langs[1].Visitors != 50 || !inRange(langs[1].RelativeVisitors, 0.42) ||
			langs[2].Language.String != "de" || langs[2].Visitors != 14 || !inRange(langs[2].RelativeVisitors, 0.11) {
			t.Fatalf("Languages not as expected: %v", langs)
		}
	}
}

func testAnalyzerLanguagesFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		langs, total, err := analyzer.Languages(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Languages must be returned, but was:  %v", err)
		}

		if total != 52+13 {
			t.Fatalf("Total number of visitors not as expected: %v", total)
		}

		if len(langs) != 2 {
			t.Fatalf("Number of languages not as expected: %v", len(langs))
		}

		if langs[0].Language.String != "jp" || langs[0].Visitors != 52 || !inRange(langs[0].RelativeVisitors, 0.8) ||
			langs[1].Language.String != "de" || langs[1].Visitors != 13 || !inRange(langs[1].RelativeVisitors, 0.2) {
			t.Fatalf("Languages not as expected: %v", langs)
		}
	}
}

func testAnalyzerReferrer(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		referrer, err := analyzer.Referrer(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Referrer must be returned, but was:  %v", err)
		}

		if len(referrer) != 3 {
			t.Fatalf("Number of referrer not as expected: %v", len(referrer))
		}

		if referrer[0].Referrer.String != "ref3" || referrer[0].Visitors != 55 ||
			referrer[1].Referrer.String != "ref2" || referrer[1].Visitors != 44 ||
			referrer[2].Referrer.String != "ref1" || referrer[2].Visitors != 33 {
			t.Fatalf("Referrer not as expected: %v", referrer)
		}
	}
}

func testAnalyzerReferrerFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		referrer, err := analyzer.Referrer(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Referrer must be returned, but was:  %v", err)
		}

		if len(referrer) != 1 {
			t.Fatalf("Number of referrer not as expected: %v", len(referrer))
		}

		if referrer[0].Referrer.String != "ref3" || referrer[0].Visitors != 54 {
			t.Fatalf("Referrer not as expected: %v", referrer)
		}
	}
}

func testAnalyzerOS(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		os, err := analyzer.OS(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("OS must be returned, but was:  %v", err)
		}

		if len(os) != 2 {
			t.Fatalf("Number of OS not as expected: %v", len(os))
		}

		if os[0].OS.String != OSWindows || os[0].Visitors != 125 || !inRange(os[0].RelativeVisitors, 0.86) ||
			os[1].OS.String != OSMac || os[1].Visitors != 20 || !inRange(os[1].RelativeVisitors, 0.13) {
			t.Fatalf("OS not as expected: %v", os)
		}
	}
}

func testAnalyzerOSFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		os, err := analyzer.OS(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("OS must be returned, but was:  %v", err)
		}

		if len(os) != 2 {
			t.Fatalf("Number of OS not as expected: %v", len(os))
		}

		if os[0].OS.String != OSWindows || os[0].Visitors != 72 || !inRange(os[0].RelativeVisitors, 0.79) ||
			os[1].OS.String != OSMac || os[1].Visitors != 19 || !inRange(os[1].RelativeVisitors, 0.2) {
			t.Fatalf("OS not as expected: %v", os)
		}
	}
}

func testAnalyzerBrowser(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		browser, err := analyzer.Browser(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Browser must be returned, but was:  %v", err)
		}

		if len(browser) != 2 {
			t.Fatalf("Number of browser not as expected: %v", len(browser))
		}

		if browser[0].Browser.String != BrowserChrome || browser[0].Visitors != 124 || !inRange(browser[0].RelativeVisitors, 0.83) ||
			browser[1].Browser.String != BrowserSafari || browser[1].Visitors != 24 || !inRange(browser[1].RelativeVisitors, 0.16) {
			t.Fatalf("Browser not as expected: %v", browser)
		}
	}
}

func testAnalyzerBrowserFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		browser, err := analyzer.Browser(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Browser must be returned, but was:  %v", err)
		}

		if len(browser) != 2 {
			t.Fatalf("Number of browser not as expected: %v", len(browser))
		}

		if browser[0].Browser.String != BrowserChrome || browser[0].Visitors != 66 || !inRange(browser[0].RelativeVisitors, 0.74) ||
			browser[1].Browser.String != BrowserSafari || browser[1].Visitors != 23 || !inRange(browser[1].RelativeVisitors, 0.25) {
			t.Fatalf("Browser not as expected: %v", browser)
		}
	}
}

func testAnalyzerPlatform(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		platform, err := analyzer.Platform(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Platform must be returned, but was:  %v", err)
		}

		if platform.PlatformDesktopVisitors != 33 ||
			platform.PlatformMobileVisitors != 12 ||
			platform.PlatformUnknownVisitors != 98 {
			t.Fatalf("Platforms not as expected: %v", platform)
		}

		if !inRange(platform.PlatformDesktopRelative, 0.23) ||
			!inRange(platform.PlatformMobileRelative, 0.08) ||
			!inRange(platform.PlatformUnknownRelative, 0.68) {
			t.Fatalf("Relative platform usage not as expected: %v", platform)
		}
	}
}

func testAnalyzerPlatformFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		platform, err := analyzer.Platform(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Platform must be returned, but was:  %v", err)
		}

		if platform.PlatformDesktopVisitors != 0 ||
			platform.PlatformMobileVisitors != 11 ||
			platform.PlatformUnknownVisitors != 97 {
			t.Fatalf("Platforms not as expected: %v", platform)
		}

		if !inRange(platform.PlatformDesktopRelative, 0) ||
			!inRange(platform.PlatformMobileRelative, 0.1) ||
			!inRange(platform.PlatformUnknownRelative, 0.89) {
			t.Fatalf("Relative platform usage not as expected: %v", platform)
		}
	}
}

func testAnalyzerHourlyVisitors(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		visitors, err := analyzer.HourlyVisitors(&Filter{
			TenantID: NewTenantID(tenantID),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 24 {
			t.Fatalf("Number of hours must always be 24, but was: %v", len(visitors))
		}

		if visitors[0].Visitors != 3 ||
			visitors[6].Visitors != 32 ||
			visitors[11].Visitors != 8 ||
			visitors[17].Visitors != 29 {
			t.Fatalf("Visitors not as expected. %v", visitors)
		}
	}
}

func testAnalyzerHourlyVisitorsFiltered(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createAnalyzerTestdata(t, store, tenantID)
		analyzer := NewAnalyzer(store)
		visitors, err := analyzer.HourlyVisitors(&Filter{
			TenantID: NewTenantID(tenantID),
			From:     pastDay(3),
			To:       pastDay(2),
		})

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 24 {
			t.Fatalf("Number of hours must always be 24, but was: %v", len(visitors))
		}

		if visitors[0].Visitors != 0 ||
			visitors[6].Visitors != 0 ||
			visitors[11].Visitors != 8 ||
			visitors[17].Visitors != 29 {
			t.Fatalf("Visitors not as expected. %v", visitors)
		}
	}
}

func testAnalyzerActiveVisitors(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*5), "", "", "", "", false, false)
		createHit(t, store, tenantID, "fp2", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*3), "", "", "", "", false, false)
		createHit(t, store, tenantID, "fp3", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*9), "", "", "", "", false, false)
		createHit(t, store, tenantID, "fp4", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*11), "", "", "", "", false, false)
		analyzer := NewAnalyzer(store)
		visitors, err := analyzer.ActiveVisitors(NewTenantID(tenantID), time.Second*10)

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if visitors != 3 {
			t.Fatalf("Visitor count not as expected, was: %v", visitors)
		}
	}
}

func testAnalyzerActiveVisitorsPages(t *testing.T, tenantID int64) {
	for _, store := range testStorageBackends() {
		cleanupDB(t)
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*5), "", "", "", "", false, false)
		createHit(t, store, tenantID, "fp2", "/bar", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*3), "", "", "", "", false, false)
		createHit(t, store, tenantID, "fp3", "/bar", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*9), "", "", "", "", false, false)
		createHit(t, store, tenantID, "fp4", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*11), "", "", "", "", false, false)
		analyzer := NewAnalyzer(store)
		visitors, err := analyzer.ActiveVisitorsPages(NewTenantID(tenantID), time.Second*10)

		if err != nil {
			t.Fatalf("Visitors must be returned, but was:  %v", err)
		}

		if len(visitors) != 2 {
			t.Fatalf("Visitor count not as expected, was: %v", visitors)
		}

		if visitors[0].Path.String != "/bar" || visitors[0].Visitors != 2 ||
			visitors[1].Path.String != "/" || visitors[1].Visitors != 1 {
			t.Fatalf("HitPaths not as expected, was: %v", visitors)
		}
	}
}

func createAnalyzerTestdata(t *testing.T, store Store, tenantID int64) {
	createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", pastDay(0), OSWindows, "10", BrowserChrome, "84.0", true, false)
	createHit(t, store, tenantID, "fp2", "/foo", "De", "ua2", "ref2", pastDay(0), OSWindows, "10", BrowserChrome, "84.0", false, true)
	createHit(t, store, tenantID, "fp3", "/bar", "jp", "ua3", "ref3", pastDay(0), OSMac, "10.14.3", BrowserSafari, "13.0", false, false)
	createVisitorPerDay(t, store, tenantID, pastDay(1), 42)
	createVisitorPerDay(t, store, tenantID, pastDay(2), 39)
	createVisitorPerDay(t, store, tenantID, pastDay(3), 26)
	createVisitorPerPage(t, store, tenantID, pastDay(1), "/", 45)
	createVisitorPerPage(t, store, tenantID, pastDay(1), "/laa", 23)
	createVisitorPerPage(t, store, tenantID, pastDay(2), "/bar", 67)
	createVisitorPerReferrer(t, store, tenantID, pastDay(1), "ref1", 32)
	createVisitorPerReferrer(t, store, tenantID, pastDay(1), "ref2", 43)
	createVisitorPerReferrer(t, store, tenantID, pastDay(2), "ref3", 54)
	createVisitorPerLanguage(t, store, tenantID, pastDay(1), "En", 49)
	createVisitorPerLanguage(t, store, tenantID, pastDay(2), "de", 13)
	createVisitorPerLanguage(t, store, tenantID, pastDay(3), "jP", 52)
	createVisitorPerHour(t, store, tenantID, pastDay(1).Add(time.Hour*6).Add(time.Minute*23), 32)
	createVisitorPerHour(t, store, tenantID, pastDay(2).Add(time.Hour*11).Add(time.Minute*11), 8)
	createVisitorPerHour(t, store, tenantID, pastDay(3).Add(time.Hour*17).Add(time.Minute*59), 29)
	createVisitorPerOS(t, store, tenantID, pastDay(1), OSWindows, "10", 51)
	createVisitorPerOS(t, store, tenantID, pastDay(2), OSWindows, "10", 72)
	createVisitorPerOS(t, store, tenantID, pastDay(3), OSMac, "10.14.3", 19)
	createVisitorPerBrowser(t, store, tenantID, pastDay(1), BrowserChrome, "84.0", 56)
	createVisitorPerBrowser(t, store, tenantID, pastDay(2), BrowserChrome, "84.0", 66)
	createVisitorPerBrowser(t, store, tenantID, pastDay(3), BrowserSafari, "13.0", 23)
	createVisitorPlatform(t, store, tenantID, pastDay(1), 32, 0, 0)
	createVisitorPlatform(t, store, tenantID, pastDay(2), 0, 11, 0)
	createVisitorPlatform(t, store, tenantID, pastDay(3), 0, 0, 97)
}

func createVisitorPerDay(t *testing.T, store Store, tenantID int64, day time.Time, visitors int) {
	visitor := VisitorsPerDay{
		BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
		Day:        day,
		Visitors:   visitors,
	}

	if err := store.SaveVisitorsPerDay(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerPage(t *testing.T, store Store, tenantID int64, day time.Time, path string, visitors int) {
	visitor := VisitorsPerPage{
		BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
		Day:        day,
		Path:       sql.NullString{String: path, Valid: path != ""},
		Visitors:   visitors,
	}

	if err := store.SaveVisitorsPerPage(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerReferrer(t *testing.T, store Store, tenantID int64, day time.Time, referrer string, visitors int) {
	visitor := VisitorsPerReferrer{
		BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
		Day:        day,
		Referrer:        sql.NullString{String: referrer, Valid: referrer != ""},
		Visitors:   visitors,
	}

	if err := store.SaveVisitorsPerReferrer(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerLanguage(t *testing.T, store Store, tenantID int64, day time.Time, lang string, visitors int) {
	visitor := VisitorsPerLanguage{
		BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
		Day:        day,
		Language:   sql.NullString{String: lang, Valid: lang != ""},
		Visitors:   visitors,
	}

	if err := store.SaveVisitorsPerLanguage(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerHour(t *testing.T, store Store, tenantID int64, dayAndHour time.Time, visitors int) {
	visitor := VisitorsPerHour{
		BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
		DayAndHour: dayAndHour,
		Visitors:   visitors,
	}

	if err := store.SaveVisitorsPerHour(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerOS(t *testing.T, store Store, tenantID int64, day time.Time, os, osVersion string, visitors int) {
	visitor := VisitorsPerOS{
		BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
		Day:        day,
		OS:         sql.NullString{String: os, Valid: os != ""},
		OSVersion:  sql.NullString{String: osVersion, Valid: osVersion != ""},
		Visitors:   visitors,
	}

	if err := store.SaveVisitorsPerOS(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerBrowser(t *testing.T, store Store, tenantID int64, day time.Time, browser, browserVersion string, visitors int) {
	visitor := VisitorsPerBrowser{
		BaseEntity:     BaseEntity{TenantID: NewTenantID(tenantID)},
		Day:            day,
		Browser:        sql.NullString{String: browser, Valid: browser != ""},
		BrowserVersion: sql.NullString{String: browserVersion, Valid: browserVersion != ""},
		Visitors:       visitors,
	}

	if err := store.SaveVisitorsPerBrowser(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPlatform(t *testing.T, store Store, tenantID int64, day time.Time, desktop, mobile, unknown int) {
	visitor := VisitorPlatform{
		BaseEntity: BaseEntity{TenantID: NewTenantID(tenantID)},
		Day:        day,
		Desktop:    desktop,
		Mobile:     mobile,
		Unknown:    unknown,
	}

	if err := store.SaveVisitorPlatform(nil, &visitor); err != nil {
		t.Fatal(err)
	}
}
*/

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
