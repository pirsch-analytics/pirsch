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

// Stats is the base entity for all statistics.
type Stats struct {
	BaseEntity

	Day      time.Time      `db:"day" json:"day"`
	Path     sql.NullString `db:"path" json:"path"`
	Visitors int            `db:"visitors" json:"visitors"`
}

// VisitorStats is the visitor count for each path on each day
// and is used to calculate the total visitor count for each day.
type VisitorStats Stats

// VisitorTimeStats is the visitor count for each path on each day and hour.
type VisitorTimeStats struct {
	Stats

	Time time.Time `db:"time" json:"time"`
}

// LanguageStats is the visitor count for each path on each day and language.
type LanguageStats struct {
	Stats

	Language sql.NullString `db:"language" json:"language"`
}

// ReferrerStats is the visitor count for each path on each day and referrer.
type ReferrerStats struct {
	Stats

	Referrer sql.NullString `db:"referrer" json:"referrer"`
}

// OSStats is the visitor count for each path on each day and operating system.
type OSStats struct {
	Stats

	OS        sql.NullString `db:"os" json:"os"`
	OSVersion sql.NullString `db:"os_version" json:"version"`
}

// BrowserStats is the visitor count for each path on each day and browser.
type BrowserStats struct {
	Stats

	Browser        sql.NullString `db:"browser" json:"browser"`
	BrowserVersion sql.NullString `db:"browser_version" json:"version"`
}

// PlatformStats is the visitor count for each path on each day and platform.
type PlatformStats struct {
	Stats

	Desktop int `db:"desktop" json:"desktop"`
	Mobile  int `db:"mobile" json:"mobile"`
	Unknown int `db:"unknown" json:"unknown"`
}

// Stats bundles all statistics into a single object.
// The meaning of the data depends on the context.
// It's not persisted in the database.
/*type Stats struct {
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
}*/
