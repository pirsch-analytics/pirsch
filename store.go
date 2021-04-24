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

	// ActiveVisitors returns the active visitor count for given duration.
	ActiveVisitors(*Filter) int

	// ActiveVisitorsByPath returns the active visitors grouped by path for given duration.
	ActiveVisitorsByPath(*Filter) ([]Stats, error)
}
