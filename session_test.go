package pirsch

import (
	"testing"
)

func TestSessionCache(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	cache := newSessionCache(store, 0, 0)
	defer cache.stop()

	// cache miss -> create in active
	session := cache.find("fp")

	if session.IsZero() {
		t.Fatal("New session must have been returned")
	}

	// find in active
	existing := cache.find("fp")

	if !existing.Equal(session) {
		t.Fatal("Existing session must have been found in active map")
	}

	// find in inactive
	cache.swap()

	if len(cache.active) != 0 || len(cache.inactive) != 1 {
		t.Fatalf("Maps not as expected: %v %v", len(cache.active), len(cache.inactive))
	}

	existing = cache.find("fp")

	if !existing.Equal(session) {
		t.Fatal("Existing session must have been found in inactive map")
	}

	// find in database
	cache.swap()

	if len(cache.active) != 0 || len(cache.inactive) != 0 {
		t.Fatalf("Maps not as expected: %v %v", len(cache.active), len(cache.inactive))
	}

	createHit(t, store, 0, "fp", "/", "en", "ua1", "", today(), session, "", "", "", "", false, false, 0, 0)
	existing = cache.find("fp")

	if existing.IsZero() {
		t.Fatal("Existing session must have been found in database")
	}
}
