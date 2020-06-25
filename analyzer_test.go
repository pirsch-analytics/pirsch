package pirsch

import (
	"testing"
	"time"
)

func TestAnalyzerVisitors(t *testing.T) {
	store := NewPostgresStore(db)
	cleanupDB(t)
	createAnalyzerTestdata(t, store)
	analyzer := NewAnalyzer(store)
	visitors, err := analyzer.Visitors(nil)

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

func TestAnalyzerVisitorsFiltered(t *testing.T) {
	store := NewPostgresStore(db)
	cleanupDB(t)
	createAnalyzerTestdata(t, store)
	analyzer := NewAnalyzer(store)
	visitors, err := analyzer.Visitors(&Filter{pastDay(3), pastDay(2)})

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

func TestAnalyzerPageVisits(t *testing.T) {
	store := NewPostgresStore(db)
	cleanupDB(t)
	createAnalyzerTestdata(t, store)
	analyzer := NewAnalyzer(store)
	visits, err := analyzer.PageVisits(nil)

	if err != nil {
		t.Fatalf("Visits must be returned, but was:  %v", err)
	}

	if len(visits) != 4 {
		t.Fatalf("Must have returns statistics for three pages, but was: %v", len(visits))
	}

	if visits[0].Path != "/" ||
		visits[1].Path != "/bar" ||
		visits[2].Path != "/foo" ||
		visits[3].Path != "/laa" {
		t.Fatal("Paths not as expected")
	}

	if len(visits[0].Visits) != 7 ||
		len(visits[1].Visits) != 7 ||
		len(visits[2].Visits) != 7 ||
		len(visits[3].Visits) != 7 {
		t.Fatal("Page visits not as expected")
	}

	if visits[0].Visits[5].Visitors != 45 ||
		visits[0].Visits[6].Visitors != 1 ||
		visits[1].Visits[4].Visitors != 67 ||
		visits[1].Visits[6].Visitors != 1 ||
		visits[2].Visits[6].Visitors != 1 ||
		visits[3].Visits[5].Visitors != 23 ||
		visits[3].Visits[6].Visitors != 0 {
		t.Fatal("Visitors not as expected")
	}
}

func TestAnalyzerPageVisitsFiltered(t *testing.T) {
	store := NewPostgresStore(db)
	cleanupDB(t)
	createAnalyzerTestdata(t, store)
	analyzer := NewAnalyzer(store)
	visits, err := analyzer.PageVisits(&Filter{pastDay(3), pastDay(2)})

	if err != nil {
		t.Fatalf("Visits must be returned, but was:  %v", err)
	}

	if len(visits) != 1 {
		t.Fatalf("Must have returns statistics for one page, but was: %v", len(visits))
	}

	if visits[0].Path != "/bar" {
		t.Fatal("Path not as expected")
	}

	if len(visits[0].Visits) != 2 {
		t.Fatal("Page visits not as expected")
	}

	if visits[0].Visits[0].Visitors != 0 ||
		visits[0].Visits[1].Visitors != 67 {
		t.Fatal("Visitors not as expected")
	}
}

func TestAnalyzerLanguages(t *testing.T) {
	store := NewPostgresStore(db)
	cleanupDB(t)
	createAnalyzerTestdata(t, store)
	analyzer := NewAnalyzer(store)
	langs, total, err := analyzer.Languages(nil)

	if err != nil {
		t.Fatalf("Languages must be returned, but was:  %v", err)
	}

	if total != 50+14+53 {
		t.Fatalf("Total number of visitors not as expected: %v", total)
	}

	if len(langs) != 3 {
		t.Fatalf("Number of languages not as expected: %v", len(langs))
	}

	if langs[0].Language != "jp" || langs[0].Visitors != 53 || !about(langs[0].RelativeVisitors, 0.45) ||
		langs[1].Language != "en" || langs[1].Visitors != 50 || !about(langs[1].RelativeVisitors, 0.42) ||
		langs[2].Language != "de" || langs[2].Visitors != 14 || !about(langs[2].RelativeVisitors, 0.11) {
		t.Fatalf("Languages not as expected: %v", langs)
	}
}

func TestAnalyzerLanguagesFiltered(t *testing.T) {
	store := NewPostgresStore(db)
	cleanupDB(t)
	createAnalyzerTestdata(t, store)
	analyzer := NewAnalyzer(store)
	langs, total, err := analyzer.Languages(&Filter{pastDay(3), pastDay(2)})

	if err != nil {
		t.Fatalf("Languages must be returned, but was:  %v", err)
	}

	if total != 52+13 {
		t.Fatalf("Total number of visitors not as expected: %v", total)
	}

	if len(langs) != 2 {
		t.Fatalf("Number of languages not as expected: %v", len(langs))
	}

	if langs[0].Language != "jp" || langs[0].Visitors != 52 || !about(langs[0].RelativeVisitors, 0.8) ||
		langs[2].Language != "de" || langs[2].Visitors != 13 || !about(langs[2].RelativeVisitors, 0.2) {
		t.Fatalf("Languages not as expected: %v", langs)
	}
}

func TestAnalyzerValidateFilter(t *testing.T) {
	store := NewPostgresStore(db)
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

func createAnalyzerTestdata(t *testing.T, store Store) {
	createHit(t, store, "fp1", "/", "en", "ua1", pastDay(0))
	createHit(t, store, "fp2", "/foo", "de", "ua2", pastDay(0))
	createHit(t, store, "fp3", "/bar", "jp", "ua3", pastDay(0))
	createVisitorPerDay(t, store, pastDay(1), 42)
	createVisitorPerDay(t, store, pastDay(2), 39)
	createVisitorPerDay(t, store, pastDay(3), 26)
	createVisitorPerPage(t, store, pastDay(1), "/", 45)
	createVisitorPerPage(t, store, pastDay(1), "/laa", 23)
	createVisitorPerPage(t, store, pastDay(2), "/bar", 67)
	createVisitorPerLanguage(t, store, pastDay(1), "en", 49)
	createVisitorPerLanguage(t, store, pastDay(2), "de", 13)
	createVisitorPerLanguage(t, store, pastDay(3), "jp", 52)
}

func createVisitorPerDay(t *testing.T, store Store, day time.Time, visitors int) {
	visitor := VisitorsPerDay{Day: day, Visitors: visitors}

	if err := store.SaveVisitorsPerDay(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerPage(t *testing.T, store Store, day time.Time, path string, visitors int) {
	visitor := VisitorsPerPage{Day: day, Path: path, Visitors: visitors}

	if err := store.SaveVisitorsPerPage(&visitor); err != nil {
		t.Fatal(err)
	}
}

func createVisitorPerLanguage(t *testing.T, store Store, day time.Time, lang string, visitors int) {
	visitor := VisitorsPerLanguage{Day: day, Language: lang, Visitors: visitors}

	if err := store.SaveVisitorsPerLanguage(&visitor); err != nil {
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
