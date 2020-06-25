package pirsch

import "time"

// Store defines an interface to persists hits and other data.
type Store interface {
	// Save persists a list of hits.
	Save([]Hit) error

	// DeleteHitsByDay deletes all hits on given day.
	DeleteHitsByDay(time.Time) error

	// SaveVisitorsPerDay persists unique visitors per day.
	SaveVisitorsPerDay(*VisitorsPerDay) error

	// SaveVisitorsPerHour persists unique visitors per day and hour.
	SaveVisitorsPerHour(*VisitorsPerHour) error

	// SaveVisitorsPerLanguage persists unique visitors per day and language.
	SaveVisitorsPerLanguage(*VisitorsPerLanguage) error

	// SaveVisitorsPerPage persists unique visitors per day and page.
	SaveVisitorsPerPage(*VisitorsPerPage) error

	// Days returns the days at least one hit exists for.
	Days() ([]time.Time, error)

	// VisitorsPerDay returns the unique visitor count for per day.
	VisitorsPerDay(time.Time) (int, error)

	// VisitorsPerHour returns the unique visitor count per day and hour.
	VisitorsPerDayAndHour(time.Time) ([]VisitorsPerHour, error)

	// VisitorsPerLanguage returns the unique visitor count per language and day.
	VisitorsPerLanguage(time.Time) ([]VisitorsPerLanguage, error)

	// VisitorsPerPage returns the unique visitor count per page and day.
	VisitorsPerPage(time.Time) ([]VisitorsPerPage, error)

	// Paths returns distinct paths for page visits.
	// This does not include today.
	Paths(time.Time, time.Time) ([]string, error)

	// Visitors returns the visitors within given time frame.
	// This does not include today.
	Visitors(time.Time, time.Time) ([]VisitorsPerDay, error)

	// PageVisits returns the page visits within given time frame for given path.
	// This does not include today.
	PageVisits(string, time.Time, time.Time) ([]VisitorsPerDay, error)

	// VisitorLanguages returns the languages within given time frame for unique visitors.
	// It does include today.
	VisitorLanguages(time.Time, time.Time) ([]VisitorLanguage, error)

	// HourlyVisitors returns unique visitors per hour for given time frame.
	// It does include today.
	HourlyVisitors(time.Time, time.Time) ([]HourlyVisitors, error)

	// ActiveVisitors returns unique visitors starting at given time.
	ActiveVisitors(time.Time) (int, error)
}
