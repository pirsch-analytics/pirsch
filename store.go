package pirsch

import (
	"database/sql"
	"time"
)

// NullTenant can be used to pass no tenant to filters and functions.
var NullTenant = sql.NullInt64{}

// Store defines an interface to persists hits and other data.
// The first parameter (if required) is always the tenant ID and can be left out (pirsch.NullTenant), if you don't want to split your data.
// This is usually the case if you integrate Pirsch into your application.
type Store interface {
	// Save persists a list of hits.
	Save([]Hit) error

	// DeleteHitsByDay deletes all hits on given day.
	DeleteHitsByDay(sql.NullInt64, time.Time) error

	// SaveVisitorsPerDay persists unique visitors per day.
	SaveVisitorsPerDay(*VisitorsPerDay) error

	// SaveVisitorsPerHour persists unique visitors per day and hour.
	SaveVisitorsPerHour(*VisitorsPerHour) error

	// SaveVisitorsPerLanguage persists unique visitors per day and language.
	SaveVisitorsPerLanguage(*VisitorsPerLanguage) error

	// SaveVisitorsPerPage persists unique visitors per day and page.
	SaveVisitorsPerPage(*VisitorsPerPage) error

	// Days returns the days at least one hit exists for.
	Days(sql.NullInt64) ([]time.Time, error)

	// VisitorsPerDay returns the unique visitor count for per day.
	VisitorsPerDay(sql.NullInt64, time.Time) (int, error)

	// VisitorsPerHour returns the unique visitor count per day and hour.
	VisitorsPerDayAndHour(sql.NullInt64, time.Time) ([]VisitorsPerHour, error)

	// VisitorsPerLanguage returns the unique visitor count per language and day.
	VisitorsPerLanguage(sql.NullInt64, time.Time) ([]VisitorsPerLanguage, error)

	// VisitorsPerPage returns the unique visitor count per page and day.
	VisitorsPerPage(sql.NullInt64, time.Time) ([]VisitorsPerPage, error)

	// Paths returns distinct paths for page visits.
	// This does not include today.
	Paths(sql.NullInt64, time.Time, time.Time) ([]string, error)

	// Visitors returns the visitors within given time frame.
	// This does not include today.
	Visitors(sql.NullInt64, time.Time, time.Time) ([]VisitorsPerDay, error)

	// PageVisits returns the page visits within given time frame for given path.
	// This does not include today.
	PageVisits(sql.NullInt64, string, time.Time, time.Time) ([]VisitorsPerDay, error)

	// VisitorLanguages returns the languages within given time frame for unique visitors.
	// It does include today.
	VisitorLanguages(sql.NullInt64, time.Time, time.Time) ([]VisitorLanguage, error)

	// HourlyVisitors returns unique visitors per hour for given time frame.
	// It does include today.
	HourlyVisitors(sql.NullInt64, time.Time, time.Time) ([]HourlyVisitors, error)

	// ActiveVisitors returns unique visitors starting at given time.
	ActiveVisitors(sql.NullInt64, time.Time) (int, error)
}
