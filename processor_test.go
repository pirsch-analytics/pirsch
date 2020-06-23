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
}

func createTestdata(t *testing.T, store Store) {
	cleanupDB(t)
	createHit(t, store, "fp1", "/", "en", "ua1", day(2020, 6, 23))
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

func day(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}
