package db

import (
	"database/sql"
	"github.com/pirsch-analytics/pirsch/model"
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SaveHits saves new hits.
	SaveHits([]model.Hit) error

	// Session returns the last session timestamp for given tenant, fingerprint, and maximum age.
	Session(sql.NullInt64, string, time.Time) (time.Time, error)
}
