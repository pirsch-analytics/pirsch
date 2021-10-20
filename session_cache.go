package pirsch

import "time"

// SessionCache caches sessions.
type SessionCache interface {
	// Get returns a session for given client ID, fingerprint, and maximum age.
	Get(uint64, uint64, time.Time) *Session

	// Put stores a session for given client ID, fingerprint, and Session.
	Put(uint64, uint64, *Session)

	// Clear clears the cache.
	Clear()
}
