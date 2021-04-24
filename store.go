package pirsch

import (
	"database/sql"
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SaveHits saves new hits.
	SaveHits([]Hit) error

	// Session returns the last session timestamp for given tenant, fingerprint, and maximum age.
	Session(sql.NullInt64, string, time.Time) (time.Time, error)

	// Count returns the number of results for given query.
	Count(*Query) (int, error)

	// Select returns the results for given query.
	Select(*Query) ([]Stats, error)
}
