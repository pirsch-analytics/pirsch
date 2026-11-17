package session

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

const (
	defaultMaxSessions = 10_000
	lockShards         = 1024
)

// MemCache caches sessions in memory.
// This does only make sense for non-distributed systems (tracking on a single machine/app).
type MemCache struct {
	sessions    map[string]model.Session
	maxSessions int
	client      db.Storage
	m           sync.RWMutex
	locks       [lockShards]sync.Mutex
}

// NewMemCache creates a new cache for a given client and maximum size.
func NewMemCache(client db.Storage, maxSessions int) *MemCache {
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
func (cache *MemCache) Get(siteID, fingerprint uint64, maxAge time.Time) *model.Session {
	key := getSessionKey(siteID, fingerprint)
	cache.m.RLock()
	session, found := cache.sessions[key]
	cache.m.RUnlock()

	if found && session.Time.After(maxAge) {
		return &session
	}

	s, _ := cache.client.Session(context.Background(), siteID, fingerprint, maxAge)
	return s
}

// Put implements the Cache interface.
func (cache *MemCache) Put(siteID, fingerprint uint64, session *model.Session) {
	key := getSessionKey(siteID, fingerprint)
	cache.m.Lock()
	defer cache.m.Unlock()

	if len(cache.sessions) >= cache.maxSessions {
		cache.sessions = make(map[string]model.Session)
	}

	cache.sessions[key] = *session
}

// Clear implements the Cache interface.
func (cache *MemCache) Clear() {
	cache.m.Lock()
	defer cache.m.Unlock()
	cache.sessions = make(map[string]model.Session)
}

// NewMutex implements the Cache interface.
func (cache *MemCache) NewMutex(siteID, fingerprint uint64) sync.Locker {
	h := siteID*0x9e3779b97f4a7c15 ^ fingerprint
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	return &cache.locks[h%lockShards]
}

// Sessions returns a copy of all sessions.
func (cache *MemCache) Sessions() map[string]model.Session {
	cache.m.RLock()
	defer cache.m.RUnlock()
	sessions := make(map[string]model.Session, len(cache.sessions))
	maps.Copy(sessions, cache.sessions)
	return sessions
}
