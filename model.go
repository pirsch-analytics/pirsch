package pirsch

import (
	"database/sql"
	"time"
)

// BaseEntity is the base entity for all other entities.
type BaseEntity struct {
	ID       int64         `db:"id" json:"id"`
	TenantID sql.NullInt64 `db:"tenant_id" json:"tenant_id"`
}

// VisitorsPerDay is the unique visitor count per day.
type VisitorsPerDay struct {
	BaseEntity

	Day      time.Time `db:"day" json:"day"`
	Visitors int       `db:"visitors" json:"visitors"`
}

// VisitorsPerHour is the unique visitor count per hour and day.
type VisitorsPerHour struct {
	BaseEntity

	DayAndHour time.Time `db:"day_and_hour" json:"day_and_hour"`
	Visitors   int       `db:"visitors" json:"visitors"`
}

// VisitorsPerLanguage is the unique visitor count per language and day.
type VisitorsPerLanguage struct {
	BaseEntity

	Day      time.Time      `db:"day" json:"day"`
	Language sql.NullString `db:"language" json:"language"`
	Visitors int            `db:"visitors" json:"visitors"`
}

// VisitorsPerPage is the unique visitor count per path and day.
type VisitorsPerPage struct {
	BaseEntity

	Day      time.Time      `db:"day" json:"day"`
	Path     sql.NullString `db:"path" json:"path"`
	Visitors int            `db:"visitors" json:"visitors"`
}

// VisitorsPerReferrer is the unique visitor count per referrer and day.
type VisitorsPerReferrer struct {
	BaseEntity

	Day      time.Time      `db:"day" json:"day"`
	Ref      sql.NullString `db:"ref" json:"ref"`
	Visitors int            `db:"visitors" json:"visitors"`
}

// VisitorsPerOS is the unique visitor count per operating system and day.
type VisitorsPerOS struct {
	BaseEntity

	Day       time.Time      `db:"day" json:"day"`
	OS        sql.NullString `db:"os" json:"os"`
	OSVersion sql.NullString `db:"os_version" json:"version"`
	Visitors  int            `db:"visitors" json:"visitors"`
}

// VisitorsPerBrowser is the unique visitor count per browser and day.
type VisitorsPerBrowser struct {
	BaseEntity

	Day            time.Time      `db:"day" json:"day"`
	Browser        sql.NullString `db:"browser" json:"browser"`
	BrowserVersion sql.NullString `db:"browser_version" json:"version"`
	Visitors       int            `db:"visitors" json:"visitors"`
}

// VisitorPlatform is the visitor count per platform and day.
type VisitorPlatform struct {
	BaseEntity

	Day     time.Time `db:"day" json:"day"`
	Desktop int       `db:"desktop" json:"desktop"`
	Mobile  int       `db:"mobile" json:"mobile"`
	Unknown int       `db:"unknown" json:"unknown"`
}

// Stats bundles all statistics into a single object.
// The meaning of the data depends on the context.
// It's not persisted in the database.
type Stats struct {
	Path                    sql.NullString        `db:"path" json:"path,omitempty"`
	Language                sql.NullString        `db:"language" json:"language,omitempty"`
	Referrer                sql.NullString        `db:"ref" json:"referrer,omitempty"`
	OS                      sql.NullString        `db:"os" json:"os,omitempty"`
	Browser                 sql.NullString        `db:"browser" json:"browser,omitempty"`
	Hour                    int                   `db:"hour" json:"hour,omitempty"`
	Visitors                int                   `db:"visitors" json:"visitors,omitempty"`
	RelativeVisitors        float64               `db:"-" json:"relative_visitors,omitempty"`
	PlatformDesktopVisitors int                   `db:"platform_desktop_visitors" json:"platform_desktop_visitors,omitempty"`
	PlatformDesktopRelative float64               `db:"-" json:"platform_desktop_relative,omitempty"`
	PlatformMobileVisitors  int                   `db:"platform_mobile_visitors" json:"platform_mobile_visitors,omitempty"`
	PlatformMobileRelative  float64               `db:"-" json:"platform_mobile_relative,omitempty"`
	PlatformUnknownVisitors int                   `db:"platform_unknown_visitors" json:"platform_unknown_visitors,omitempty"`
	PlatformUnknownRelative float64               `db:"-" json:"platform_unknown_relative,omitempty"`
	VisitorsPerDay          []VisitorsPerDay      `db:"-" json:"visitors_per_day,omitempty"`
	VisitorsPerReferrer     []VisitorsPerReferrer `db:"-" json:"visitors_per_referrer,omitempty"`
}
