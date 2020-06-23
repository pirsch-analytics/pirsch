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
	checkVisitorCount(t, store)
	checkVisitorCountHour(t, store)
	checkLanguageCount(t, store)
	checkPageViewCount(t, store)
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

func checkVisitorCount(t *testing.T, store *PostgresStore) {
	var visitors []VisitorsPerDay

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_day" ORDER BY day`); err != nil {
		t.Fatal(err)
	}

	if len(visitors) != 2 {
		t.Fatalf("Two visitors per day must have been created, but was: %v", len(visitors))
	}
}

func checkVisitorCountHour(t *testing.T, store *PostgresStore) {
	var visitors []VisitorsPerHour

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_hour" ORDER BY day_and_hour`); err != nil {
		t.Fatal(err)
	}

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per hour must have been created, but was: %v", len(visitors))
	}
}

func checkLanguageCount(t *testing.T, store *PostgresStore) {
	var visitors []VisitorsPerLanguage

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_language" ORDER BY day`); err != nil {
		t.Fatal(err)
	}

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per language must have been created, but was: %v", len(visitors))
	}
}

func checkPageViewCount(t *testing.T, store *PostgresStore) {
	var visitors []VisitorsPerPage

	if err := store.DB.Select(&visitors, `SELECT * FROM "visitors_per_page" ORDER BY day`); err != nil {
		t.Fatal(err)
	}

	if len(visitors) != 4 {
		t.Fatalf("Four visitors per page must have been created, but was: %v", len(visitors))
	}
}

func createTestdata(t *testing.T, store Store) {
	cleanupDB(t)
	createHit(t, store, "fp1", "/", "en", "ua1", day(2020, 6, 22, 7))
	createHit(t, store, "fp2", "/", "en", "ua2", day(2020, 6, 22, 7))
	createHit(t, store, "fp3", "/page", "de", "ua3", day(2020, 6, 22, 8))
	createHit(t, store, "fp4", "/", "en", "ua4", day(2020, 6, 23, 9))
	createHit(t, store, "fp5", "/", "en", "ua5", day(2020, 6, 23, 9))
	createHit(t, store, "fp6", "/page", "de", "ua6", day(2020, 6, 23, 10))
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
	return time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local)
}
