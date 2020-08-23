package pirsch

import (
	"database/sql"
	"testing"
	"time"
)

func TestProcessor_Process(t *testing.T) {
	for _, store := range testStorageBackends() {
		createTestdata(t, store, 0)
		processor := NewProcessor(store, nil)

		if err := processor.Process(); err != nil {
			t.Fatalf("Data must have been processed, but was: %v", err)
		}

		checkHits(t, store, 0)
		checkVisitorCount(t, store, 0, 3, 3)
		checkVisitorCountHour(t, store, 0, 2, 1, 2, 1)
		checkLanguageCount(t, store, 0, 1, 2, 2, 1)
		checkPageViewCount(t, store, 0, "/", "/page", "/", "/different-page", 2, 1, 2, 1)
		checkReferrerCount(t, store, 0, 3, 2, 1)
		checkOSCount(t, store, 0, 2, 1, 1, 1, 1)
		checkBrowserCount(t, store, 0, 3, 1, 1, 1)
	}
}

func TestProcessor_ProcessTenant(t *testing.T) {
	for _, store := range testStorageBackends() {
		createTestdata(t, store, 1)
		processor := NewProcessor(store, nil)

		if err := processor.ProcessTenant(NewTenantID(1)); err != nil {
			t.Fatalf("Data must have been processed, but was: %v", err)
		}

		checkHits(t, store, 1)
		checkVisitorCount(t, store, 1, 3, 3)
		checkVisitorCountHour(t, store, 1, 2, 1, 2, 1)
		checkLanguageCount(t, store, 1, 1, 2, 2, 1)
		checkPageViewCount(t, store, 1, "/", "/page", "/", "/different-page", 2, 1, 2, 1)
		checkReferrerCount(t, store, 1, 3, 2, 1)
		checkOSCount(t, store, 1, 2, 1, 1, 1, 1)
		checkBrowserCount(t, store, 1, 3, 1, 1, 1)
	}
}

func TestProcessor_ProcessSameDay(t *testing.T) {
	for _, store := range testStorageBackends() {
		createTestdata(t, store, 0)
		createTestDays(t, store)
		processor := NewProcessor(store, nil)

		if err := processor.Process(); err != nil {
			t.Fatalf("Data must have been processed, but was: %v", err)
		}

		checkHits(t, store, 0)
		checkVisitorCount(t, store, 0, 42+3, 3)
		checkVisitorCountHour(t, store, 0, 2, 1, 31+2, 1)
		checkLanguageCount(t, store, 0, 1, 7+2, 2, 1)
		checkPageViewCount(t, store, 0, "/", "/page", "/different-page", "/", 2, 1, 66+1, 2)
		checkReferrerCount(t, store, 0, 13+3, 2, 1)
		checkOSCount(t, store, 0, 19+2, 1, 1, 1, 1)
		checkBrowserCount(t, store, 0, 18+3, 1, 1, 1)
	}
}

func checkHits(t *testing.T, store Store, tenantID int64) {
	if count := store.CountHits(NewTenantID(tenantID)); count != 0 {
		t.Fatalf("Hits must have been cleaned up after processing days, but was: %v", count)
	}
}

func checkVisitorCount(t *testing.T, store Store, tenantID int64, day1, day2 int) {
	visitors := store.VisitorsPerDay(NewTenantID(tenantID))

	if len(visitors) != 2 {
		t.Fatalf("Two visitors per day must have been created, but was: %v", len(visitors))
	}

	if visitors[0].Visitors != day1 || visitors[1].Visitors != day2 {
		t.Fatal("Visitors not as expected")
	}
}

func checkVisitorCountHour(t *testing.T, store Store, tenantID int64, hour1, hour2, hour3, hour4 int) {
	visitors := store.VisitorsPerHour(NewTenantID(tenantID))

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per hour must have been created, but was: %v", len(visitors))
	}

	if visitors[0].DayAndHour.Hour() != 7 ||
		visitors[1].DayAndHour.Hour() != 8 ||
		visitors[2].DayAndHour.Hour() != 9 ||
		visitors[3].DayAndHour.Hour() != 10 {
		t.Fatal("Times not as expected")
	}

	if visitors[0].Visitors != hour1 ||
		visitors[1].Visitors != hour2 ||
		visitors[2].Visitors != hour3 ||
		visitors[3].Visitors != hour4 {
		t.Fatal("Visitors not as expected")
	}
}

