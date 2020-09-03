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

	/*
		// Paths returns distinct paths for page visits.
		// This does not include today.
		Paths(sql.NullInt64, time.Time, time.Time) ([]string, error)

		// Referrer returns distinct referrer for page visits.
		// This does not include today.
		Referrer(sql.NullInt64, time.Time, time.Time) ([]string, error)

		// Visitors returns the visitors within given time frame.
		// This does not include today.
		Visitors(sql.NullInt64, time.Time, time.Time) ([]VisitorsPerDay, error)

		// Stats returns the page visits within given time frame for given path.
		// This does not include today.
		PageVisits(sql.NullInt64, string, time.Time, time.Time) ([]VisitorsPerDay, error)

		// ReferrerVisits returns the referrer visits within given time frame for given referrer.
		// This does not include today.
		ReferrerVisits(sql.NullInt64, string, time.Time, time.Time) ([]VisitorsPerReferrer, error)

		// VisitorPages returns the pages and unique visitor count for given time frame.
		// It does include today.
		VisitorPages(sql.NullInt64, time.Time, time.Time) ([]Stats, error)

		// VisitorLanguages returns the language and unique visitor count for given time frame.
		// It does include today.
		VisitorLanguages(sql.NullInt64, time.Time, time.Time) ([]Stats, error)

		// VisitorReferrer returns the referrer and unique visitor count for given time frame.
		// It does include today.
		VisitorReferrer(sql.NullInt64, time.Time, time.Time) ([]Stats, error)

		// VisitorOS returns the OS and unique visitor count for given time frame.
		// It does include today.
		VisitorOS(sql.NullInt64, time.Time, time.Time) ([]Stats, error)

		// VisitorBrowser returns the browser and unique visitor count for given time frame.
		// It does include today.
		VisitorBrowser(sql.NullInt64, time.Time, time.Time) ([]Stats, error)

		// VisitorPlatform returns the platform and unique visitor count for given time frame.
		// It does include today.
		VisitorPlatform(sql.NullInt64, time.Time, time.Time) (*Stats, error)

		// HourlyVisitors returns unique visitors per hour for given time frame.
		// It does include today.
		HourlyVisitors(sql.NullInt64, time.Time, time.Time) ([]Stats, error)

		// ActiveVisitors returns unique visitors starting at given time.
		ActiveVisitors(sql.NullInt64, time.Time) (int, error)

		// ActiveVisitorsPerPage returns unique visitors per page starting at given time.
		ActiveVisitorsPerPage(sql.NullInt64, time.Time) ([]Stats, error)

		// CountHits returns the number of hits for given tenant ID.
		CountHits(sql.NullInt64) int

		// VisitorsPerDay returns all visitors per day for given tenant ID sorted by days.
		VisitorsPerDay(sql.NullInt64) []VisitorsPerDay

		// VisitorsPerHour returns all visitors per hour for given tenant ID sorted by days.
		VisitorsPerHour(sql.NullInt64) []VisitorsPerHour

		// VisitorsPerLanguage returns all visitors per language for given tenant ID in alphabetical order.
		VisitorsPerLanguage(sql.NullInt64) []VisitorsPerLanguage

		// VisitorsPerPage returns all visitors per page for given tenant ID sorted by days.
		VisitorsPerPage(sql.NullInt64) []VisitorsPerPage

		// VisitorsPerReferrer returns all visitors per referrer for given tenant ID sorted by days.
		VisitorsPerReferrer(sql.NullInt64) []VisitorsPerReferrer

		// VisitorsPerOS returns all visitors per operating system for given tenant ID sorted by days.
		VisitorsPerOS(sql.NullInt64) []VisitorsPerOS

		// VisitorsPerBrowser returns all visitors per browsers for given tenant ID sorted by days.
		VisitorsPerBrowser(sql.NullInt64) []VisitorsPerBrowser

		// VisitorsPerPlatform returns all visitor platforms for given tenant ID sorted by days.
		VisitorsPerPlatform(sql.NullInt64) []VisitorPlatform*/
}
