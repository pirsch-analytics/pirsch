package pirsch

import "time"

// VisitorsPerDay is the unique visitor count per day.
type VisitorsPerDay struct {
	ID       int64     `db:"id" json:"id"`
	Day      time.Time `db:"day" json:"day"`
	Visitors int       `db:"visitors" json:"visitors"`
}

// VisitorsPerHour is the unique visitor count per hour and day.
type VisitorsPerHour struct {
	ID         int64     `db:"id" json:"id"`
	DayAndHour time.Time `db:"day_and_hour" json:"day_and_hour"`
	Visitors   int       `db:"visitors" json:"visitors"`
}

// VisitorsPerLanguage is the unique visitor count per language and day.
type VisitorsPerLanguage struct {
	ID       int64     `db:"id" json:"id"`
	Day      time.Time `db:"day" json:"day"`
	Language string    `db:"language" json:"language"`
	Visitors int       `db:"visitors" json:"visitors"`
}

// VisitorsPerPage is the unique visitor count per path and day.
type VisitorsPerPage struct {
	ID       int64     `db:"id" json:"id"`
	Day      time.Time `db:"day" json:"day"`
	Path     string    `db:"path" json:"path"`
	Visitors int       `db:"visitors" json:"visitors"`
}
