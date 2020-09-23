package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

// NullTenant can be used to pass no (null) tenant to filters and functions.
// This is a sql.NullInt64 with a value of 0.
var NullTenant = NewTenantID(0)

// QueryParams are the default query parameters for (almost) all queries.
type QueryParams struct {
	// TenantID is the (optional) tenant ID.
	TenantID sql.NullInt64

	// Timezone is the time zone used to return results.
	// It's important to keep this consistent across all actions, else it can lead to wrong results or data loss
	// (when deleting hits for a day with a different time zone then intended for example).
	// Note that all time zones are stored as UTC. This will map them to the according time zone.
	Timezone *time.Location
}

// DefaultQueryParams returns a new set of default query parameters.
func DefaultQueryParams() QueryParams {
	return QueryParams{
		TenantID: NullTenant,
		Timezone: time.UTC,
	}
}

func (params *QueryParams) validate() {
	if params.Timezone == nil {
		params.Timezone = time.UTC
	}
}

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
	DeleteHitsByDay(*sqlx.Tx, QueryParams, time.Time) error

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

	// SaveScreenStats saves ScreenStats.
	SaveScreenStats(*sqlx.Tx, *ScreenStats) error

	// SaveCountryStats saves CountryStats.
	SaveCountryStats(*sqlx.Tx, *CountryStats) error

	// Session returns the hits session timestamp for given fingerprint and max age.
	Session(QueryParams, string, time.Time) time.Time

	// HitDays returns the distinct days with at least one hit.
	HitDays(QueryParams) ([]time.Time, error)

	// HitPaths returns the distinct paths for given day.
	HitPaths(QueryParams, time.Time) ([]string, error)

	// Paths returns the distinct paths for given time frame.
	Paths(QueryParams, time.Time, time.Time) ([]string, error)

	// CountVisitors returns the visitor count for given day.
	CountVisitors(*sqlx.Tx, QueryParams, time.Time) *Stats

	// CountVisitorsByPath returns the visitor count for given day, path, and if the platform should be included or not.
	CountVisitorsByPath(*sqlx.Tx, QueryParams, time.Time, string, bool) ([]VisitorStats, error)

	// CountVisitorsByPathAndHour returns the visitor count for given day and path grouped by hour of day.
	CountVisitorsByPathAndHour(*sqlx.Tx, QueryParams, time.Time, string) ([]VisitorTimeStats, error)

	// CountVisitorsByPathAndLanguage returns the visitor count for given day and path grouped by language.
	CountVisitorsByPathAndLanguage(*sqlx.Tx, QueryParams, time.Time, string) ([]LanguageStats, error)

	// CountVisitorsByPathAndReferrer returns the visitor count for given day and path grouped by referrer.
	CountVisitorsByPathAndReferrer(*sqlx.Tx, QueryParams, time.Time, string) ([]ReferrerStats, error)

	// CountVisitorsByPathAndOS returns the visitor count for given day and path grouped by operating system and operating system version.
	CountVisitorsByPathAndOS(*sqlx.Tx, QueryParams, time.Time, string) ([]OSStats, error)

	// CountVisitorsByPathAndBrowser returns the visitor count for given day and path grouped by browser and browser version.
	CountVisitorsByPathAndBrowser(*sqlx.Tx, QueryParams, time.Time, string) ([]BrowserStats, error)

	// CountVisitorsByLanguage returns the visitor count for given day grouped by language.
	CountVisitorsByLanguage(*sqlx.Tx, QueryParams, time.Time) ([]LanguageStats, error)

	// CountVisitorsByReferrer returns the visitor count for given day grouped by referrer.
	CountVisitorsByReferrer(*sqlx.Tx, QueryParams, time.Time) ([]ReferrerStats, error)

	// CountVisitorsByOS returns the visitor count for given day grouped by operating system.
	CountVisitorsByOS(*sqlx.Tx, QueryParams, time.Time) ([]OSStats, error)

	// CountVisitorsByBrowser returns the visitor count for given day grouped by browser.
	CountVisitorsByBrowser(*sqlx.Tx, QueryParams, time.Time) ([]BrowserStats, error)

	// CountVisitorsByScreenSize returns the visitor count for given day grouped by screen size (width and height).
	CountVisitorsByScreenSize(*sqlx.Tx, QueryParams, time.Time) ([]ScreenStats, error)

	// CountVisitorsByCountryCode returns the visitor count for given day grouped by country code.
	CountVisitorsByCountryCode(*sqlx.Tx, QueryParams, time.Time) ([]CountryStats, error)

	// CountVisitorsByPlatform returns the visitor count for given day grouped by platform.
	CountVisitorsByPlatform(*sqlx.Tx, QueryParams, time.Time) *VisitorStats

	// CountVisitorsByPathAndMaxOneHit returns the visitor count for given day and optional path with a maximum of one hit.
	// This returns the absolut number of hits without further page calls and is used to calculate the bounce rate.
	CountVisitorsByPathAndMaxOneHit(*sqlx.Tx, QueryParams, time.Time, string) int

	// ActiveVisitors returns the active visitor count for given duration.
	ActiveVisitors(QueryParams, time.Time) int

	// ActivePageVisitors returns the active visitors grouped by path for given duration.
	ActivePageVisitors(QueryParams, time.Time) ([]Stats, error)

	// Visitors returns the visitors for given time frame grouped by days.
	Visitors(QueryParams, time.Time, time.Time) ([]Stats, error)

	// VisitorHours returns the visitors for given time frame grouped by hour of day.
	VisitorHours(QueryParams, time.Time, time.Time) ([]VisitorTimeStats, error)

	// VisitorLanguages returns the visitors for given time frame grouped by language.
	VisitorLanguages(QueryParams, time.Time, time.Time) ([]LanguageStats, error)

	// VisitorReferrer returns the visitor count for given time frame grouped by referrer.
	VisitorReferrer(QueryParams, time.Time, time.Time) ([]ReferrerStats, error)

	// VisitorOS returns the visitor count for given time frame grouped by operating system.
	VisitorOS(QueryParams, time.Time, time.Time) ([]OSStats, error)

	// VisitorBrowser returns the visitor count for given time frame grouped by browser.
	VisitorBrowser(QueryParams, time.Time, time.Time) ([]BrowserStats, error)

	// VisitorPlatform returns the visitor count for given time frame grouped by platform.
	VisitorPlatform(QueryParams, time.Time, time.Time) *VisitorStats

	// VisitorScreenSize returns the visitor count for given time frame grouped by screen size (width and height).
	VisitorScreenSize(QueryParams, time.Time, time.Time) ([]ScreenStats, error)

	// VisitorCountry returns the visitor count for given time frame grouped by country code.
	VisitorCountry(QueryParams, time.Time, time.Time) ([]CountryStats, error)

	// PageVisitors returns the visitors for given path and time frame grouped by days.
	PageVisitors(QueryParams, string, time.Time, time.Time) ([]Stats, error)

	// PageReferrer returns the visitors for given path and time frame grouped by referrer.
	PageReferrer(QueryParams, string, time.Time, time.Time) ([]ReferrerStats, error)

	// PageLanguages returns the visitors for given path and time frame grouped by language.
	PageLanguages(QueryParams, string, time.Time, time.Time) ([]LanguageStats, error)

	// PageOS returns the visitors for given path and time frame grouped by operating system.
	PageOS(QueryParams, string, time.Time, time.Time) ([]OSStats, error)

	// PageBrowser returns the visitors for given path and time frame grouped by browser.
	PageBrowser(QueryParams, string, time.Time, time.Time) ([]BrowserStats, error)

	// PagePlatform returns the visitors for given path and time frame grouped by platform.
	PagePlatform(QueryParams, string, time.Time, time.Time) *VisitorStats
}
