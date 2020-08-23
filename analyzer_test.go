package pirsch

import (
	"database/sql"
	"testing"
	"time"
)

func TestAnalyzerVisitors(t *testing.T) {
	testAnalyzerVisitors(t, 0)
	testAnalyzerVisitors(t, 1)
}

func TestAnalyzerVisitorsFiltered(t *testing.T) {
	testAnalyzerVisitorsFiltered(t, 0)
	testAnalyzerVisitorsFiltered(t, 1)
}

func TestAnalyzerPageVisits(t *testing.T) {
	testAnalyzerPageVisits(t, 0)
	testAnalyzerPageVisits(t, 1)
}

func TestAnalyzerReferrerVisits(t *testing.T) {
	testAnalyzerReferrerVisits(t, 0)
	testAnalyzerReferrerVisits(t, 1)
}

func TestAnalyzerPageVisitsFiltered(t *testing.T) {
	testAnalyzerPageVisitsFiltered(t, 0)
	testAnalyzerPageVisitsFiltered(t, 1)
}

func TestAnalyzerPages(t *testing.T) {
	testAnalyzerPages(t, 0)
	testAnalyzerPages(t, 1)
}

func TestAnalyzerPagesFiltered(t *testing.T) {
	testAnalyzerPagesFiltered(t, 0)
	testAnalyzerPagesFiltered(t, 1)
}

func TestAnalyzerLanguages(t *testing.T) {
	testAnalyzerLanguages(t, 0)
	testAnalyzerLanguages(t, 1)
}

func TestAnalyzerLanguagesFiltered(t *testing.T) {
	testAnalyzerLanguagesFiltered(t, 0)
	testAnalyzerLanguagesFiltered(t, 1)
}

func TestAnalyzerReferrer(t *testing.T) {
	testAnalyzerReferrer(t, 0)
	testAnalyzerReferrer(t, 1)
}

func TestAnalyzerReferrerFiltered(t *testing.T) {
	testAnalyzerReferrerFiltered(t, 0)
	testAnalyzerReferrerFiltered(t, 1)
}

func TestAnalyzerOS(t *testing.T) {
	testAnalyzerOS(t, 0)
	testAnalyzerOS(t, 1)
}

func TestAnalyzerOSFiltered(t *testing.T) {
	testAnalyzerOSFiltered(t, 0)
	testAnalyzerOSFiltered(t, 1)
}

func TestAnalyzerBrowser(t *testing.T) {
	testAnalyzerBrowser(t, 0)
	testAnalyzerBrowser(t, 1)
}

func TestAnalyzerBrowserFiltered(t *testing.T) {
	testAnalyzerBrowserFiltered(t, 0)
	testAnalyzerBrowserFiltered(t, 1)
}

func TestAnalyzerHourlyVisitors(t *testing.T) {
	testAnalyzerHourlyVisitors(t, 0)
	testAnalyzerHourlyVisitors(t, 1)
}

func TestAnalyzerActiveVisitors(t *testing.T) {
	testAnalyzerActiveVisitors(t, 0)
	testAnalyzerActiveVisitors(t, 1)
}

func TestAnalyzerActiveVisitorsPages(t *testing.T) {
	testAnalyzerActiveVisitorsPages(t, 0)
	testAnalyzerActiveVisitorsPages(t, 1)
}

func TestAnalyzerHourlyVisitorsFiltered(t *testing.T) {
	testAnalyzerHourlyVisitorsFiltered(t, 0)
	testAnalyzerHourlyVisitorsFiltered(t, 1)
}

