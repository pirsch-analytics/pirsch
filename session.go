package pirsch

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

const (
	defaultMaxAge      = time.Minute * 15
	maxMaxAge          = time.Hour * 24
	minCleanupInterval = maxWorkerTimeout * 2
	maxCleanupInterval = minCleanupInterval + time.Hour
)

// sessionCache caches sessions in two maps and swaps them after some time to clean up unused/outdated sessions.
// If a session is not found in cache, it will be looked up in database and added to the cache.
type sessionCache struct {
	store      Store
	maxAge     time.Duration
	active     map[string]time.Time // fingerprint -> session timestamp
	inactive   map[string]time.Time
	cancelFunc context.CancelFunc
	m          sync.RWMutex
}

// sessionCacheConfig is the (optional) configuration for the sessionCache.
type sessionCacheConfig struct {
	maxAge          time.Duration
	cleanupInterval time.Duration
}

func (config *sessionCacheConfig) validate() {
	if config.cleanupInterval < minCleanupInterval {
		config.cleanupInterval = minCleanupInterval
	} else if config.cleanupInterval > maxCleanupInterval {
		config.cleanupInterval = maxCleanupInterval
	}

	if config.maxAge == 0 {
		config.maxAge = defaultMaxAge
	} else if config.maxAge > maxMaxAge {
		config.maxAge = maxMaxAge
	}
}

// newSessionCache creates a new sessionCache.
// If maxAge or cleanupInterval are set to 0, the default will be used.
// maxAge is used to define how long a session runs at maximum. The session will be reset once a request is made within that time frame.
// cleanupInterval has a minimum of two times the Tracker worker count and a maximum, to prevent running out of RAM.
func newSessionCache(store Store, config *sessionCacheConfig) *sessionCache {
	if config == nil {
		config = new(sessionCacheConfig)
	}

	config.validate()
	ctx, cancel := context.WithCancel(context.Background())
	cache := &sessionCache{
		store:      store,
		maxAge:     config.maxAge,
		active:     make(map[string]time.Time),
		inactive:   make(map[string]time.Time),
		cancelFunc: cancel,
	}
	go cache.cleanup(ctx, config.cleanupInterval)
	return cache
}

func (cache *sessionCache) stop() {
	cache.cancelFunc()
}

func (cache *sessionCache) cleanup(ctx context.Context, interval time.Duration) {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		timer.Reset(interval)

		select {
		case <-timer.C:
			cache.swap()
		case <-ctx.Done():
			break
		}
	}
}

func (cache *sessionCache) swap() {
	// swap the active and inactive maps and create a new one for the active map
	cache.m.Lock()
	cache.inactive = cache.active
	cache.active = make(map[string]time.Time)
	cache.m.Unlock()
}

func (cache *sessionCache) find(tenantID sql.NullInt64, fingerprint string) time.Time {
	// look up the active cache, the non-active cache and the database (in that order)
	// to find an existing session, or add and return a new timestamp if we can't find one
	now := time.Now().UTC()
	cache.m.RLock()
	session := cache.active[fingerprint]

	if session.IsZero() {
		session = cache.inactive[fingerprint]

		if session.IsZero() {
			session = cache.store.Session(tenantID, fingerprint, now.Add(-cache.maxAge))

			if session.IsZero() {
				cache.m.RUnlock()
				cache.m.Lock()
				cache.active[fingerprint] = now
				cache.m.Unlock()
				return now
			}
		}
	}

	cache.m.RUnlock()
	return session
}
