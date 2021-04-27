package pirsch

import (
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SaveHits saves new hits.
	SaveHits([]Hit) error

	// Session returns the last session timestamp for given tenant, fingerprint, and maximum age.
	Session(int64, string, time.Time) (time.Time, error)

	// Count returns the number of results for given query.
	Count(string, ...interface{}) (int, error)

	// Get returns a single result for given query.
	Get(string, ...interface{}) (*Stats, error)

	// Select returns the results for given query.
	Select(string, ...interface{}) ([]Stats, error)
}