func checkLanguageCount(t *testing.T, store Store, tenantID int64, lang1, lang2, lang3, lang4 int) {
	visitors := store.VisitorsPerLanguage(NewTenantID(tenantID))

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per language must have been created, but was: %v", len(visitors))
	}

	if visitors[0].Language != "de" ||
		visitors[1].Language != "en" ||
		visitors[2].Language != "en" ||
		visitors[3].Language != "jp" {
		t.Fatal("Languages not as expected")
	}

	if visitors[0].Visitors != lang1 ||
		visitors[1].Visitors != lang2 ||
		visitors[2].Visitors != lang3 ||
		visitors[3].Visitors != lang4 {
		t.Fatal("Visitors not as expected")
	}
}

func checkPageViewCount(t *testing.T, store Store, tenantID int64, path1, path2, path3, path4 string, views1, views2, views3, views4 int) {
	visitors := store.VisitorsPerPage(NewTenantID(tenantID))

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per page must have been created, but was: %v", len(visitors))
	}

	if visitors[0].Path != path1 ||
		visitors[1].Path != path2 ||
		visitors[2].Path != path3 ||
		visitors[3].Path != path4 {
		t.Fatal("Paths not as expected")
	}

	if visitors[0].Visitors != views1 ||
		visitors[1].Visitors != views2 ||
		visitors[2].Visitors != views3 ||
		visitors[3].Visitors != views4 {
		t.Fatal("Visitors not as expected")
	}
}

func checkReferrerCount(t *testing.T, store Store, tenantID int64, views1, views2, views3 int) {
	visitors := store.VisitorsPerReferrer(NewTenantID(tenantID))

	if len(visitors) != 3 {
		t.Fatalf("Three visitors per referrer must have been created, but was: %v", len(visitors))
	}

	if visitors[0].Ref != "ref1" ||
		visitors[1].Ref != "ref2" ||
		visitors[2].Ref != "ref3" {
		t.Fatal("Referrer not as expected")
	}

	if visitors[0].Visitors != views1 ||
		visitors[1].Visitors != views2 ||
		visitors[2].Visitors != views3 {
		t.Fatal("Visitors not as expected")
	}
}

func checkOSCount(t *testing.T, store Store, tenantID int64, views1, views2, views3, views4, views5 int) {
	visitors := store.VisitorsPerOS(NewTenantID(tenantID))

	if len(visitors) != 5 {
		t.Fatalf("Five visitors per OS must have been created, but was: %v", len(visitors))
	}

	if visitors[0].OS.String != OSWindows || visitors[0].OSVersion.String != "10" ||
		visitors[1].OS.String != OSMac || visitors[1].OSVersion.String != "10.15.3" ||
		visitors[2].OS.String != OSAndroid || visitors[2].OSVersion.String != "8.0" ||
		visitors[3].OS.String != OSLinux || visitors[3].OSVersion.Valid ||
		visitors[4].OS.String != OSWindows || visitors[4].OSVersion.String != "10" {
		t.Fatal("OS not as expected")
	}

	if visitors[0].Visitors != views1 ||
		visitors[1].Visitors != views2 ||
		visitors[2].Visitors != views3 ||
		visitors[3].Visitors != views4 ||
		visitors[4].Visitors != views5 {
		t.Fatal("Visitors not as expected")
	}
}

func checkBrowserCount(t *testing.T, store Store, tenantID int64, views1, views2, views3, views4 int) {
	visitors := store.VisitorsPerBrowser(NewTenantID(tenantID))

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per brower must have been created, but was: %v", len(visitors))
	}

	if visitors[0].Browser.String != BrowserChrome || visitors[0].BrowserVersion.String != "84.0" ||
		visitors[1].Browser.String != BrowserChrome || visitors[1].BrowserVersion.String != "84.0" ||
		visitors[2].Browser.String != BrowserFirefox || visitors[2].BrowserVersion.String != "53.0" ||
		visitors[3].Browser.String != BrowserFirefox || visitors[3].BrowserVersion.String != "54.0" {
		t.Fatal("Browser not as expected")
	}

	if visitors[0].Visitors != views1 ||
		visitors[1].Visitors != views2 ||
		visitors[2].Visitors != views3 ||
		visitors[3].Visitors != views4 {
		t.Fatal("Visitors not as expected")
	}
}

