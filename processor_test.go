package pirsch

import (
	"testing"
	"time"
)

func TestProcessor_Process(t *testing.T) {
	store := NewPostgresStore(db)
	createTestdata(t, store)
	processor := NewProcessor(store)
	processor.Process()
	checkhits(t, store)
	checkVisitorCount(t, store, 3, 3)
	checkVisitorCountHour(t, store, 2, 1, 2, 1)
	checkLanguageCount(t, store, 1, 2, 2, 1)
	checkPageViewCount(t, store, 2, 1, 2, 1)
}

func TestProcessor_ProcessSameDay(t *testing.T) {
	store := NewPostgresStore(db)
	createTestdata(t, store)
	createTestDays(t, store)
	processor := NewProcessor(store)
	processor.Process()
	checkhits(t, store)
	checkVisitorCount(t, store, 42+3, 3)
	checkVisitorCountHour(t, store, 2, 1, 31+2, 1)
	checkLanguageCount(t, store, 1, 7+2, 2, 1)
	checkPageViewCount(t, store, 2, 1, 2, 66+1)
}

func checkhits(t *testing.T, store *PostgresStore) {
	var count int

	if err := store.DB.Get(&count, `SELECT COUNT(1) FROM "hit"`); err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf("Hits must have been cleaned up after processing days, but was: %v", count)
	}
}

func checkVisitorCount(t *testing.T, store *PostgresStore, day1, day2 int) {
	var visitors []VisitorsPerDay

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_day" ORDER BY "day"`); err != nil {
		t.Fatal(err)
	}

	if len(visitors) != 2 {
		t.Fatalf("Two visitors per day must have been created, but was: %v", len(visitors))
	}

	if visitors[0].Visitors != day1 || visitors[1].Visitors != day2 {
		t.Fatal("Visitors not as expected")
	}
}

func checkVisitorCountHour(t *testing.T, store *PostgresStore, hour1, hour2, hour3, hour4 int) {
	var visitors []VisitorsPerHour

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_hour" ORDER BY "day_and_hour"`); err != nil {
		t.Fatal(err)
	}

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

func checkLanguageCount(t *testing.T, store *PostgresStore, lang1, lang2, lang3, lang4 int) {
	var visitors []VisitorsPerLanguage

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_language" ORDER BY "day", "language"`); err != nil {
		t.Fatal(err)
	}

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

func checkPageViewCount(t *testing.T, store *PostgresStore, views1, views2, views3, views4 int) {
	var visitors []VisitorsPerPage

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_page" ORDER BY "day", "path"`); err != nil {
		t.Fatal(err)
	}

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per page must have been created, but was: %v", len(visitors))
	}

	if visitors[0].Path != "/" ||
		visitors[1].Path != "/page" ||
		visitors[2].Path != "/" ||
		visitors[3].Path != "/different-page" {
		t.Fatal("Paths not as expected")
	}

	if visitors[0].Visitors != views1 ||
		visitors[1].Visitors != views2 ||
		visitors[2].Visitors != views3 ||
		visitors[3].Visitors != views4 {
		t.Fatal("Visitors not as expected")
	}
}

func createTestdata(t *testing.T, store Store) {
	cleanupDB(t)
	createHit(t, store, "fp1", "/", "en", "ua1", day(2020, 6, 21, 7))
	createHit(t, store, "fp2", "/", "en", "ua2", day(2020, 6, 21, 7))
	createHit(t, store, "fp3", "/page", "de", "ua3", day(2020, 6, 21, 8))
	createHit(t, store, "fp4", "/", "en", "ua4", day(2020, 6, 22, 9))
	createHit(t, store, "fp5", "/", "en", "ua5", day(2020, 6, 22, 9))
	createHit(t, store, "fp6", "/different-page", "jp", "ua6", day(2020, 6, 22, 10))
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
}

func createHit(t *testing.T, store Store, fingerprint, path, lang, userAgent string, time time.Time) {
	hit := Hit{
		Fingerprint: fingerprint,
		Path:        path,
		Language:    lang,
		UserAgent:   userAgent,
		Time:        time,
	}

	if err := store.Save([]Hit{hit}); err != nil {
		t.Fatal(err)
	}
}

func day(year, month, day, hour int) time.Time {
	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC)
}
