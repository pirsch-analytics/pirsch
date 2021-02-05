package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSessionCacheConfig(t *testing.T) {
	config := sessionCacheConfig{}
	config.validate()
	assert.Equal(t, defaultMaxAge, config.maxAge)
	assert.Equal(t, minCleanupInterval, config.cleanupInterval)
	config.maxAge = time.Hour * 72
	config.cleanupInterval = time.Hour * 4
	config.validate()
	assert.Equal(t, maxMaxAge, config.maxAge)
	assert.Equal(t, maxCleanupInterval, config.cleanupInterval)
	config.maxAge = time.Hour * 4
	config.cleanupInterval = time.Minute * 45
	config.validate()
	assert.Equal(t, time.Hour*4, config.maxAge)
	assert.Equal(t, time.Minute*45, config.cleanupInterval)
}

func TestSessionCache(t *testing.T) {
	cleanupDB(t)
	store := NewPostgresStore(postgresDB, nil)
	cache := newSessionCache(store, nil)
	defer cache.stop()

	// cache miss -> create in active
	session := cache.find(NullTenant, "fp")
	assert.False(t, session.IsZero())

	// find in active
	existing := cache.find(NullTenant, "fp")
	assert.Equal(t, existing, session)

	// find in inactive
	cache.swap()
	assert.Len(t, cache.active, 0)
	assert.Len(t, cache.inactive, 1)
	existing = cache.find(NullTenant, "fp")
	assert.Equal(t, existing, session)

	// find in database
	cache.swap()
	assert.Len(t, cache.active, 0)
	assert.Len(t, cache.inactive, 0)
	createHit(t, store, 0, "fp", "/", "en", "ua1", "", today(), session, "", "", "", "", "", false, false, 0, 0)
	existing = cache.find(NullTenant, "fp")
	assert.False(t, existing.IsZero())
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
		s := cache.find(NullTenant, "fp")

		if found[i] {
			assert.Equal(t, s.Year(), session.Year())
			assert.Equal(t, s.Month(), session.Month())
			assert.Equal(t, s.Day(), session.Day())
			assert.Equal(t, s.Hour(), session.Hour())
			assert.Equal(t, s.Minute(), session.Minute())
			assert.Equal(t, s.Second(), session.Second())
		} else if !found[1] {
			assert.NotEqual(t, s.Year(), session.Year())
			assert.NotEqual(t, s.Month(), session.Month())
			assert.NotEqual(t, s.Day(), session.Day())
			assert.NotEqual(t, s.Hour(), session.Hour())
			assert.NotEqual(t, s.Minute(), session.Minute())
			assert.NotEqual(t, s.Second(), session.Second())
		}

		cache.stop()
	}
}
