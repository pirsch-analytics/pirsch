package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

// NullTenant can be used to pass no (null) tenant to filters and functions.
// This is a sql.NullInt64 with a value of 0.
var NullTenant = NewTenantID(0)

// Store defines an interface to persists hits and other data.
// The first parameter (if required) is always the tenant ID and can be left out (pirsch.NullTenant), if you don't want to split your data.
// This is usually the case if you integrate Pirsch into your application.
type Store interface {
	// NewTx creates a new transaction and panic on failure.
	NewTx() *sqlx.Tx

	// Commit commits given transaction and logs the error.
	Commit(*sqlx.Tx)

	// Rollback rolls back given transaction and logs the error.
	Rollback(*sqlx.Tx)

	// SaveHits persists a list of hits.
	SaveHits([]Hit) error

	// DeleteHitsByDay deletes all hits on given day.
	DeleteHitsByDay(*sqlx.Tx, sql.NullInt64, time.Time) error

	// Days returns the distinct days with at least one hit.
	Days(sql.NullInt64) ([]time.Time, error)

	// Paths returns the distinct paths for given day.
	Paths(sql.NullInt64, time.Time) ([]string, error)

	// SaveVisitorStats saves VisitorStats.
	SaveVisitorStats(*sqlx.Tx, *VisitorStats) error

	// SaveVisitorTimeStats saves VisitorTimeStats.
	SaveVisitorTimeStats(*sqlx.Tx, *VisitorTimeStats) error

	// SaveLanguageStats saves LanguageStats.
	SaveLanguageStats(*sqlx.Tx, *LanguageStats) error

	// SaveReferrerStats saves ReferrerStats.
	SaveReferrerStats(*sqlx.Tx, *ReferrerStats) error

	// SaveOSStats saves OSStats.
	SaveOSStats(*sqlx.Tx, *OSStats) error

	// SaveBrowserStats saves BrowserStats.
	SaveBrowserStats(*sqlx.Tx, *BrowserStats) error

	// CountVisitorsByPath returns the visitor count for given day and path.
	CountVisitorsByPath(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]VisitorStats, error)

	// CountVisitorsByPathAndHour returns the visitor count for given day and path grouped by hour of day.
	CountVisitorsByPathAndHour(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]VisitorTimeStats, error)

	// CountVisitorsByPathAndLanguage returns the visitor count for given day and path grouped by language.
	CountVisitorsByPathAndLanguage(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]LanguageStats, error)

	// CountVisitorsByPathAndReferrer returns the visitor count for given day and path grouped by referrer.
	CountVisitorsByPathAndReferrer(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]ReferrerStats, error)

	// CountVisitorsByPathAndOS returns the visitor count for given day and path grouped by operating system and operating system version.
	CountVisitorsByPathAndOS(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]OSStats, error)

	// CountVisitorsByPathAndBrowser returns the visitor count for given day and path grouped by browser and browser version.
	CountVisitorsByPathAndBrowser(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]BrowserStats, error)

	// ActiveVisitors returns the active visitors grouped by path for given duration and path.
	// The path is optional and can be left empty to disable path filtering.
	ActiveVisitors(sql.NullInt64, string, time.Time) ([]VisitorStats, error)
}
