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

	Day              time.Time `db:"day" json:"day"`
	Path             string    `db:"path" json:"path"`
	Visitors         int       `db:"visitors" json:"visitors"`
	Sessions         int       `db:"sessions" json:"sessions"`
	RelativeVisitors float64   `db:"-" json:"relative_visitors"`
}

func (stats *Stats) GetID() int64 {
	return stats.BaseEntity.ID
}

func (stats *Stats) GetVisitors() int {
	return stats.Visitors
}

// VisitorStats is the visitor count for each path on each day and platform
// and it is used to calculate the total visitor count for each day.
type VisitorStats struct {
	Stats

	PlatformDesktop         int     `db:"platform_desktop" json:"platform_desktop"`
	PlatformMobile          int     `db:"platform_mobile" json:"platform_mobile"`
	PlatformUnknown         int     `db:"platform_unknown" json:"platform_unknown"`
	RelativePlatformDesktop float64 `db:"-" json:"relative_platform_desktop"`
	RelativePlatformMobile  float64 `db:"-" json:"relative_platform_mobile"`
	RelativePlatformUnknown float64 `db:"-" json:"relative_platform_unknown"`
}

// VisitorTimeStats is the visitor count for each path on each day and hour (ranging from 0 to 23).
type VisitorTimeStats struct {
	Stats

	Hour int `db:"hour" json:"hour"`
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
