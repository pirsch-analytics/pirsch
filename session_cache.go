package pirsch

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultMaxSessions = 100_000
)

type sessionCache struct {
	sessions    map[string]Session
	maxSessions int
	m           sync.RWMutex
}

func newSessionCache(maxSessions int) *sessionCache {
	if maxSessions <= 0 {
		maxSessions = defaultMaxSessions
	}

	return &sessionCache{
		sessions:    make(map[string]Session),
		maxSessions: maxSessions,
	}
}

func (cache *sessionCache) get(client Store, clientID int64, fingerprint string, maxAge time.Time) Session {
	key := cache.getKey(clientID, fingerprint)
	cache.m.RLock()
	session, ok := cache.sessions[key]
	cache.m.RUnlock()

	if ok && session.Time.After(maxAge) {
		return session
	}

	s, _ := client.Session(clientID, fingerprint, maxAge)
	return s
}

func (cache *sessionCache) put(clientID int64, fingerprint, path string, now, session time.Time) {
	key := cache.getKey(clientID, fingerprint)
	cache.m.Lock()
	defer cache.m.Unlock()

	if len(cache.sessions) >= cache.maxSessions {
		cache.sessions = make(map[string]Session)
	}

	cache.sessions[key] = Session{
		Path:    path,
		Time:    now,
		Session: session,
	}
}

func (cache *sessionCache) getKey(clientID int64, fingerprint string) string {
	return fmt.Sprintf("%d%s", clientID, fingerprint)
}
