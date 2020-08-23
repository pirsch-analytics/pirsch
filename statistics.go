package pirsch

import (
	"database/sql"
	"time"
)

// VisitorsPerDay is the unique visitor count per day.
type VisitorsPerDay struct {
	ID       int64         `db:"id" json:"id"`
	TenantID sql.NullInt64 `db:"tenant_id" json:"tenant_id"`
	Day      time.Time     `db:"day" json:"day"`
	Visitors int           `db:"visitors" json:"visitors"`
}

// VisitorsPerHour is the unique visitor count per hour and day.
type VisitorsPerHour struct {
	ID         int64         `db:"id" json:"id"`
	TenantID   sql.NullInt64 `db:"tenant_id" json:"tenant_id"`
	DayAndHour time.Time     `db:"day_and_hour" json:"day_and_hour"`
	Visitors   int           `db:"visitors" json:"visitors"`
}

// VisitorsPerLanguage is the unique visitor count per language and day.
type VisitorsPerLanguage struct {
	ID       int64         `db:"id" json:"id"`
	TenantID sql.NullInt64 `db:"tenant_id" json:"tenant_id"`
	Day      time.Time     `db:"day" json:"day"`
	Language string        `db:"language" json:"language"`
	Visitors int           `db:"visitors" json:"visitors"`
}

// VisitorsPerPage is the unique visitor count per path and day.
type VisitorsPerPage struct {
	ID       int64         `db:"id" json:"id"`
	TenantID sql.NullInt64 `db:"tenant_id" json:"tenant_id"`
	Day      time.Time     `db:"day" json:"day"`
	Path     string        `db:"path" json:"path"`
	Visitors int           `db:"visitors" json:"visitors"`
}

// VisitorsPerReferrer is the unique visitor count per referrer and day.
type VisitorsPerReferrer struct {
	ID       int64         `db:"id" json:"id"`
	TenantID sql.NullInt64 `db:"tenant_id" json:"tenant_id"`
	Day      time.Time     `db:"day" json:"day"`
	Ref      string        `db:"ref" json:"ref"`
	Visitors int           `db:"visitors" json:"visitors"`
}

// VisitorsPerOS is the unique visitor count per operating system and day.
type VisitorsPerOS struct {
	ID        int64          `db:"id" json:"id"`
	TenantID  sql.NullInt64  `db:"tenant_id" json:"tenant_id"`
	Day       time.Time      `db:"day" json:"day"`
	OS        sql.NullString `db:"os" json:"os"`
	OSVersion sql.NullString `db:"os_version" json:"version"`
	Visitors  int            `db:"visitors" json:"visitors"`
}

// VisitorsPerBrowser is the unique visitor count per browser and day.
type VisitorsPerBrowser struct {
	ID             int64          `db:"id" json:"id"`
	TenantID       sql.NullInt64  `db:"tenant_id" json:"tenant_id"`
	Day            time.Time      `db:"day" json:"day"`
	Browser        sql.NullString `db:"browser" json:"browser"`
	BrowserVersion sql.NullString `db:"browser_version" json:"version"`
	Visitors       int            `db:"visitors" json:"visitors"`
}

// Stats bundles all statistics into a single object.
// The meaning of the data depends on the actual use of this struct.
type Stats struct {
	Path                string                `db:"path" json:"path"`
	Language            string                `db:"language" json:"language"`
	Referrer            string                `db:"ref" json:"referrer"`
	Hour                int                   `db:"hour" json:"hour"`
	Visitors            int                   `db:"visitors" json:"visitors"`
	RelativeVisitors    float64               `db:"-" json:"relative_visitors"`
	VisitorsPerDay      []VisitorsPerDay      `db:"-" json:"visitors_per_day"`
	VisitorsPerReferrer []VisitorsPerReferrer `db:"-" json:"visitors_per_referrer"`
}
