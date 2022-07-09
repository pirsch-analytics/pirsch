package session

import (
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"sync"
	"time"
)

const (
	defaultMaxSessions = 10_000
)

// MemCache caches sessions in memory.
// This does only make sense for non-distributed systems (tracking on a single machine/app).
type MemCache struct {
	sessions    map[string]model.Session
	maxSessions int
	client      db.Store
	m           sync.RWMutex
}

// NewMemCache creates a new cache for given client and maximum size.
func NewMemCache(client db.Store, maxSessions int) *MemCache {
	if maxSessions <= 0 {
		maxSessions = defaultMaxSessions
	}

	return &MemCache{
		sessions:    make(map[string]model.Session),
		maxSessions: maxSessions,
		client:      client,
	}
}

// Get implements the Cache interface.
func (cache *MemCache) Get(clientID, fingerprint uint64, maxAge time.Time) *model.Session {
	key := getSessionKey(clientID, fingerprint)
	cache.m.RLock()
	session, found := cache.sessions[key]
	cache.m.RUnlock()

	if found && session.Time.After(maxAge) {
		return &session
	}

	s, _ := cache.client.Session(clientID, fingerprint, maxAge)
	return s
}

// Put implements the Cache interface.
func (cache *MemCache) Put(clientID, fingerprint uint64, session *model.Session) {
	key := getSessionKey(clientID, fingerprint)
	cache.m.Lock()
	defer cache.m.Unlock()

	if len(cache.sessions) >= cache.maxSessions {
		cache.sessions = make(map[string]model.Session)
	}

	existing, found := cache.sessions[key]

	if !found || existing.Time.Equal(session.Time) || existing.Time.Before(session.Time) {
		cache.sessions[key] = *session
	}
}

// Clear implements the Cache interface.
func (cache *MemCache) Clear() {
	cache.m.Lock()
	defer cache.m.Unlock()
	cache.sessions = make(map[string]model.Session)
}

// NewMutex implements the Cache interface.
func (cache *MemCache) NewMutex(uint64, uint64) sync.Locker {
	return new(sync.Mutex)
}

// Sessions returns all sessions.
// This is insecure and should only be used in testing.
func (cache *MemCache) Sessions() map[string]model.Session {
	return cache.sessions
}
