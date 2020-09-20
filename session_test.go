package pirsch

import (
	"testing"
	"time"
)

func TestSessionCacheConfig(t *testing.T) {
	config := sessionCacheConfig{}
	config.validate()

	if config.maxAge != defaultMaxAge ||
		config.cleanupInterval != minCleanupInterval {
		t.Fatalf("Config must have default values, but was: %v", config)
	}

	config.maxAge = time.Hour * 72
	config.cleanupInterval = time.Hour * 4
	config.validate()

	if config.maxAge != maxMaxAge ||
		config.cleanupInterval != maxCleanupInterval {
		t.Fatalf("Config must have max values, but was: %v", config)
	}

	config.maxAge = time.Hour * 4
	config.cleanupInterval = time.Minute * 45
	config.validate()

	if config.maxAge != time.Hour*4 ||
		config.cleanupInterval != time.Minute*45 {
		t.Fatalf("Config must have set values, but was: %v", config)
	}
}

func TestSessionCache(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	cache := newSessionCache(store, nil)
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

	createHit(t, store, 0, "fp", "/", "en", "ua1", "", today(), session, "", "", "", "", "", false, false, 0, 0)
	existing = cache.find("fp")

	if existing.IsZero() {
		t.Fatal("Existing session must have been found in database")
	}
}

func TestSessionCacheRenewal(t *testing.T) {
	store := NewPostgresStore(postgresDB, nil)
	session := time.Now().UTC()
	times := []time.Time{
		time.Now().UTC(),
		time.Now().UTC().Add(-time.Minute * 30),
		time.Now().UTC().Add(-time.Minute * 61),
	}
	found := []bool{
		true,
		true,
		false,
	}

	for i, created := range times {
		cleanupDB(t)
		createHit(t, store, 0, "fp", "/", "en", "ua1", "", created, session, "", "", "", "", "", false, false, 0, 0)
		cache := newSessionCache(store, &sessionCacheConfig{
			maxAge: time.Hour,
		})
		s := cache.find("fp")

		if found[i] && (s.Year() != session.Year() || s.Month() != session.Month() || s.Day() != session.Day() || s.Hour() != session.Hour() || s.Minute() != session.Minute() || s.Second() != session.Second()) {
			t.Fatalf("Session  must have been found, but was: %v", s)
		} else if !found[1] && (s.Year() == session.Year() || s.Month() == session.Month() || s.Day() == session.Day() || s.Hour() == session.Hour() || s.Minute() == session.Minute() || s.Second() == session.Second()) {
			t.Fatalf("Session  must not have been found, but was: %v", s)
		}

		cache.stop()
	}
}
