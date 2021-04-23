package hit

import (
	"database/sql"
	"github.com/pirsch-analytics/pirsch/analyze"
	"github.com/pirsch-analytics/pirsch/model"
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
	cleanupDB()
	cache := newSessionCache(dbClient, nil)
	defer cache.stop()

	// cache miss -> create in active
	session := cache.find(analyze.NullTenant, "fp")
	assert.False(t, session.IsZero())

	// find in active
	existing := cache.find(analyze.NullTenant, "fp")
	assert.Equal(t, existing, session)

	// find in inactive
	cache.swap()
	assert.Len(t, cache.active, 0)
	assert.Len(t, cache.inactive, 1)
	existing = cache.find(analyze.NullTenant, "fp")
	assert.Equal(t, existing, session)

	// find in database
	cache.swap()
	assert.Len(t, cache.active, 0)
	assert.Len(t, cache.inactive, 0)
	createHit(t, 0, "fp", "/", "en", "ua1", "", today(), session, "", "", "", "", "", false, false, 0, 0)
	existing = cache.find(analyze.NullTenant, "fp")
	assert.False(t, existing.IsZero())
}

func TestSessionCacheRenewal(t *testing.T) {
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
		cleanupDB()
		createHit(t, 0, "fp", "/", "en", "ua1", "", created, session, "", "", "", "", "", false, false, 0, 0)
		cache := newSessionCache(dbClient, &sessionCacheConfig{
			maxAge: time.Hour,
		})
		s := cache.find(analyze.NullTenant, "fp")

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

func today() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func cleanupDB() {
	dbClient.MustExec(`DELETE FROM "hit"`)
}

func createHit(t *testing.T, tenantID int64, fingerprint, path, lang, userAgent, ref string, time, session time.Time, os, osVersion, browser, browserVersion, countryCode string, desktop, mobile bool, w, h int) {
	screenClass := GetScreenClass(w)
	hit := model.Hit{
		TenantID:       analyze.NewTenantID(tenantID),
		Fingerprint:    fingerprint,
		Time:           time,
		Session:        sql.NullTime{Time: session, Valid: !session.IsZero()},
		UserAgent:      userAgent,
		Path:           path,
		Language:       sql.NullString{String: lang, Valid: lang != ""},
		Referrer:       sql.NullString{String: ref, Valid: ref != ""},
		OS:             sql.NullString{String: os, Valid: os != ""},
		OSVersion:      sql.NullString{String: osVersion, Valid: osVersion != ""},
		Browser:        sql.NullString{String: browser, Valid: browser != ""},
		BrowserVersion: sql.NullString{String: browserVersion, Valid: browserVersion != ""},
		CountryCode:    sql.NullString{String: countryCode, Valid: countryCode != ""},
		Desktop:        desktop,
		Mobile:         mobile,
		ScreenWidth:    w,
		ScreenHeight:   h,
		ScreenClass:    sql.NullString{String: screenClass, Valid: screenClass != ""},
	}

	assert.NoError(t, dbClient.SaveHits([]model.Hit{hit}))
}