func createTestdata(t *testing.T, store Store, tenantID int64) {
	cleanupDB(t)
	createHit(t, store, tenantID, "fp1", "/", "en", "ua1", "ref1", day(2020, 6, 21, 7), OSWindows, "10", BrowserChrome, "84.0")
	createHit(t, store, tenantID, "fp2", "/", "en", "ua2", "ref1", day(2020, 6, 21, 7), OSWindows, "10", BrowserChrome, "84.0")
	createHit(t, store, tenantID, "fp3", "/page", "de", "ua3", "ref1", day(2020, 6, 21, 8), OSMac, "10.15.3", BrowserChrome, "84.0")
	createHit(t, store, tenantID, "fp4", "/", "en", "ua4", "ref2", day(2020, 6, 22, 9), OSWindows, "10", BrowserFirefox, "53.0")
	createHit(t, store, tenantID, "fp5", "/", "en", "ua5", "ref2", day(2020, 6, 22, 9), OSLinux, "", BrowserFirefox, "54.0")
	createHit(t, store, tenantID, "fp6", "/different-page", "jp", "ua6", "ref3", day(2020, 6, 22, 10), OSAndroid, "8.0", BrowserChrome, "84.0")
}

func createTestDays(t *testing.T, store Store) {
	visitorsPerDay := VisitorsPerDay{
		Day:      day(2020, 6, 21, 5),
		Visitors: 42,
	}

	if err := store.SaveVisitorsPerDay(&visitorsPerDay); err != nil {
		t.Fatal(err)
	}

	visitorsPerHour := VisitorsPerHour{
		DayAndHour: day(2020, 6, 22, 9),
		Visitors:   31,
	}

	if err := store.SaveVisitorsPerHour(&visitorsPerHour); err != nil {
		t.Fatal(err)
	}

	visitorsPerLanguage := VisitorsPerLanguage{
		Day:      day(2020, 6, 21, 5),
		Language: "en",
		Visitors: 7,
	}

	if err := store.SaveVisitorsPerLanguage(&visitorsPerLanguage); err != nil {
		t.Fatal(err)
	}

	visitorsPerPage := VisitorsPerPage{
		Day:      day(2020, 6, 22, 5),
		Path:     "/different-page",
		Visitors: 66,
	}

	if err := store.SaveVisitorsPerPage(&visitorsPerPage); err != nil {
		t.Fatal(err)
	}

	visitorsPerReferrer := VisitorsPerReferrer{
		Day:      day(2020, 6, 21, 7),
		Ref:      "ref1",
		Visitors: 13,
	}

	if err := store.SaveVisitorsPerReferrer(&visitorsPerReferrer); err != nil {
		t.Fatal(err)
	}

	visitorsPerOS := VisitorsPerOS{
		Day:       day(2020, 6, 21, 7),
		OS:        sql.NullString{String: OSWindows, Valid: true},
		OSVersion: sql.NullString{String: "10", Valid: true},
		Visitors:  19,
	}

	if err := store.SaveVisitorsPerOS(&visitorsPerOS); err != nil {
		t.Fatal(err)
	}

	visitorsPerBrowser := VisitorsPerBrowser{
		Day:            day(2020, 6, 21, 7),
		Browser:        sql.NullString{String: BrowserChrome, Valid: true},
		BrowserVersion: sql.NullString{String: "84.0", Valid: true},
		Visitors:       18,
	}

	if err := store.SaveVisitorsPerBrowser(&visitorsPerBrowser); err != nil {
		t.Fatal(err)
	}
}

func createHit(t *testing.T, store Store, tenantID int64, fingerprint, path, lang, userAgent, ref string, time time.Time, os, osVersion, browser, browserVersion string) {
	hit := Hit{
		TenantID:       NewTenantID(tenantID),
		Fingerprint:    fingerprint,
		Path:           sql.NullString{String: path, Valid: path != ""},
		Language:       sql.NullString{String: lang, Valid: path != ""},
		UserAgent:      sql.NullString{String: userAgent, Valid: path != ""},
		Ref:            sql.NullString{String: ref, Valid: path != ""},
		OS:             sql.NullString{String: os, Valid: os != ""},
		OSVersion:      sql.NullString{String: osVersion, Valid: osVersion != ""},
		Browser:        sql.NullString{String: browser, Valid: browser != ""},
		BrowserVersion: sql.NullString{String: browserVersion, Valid: browserVersion != ""},
		Time:           time,
	}

	if err := store.Save([]Hit{hit}); err != nil {
		t.Fatal(err)
	}
}

func day(year, month, day, hour int) time.Time {
	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC)
}