func TestAnalyzerValidateFilter(t *testing.T) {
	for _, store := range testStorageBackends() {
		analyzer := NewAnalyzer(store)
		filter := analyzer.validateFilter(nil)

		if filter == nil || !filter.From.Equal(pastDay(6)) || !filter.To.Equal(pastDay(0)) {
			t.Fatalf("Filter not as expected: %v", filter)
		}

		filter = analyzer.validateFilter(&Filter{From: pastDay(2), To: pastDay(5)})

		if filter == nil || !filter.From.Equal(pastDay(5)) || !filter.To.Equal(pastDay(2)) {
			t.Fatalf("Filter not as expected: %v", filter)
		}
	}
}

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
			t.Fatalf("Days not as expected: %v", visitors)
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
			t.Fatalf("Days not as expected: %v", visitors)
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
			t.Fatal("Paths not as expected")
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

		if langs[0].Language.String != "jp" || langs[0].Visitors != 53 || !about(langs[0].RelativeVisitors, 0.45) ||
			langs[1].Language.String != "en" || langs[1].Visitors != 50 || !about(langs[1].RelativeVisitors, 0.42) ||
			langs[2].Language.String != "de" || langs[2].Visitors != 14 || !about(langs[2].RelativeVisitors, 0.11) {
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

		if langs[0].Language.String != "jp" || langs[0].Visitors != 52 || !about(langs[0].RelativeVisitors, 0.8) ||
			langs[1].Language.String != "de" || langs[1].Visitors != 13 || !about(langs[1].RelativeVisitors, 0.2) {
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

		if os[0].OS.String != OSWindows || os[0].Visitors != 125 ||
			os[1].OS.String != OSMac || os[1].Visitors != 20 {
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

		if os[0].OS.String != OSWindows || os[0].Visitors != 72 ||
			os[1].OS.String != OSMac || os[1].Visitors != 19 {
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

		if browser[0].Browser.String != BrowserChrome || browser[0].Visitors != 124 ||
			browser[1].Browser.String != BrowserSafari || browser[1].Visitors != 24 {
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

		if browser[0].Browser.String != BrowserChrome || browser[0].Visitors != 66 ||
			browser[1].Browser.String != BrowserSafari || browser[1].Visitors != 23 {
			t.Fatalf("Browser not as expected: %v", browser)
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
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*5), "", "", "", "")
		createHit(t, store, tenantID, "fp2", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*3), "", "", "", "")
		createHit(t, store, tenantID, "fp3", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*9), "", "", "", "")
		createHit(t, store, tenantID, "fp4", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*11), "", "", "", "")
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
		createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*5), "", "", "", "")
		createHit(t, store, tenantID, "fp2", "/bar", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*3), "", "", "", "")
		createHit(t, store, tenantID, "fp3", "/bar", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*9), "", "", "", "")
		createHit(t, store, tenantID, "fp4", "/", "en", "ua1", "ref", time.Now().UTC().Add(-time.Second*11), "", "", "", "")
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
			t.Fatalf("Paths not as expected, was: %v", visitors)
		}
	}
}

func createAnalyzerTestdata(t *testing.T, store Store, tenantID int64) {
	createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", pastDay(0), OSWindows, "10", BrowserChrome, "84.0")
	createHit(t, store, tenantID, "fp2", "/foo", "De", "ua2", "ref2", pastDay(0), OSWindows, "10", BrowserChrome, "84.0")
	createHit(t, store, tenantID, "fp3", "/bar", "jp", "ua3", "ref3", pastDay(0), OSMac, "10.14.3", BrowserSafari, "13.0")
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
}

func createVisitorPerDay(t *testing.T, store Store, tenantID int64, day time.Time, visitors int) {
	visitor := VisitorsPerDay{
		TenantID: NewTenantID(tenantID),
		Day:      day,
		Visitors: visitors,
	}

	if err := store.SaveVisitorsPerDay(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerPage(t *testing.T, store Store, tenantID int64, day time.Time, path string, visitors int) {
	visitor := VisitorsPerPage{
		TenantID: NewTenantID(tenantID),
		Day:      day,
		Path:     path,
		Visitors: visitors,
	}

	if err := store.SaveVisitorsPerPage(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerReferrer(t *testing.T, store Store, tenantID int64, day time.Time, referrer string, visitors int) {
	visitor := VisitorsPerReferrer{
		TenantID: NewTenantID(tenantID),
		Day:      day,
		Ref:      referrer,
		Visitors: visitors,
	}

	if err := store.SaveVisitorsPerReferrer(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerLanguage(t *testing.T, store Store, tenantID int64, day time.Time, lang string, visitors int) {
	visitor := VisitorsPerLanguage{
		TenantID: NewTenantID(tenantID),
		Day:      day,
		Language: lang,
		Visitors: visitors,
	}

	if err := store.SaveVisitorsPerLanguage(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerHour(t *testing.T, store Store, tenantID int64, dayAndHour time.Time, visitors int) {
	visitor := VisitorsPerHour{
		TenantID:   NewTenantID(tenantID),
		DayAndHour: dayAndHour,
		Visitors:   visitors,
	}

	if err := store.SaveVisitorsPerHour(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerOS(t *testing.T, store Store, tenantID int64, day time.Time, os, osVersion string, visitors int) {
	visitor := VisitorsPerOS{
		TenantID:  NewTenantID(tenantID),
		Day:       day,
		OS:        sql.NullString{String: os, Valid: os != ""},
		OSVersion: sql.NullString{String: osVersion, Valid: osVersion != ""},
		Visitors:  visitors,
	}

	if err := store.SaveVisitorsPerOS(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerBrowser(t *testing.T, store Store, tenantID int64, day time.Time, browser, browserVersion string, visitors int) {
	visitor := VisitorsPerBrowser{
		TenantID:       NewTenantID(tenantID),
		Day:            day,
		Browser:        sql.NullString{String: browser, Valid: browser != ""},
		BrowserVersion: sql.NullString{String: browserVersion, Valid: browserVersion != ""},
		Visitors:       visitors,
	}

	if err := store.SaveVisitorsPerBrowser(&visitor); err != nil {
		t.Fatal(err)
	}
}

func pastDay(n int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}

func equalDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func about(f, target float64) bool {
	return f > target-0.01 && f < target+0.01
}
