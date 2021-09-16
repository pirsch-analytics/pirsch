package pirsch

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultMaxSessions = 20_000
)

// SessionCache caches sessions.
type SessionCache struct {
	sessions    map[string]Session
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
		sessions:    make(map[string]Session),
		maxSessions: maxSessions,
		client:      client,
	}
}

func (cache *SessionCache) get(clientID int64, fingerprint string, maxAge time.Time) *Session {
	key := cache.getKey(clientID, fingerprint)
	cache.m.RLock()
	session, ok := cache.sessions[key]
	cache.m.RUnlock()

	if ok && session.Time.After(maxAge) {
		return &session
	}

	s, _ := cache.client.Session(clientID, fingerprint, maxAge)
	return s
}

func (cache *SessionCache) put(clientID int64, fingerprint, path, entryPath string, pageViews int, now, session time.Time) {
	key := cache.getKey(clientID, fingerprint)
	cache.m.Lock()
	defer cache.m.Unlock()

	if len(cache.sessions) >= cache.maxSessions {
		cache.sessions = make(map[string]Session)
	}

	cache.sessions[key] = Session{
		Time:      now,
		Session:   session,
		Path:      path,
		EntryPath: entryPath,
		PageViews: pageViews,
	}
}

func (cache *SessionCache) getKey(clientID int64, fingerprint string) string {
	return fmt.Sprintf("%d%s", clientID, fingerprint)
}
