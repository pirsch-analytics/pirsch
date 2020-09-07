package pirsch

import (
	"context"
	"sync"
	"time"
)

const (
	defaultMaxAge      = time.Hour
	minCleanupInterval = maxWorkerTimeout * 2
	maxCleanupInterval = maxWorkerTimeout + time.Minute*13
)

type sessionCache struct {
	store      Store
	maxAge     time.Duration
	active     map[string]time.Time // fingerprint -> session timestamp
	inactive   map[string]time.Time
	cancelFunc context.CancelFunc
	m          sync.RWMutex
}

// newSessionCache creates a new sessionCache.
// If maxAge or cleanupInterval are set to 0, the default will be used.
// maxAge is used to define how long a session runs at maximum. The session will be reset once a request is made within that time frame.
// cleanupInterval has a minimum of two times the Tracker worker count and a maximum, to prevent running out of RAM.
func newSessionCache(store Store, maxAge, cleanupInterval time.Duration) *sessionCache {
	if cleanupInterval < minCleanupInterval {
		cleanupInterval = minCleanupInterval
	} else if cleanupInterval > maxCleanupInterval {
		cleanupInterval = maxCleanupInterval
	}

	if maxAge == 0 {
		maxAge = defaultMaxAge
	}

	ctx, cancel := context.WithCancel(context.Background())
	cache := &sessionCache{
		store:      store,
		maxAge:     maxAge,
		active:     make(map[string]time.Time),
		inactive:   make(map[string]time.Time),
		cancelFunc: cancel,
	}
	go cache.cleanup(ctx, cleanupInterval)
	return cache
}

func (cache *sessionCache) stop() {
	cache.cancelFunc()
}

func (cache *sessionCache) cleanup(ctx context.Context, interval time.Duration) {
	for {
		select {
		case <-time.After(interval):
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

func (cache *sessionCache) find(fingerprint string) time.Time {
	// look up the active cache, the non-active cache and the database (in that order)
	// to find an existing session, or add and return a new timestamp if we can't find one
	now := time.Now().UTC()
	cache.m.RLock()
	session := cache.active[fingerprint]

	if session.IsZero() {
		session = cache.inactive[fingerprint]

		if session.IsZero() {
			session = cache.store.Session(fingerprint, now.Add(-cache.maxAge))

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
