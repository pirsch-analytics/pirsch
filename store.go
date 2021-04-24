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

	// CountActiveVisitors returns the active visitor count for given duration.
	//CountActiveVisitors(*Run) int

	// ActiveVisitors returns the active visitors grouped by path for given duration.
	//ActiveVisitors(*Run) ([]Stats, error)

	// VisitorLanguages returns the visitors grouped by language.
	//VisitorLanguages(*Run) ([]Stats, error)

	// Run returns the results for given query.
	Run(*Query) ([]Stats, error)
}
