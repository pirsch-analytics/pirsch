package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
)

// Cache is a session cache.
type Cache interface {
	// Get returns a session for a given client ID, fingerprint, and offset.
	Get(uint64, uint64, time.Time) *model.Session

	// Put stores a session for given client ID and fingerprint.
	Put(uint64, uint64, *model.Session)

	// Clear clears the cache.
	Clear()

	// NewMutex creates a new mutex for given client ID and fingerprint.
	NewMutex(uint64, uint64) sync.Locker
}

func getSessionKey(clientID, fingerprint uint64) string {
	return fmt.Sprintf("%d_%d", clientID, fingerprint)
}
