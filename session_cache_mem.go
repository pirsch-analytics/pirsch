package pirsch

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultMaxSessions = 10_000
)

// SessionCacheMem caches sessions in memory.
// This does only make sense for non-distributed systems (tracking on a single machine/app).
type SessionCacheMem struct {
	sessions    map[string]Session
	maxSessions int
	client      Store
	m           sync.RWMutex
}

// NewSessionCacheMem creates a new cache for given client and maximum size.
func NewSessionCacheMem(client Store, maxSessions int) *SessionCacheMem {
	if maxSessions <= 0 {
		maxSessions = defaultMaxSessions
	}

	return &SessionCacheMem{
		sessions:    make(map[string]Session),
		maxSessions: maxSessions,
		client:      client,
	}
}

// Get implements the SessionCache interface.
func (cache *SessionCacheMem) Get(clientID, fingerprint uint64, maxAge time.Time) *Session {
	key := getSessionKey(clientID, fingerprint)
	cache.m.RLock()
	hit, ok := cache.sessions[key]
	cache.m.RUnlock()

	if ok && hit.Time.After(maxAge) {
		return &hit
	}

	s, _ := cache.client.Session(clientID, fingerprint, maxAge)
	return s
}

// Put implements the SessionCache interface.
func (cache *SessionCacheMem) Put(clientID, fingerprint uint64, hit *Session) {
	key := getSessionKey(clientID, fingerprint)
	cache.m.Lock()
	defer cache.m.Unlock()

	if len(cache.sessions) >= cache.maxSessions {
		cache.sessions = make(map[string]Session)
	}

	cache.sessions[key] = *hit
}

// Clear implements the SessionCache interface.
func (cache *SessionCacheMem) Clear() {
	cache.m.Lock()
	defer cache.m.Unlock()
	cache.sessions = make(map[string]Session)
}

func getSessionKey(clientID, fingerprint uint64) string {
	return fmt.Sprintf("%d_%d", clientID, fingerprint)
}
