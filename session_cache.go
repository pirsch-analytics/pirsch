package pirsch

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultMaxSessions = 10_000
)

// SessionCache caches sessions.
type SessionCache struct {
	sessions    map[string]Hit
	maxSessions int
	client      Store
	m           sync.RWMutex
}

// NewSessionCache creates a new cache for given client and maximum size.
func NewSessionCache(client Store, maxSessions int) *SessionCache {
	if maxSessions <= 0 {
		maxSessions = defaultMaxSessions
	}

	return &SessionCache{
		sessions:    make(map[string]Hit),
		maxSessions: maxSessions,
		client:      client,
	}
}

func (cache *SessionCache) get(clientID int64, fingerprint string, maxAge time.Time) *Hit {
	key := cache.getKey(clientID, fingerprint)
	cache.m.RLock()
	hit, ok := cache.sessions[key]
	cache.m.RUnlock()

	if ok && hit.Time.After(maxAge) {
		return &hit
	}

	s, _ := cache.client.Session(clientID, fingerprint, maxAge)
	return s
}

func (cache *SessionCache) put(clientID int64, fingerprint string, hit *Hit) {
	key := cache.getKey(clientID, fingerprint)
	cache.m.Lock()
	defer cache.m.Unlock()

	if len(cache.sessions) >= cache.maxSessions {
		cache.sessions = make(map[string]Hit)
	}

	cache.sessions[key] = *hit
}

func (cache *SessionCache) getKey(clientID int64, fingerprint string) string {
	return fmt.Sprintf("%d%s", clientID, fingerprint)
}
