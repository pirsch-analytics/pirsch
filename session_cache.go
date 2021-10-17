package pirsch

import "time"

// SessionCache caches sessions.
type SessionCache interface {
	// Get returns a session for given client ID, fingerprint, and maximum age.
	Get(uint64, uint64, time.Time) *Hit

	// Put stores a session for given client ID, fingerprint, and Hit.
	Put(uint64, uint64, *Hit)

	// Clear clears the cache.
	Clear()
}
