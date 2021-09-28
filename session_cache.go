package pirsch

import "time"

// SessionCache caches sessions.
type SessionCache interface {
	// Get returns a session for given client ID, fingerprint, and maximum age.
	Get(uint64, string, time.Time) *Hit

	// Put stores a session for given client ID, fingerprint, and Hit.
	Put(uint64, string, *Hit)

	// Clear clears the cache.
	Clear()
}
