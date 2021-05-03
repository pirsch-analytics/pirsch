package pirsch

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Hit represents a single data point/page visit and is the central entity of Pirsch.
type Hit struct {
	ClientID                  int64 `db:"client_id"`
	Fingerprint               string
	Time                      time.Time
	Session                   sql.NullTime
	PreviousTimeOnPageSeconds int    `db:"previous_time_on_page_seconds"`
	UserAgent                 string `db:"user_agent"`
	Path                      string
	URL                       string
	Language                  string
	CountryCode               string `db:"country_code"`
	Referrer                  sql.NullString
	ReferrerName              sql.NullString `db:"referrer_name"`
	ReferrerIcon              sql.NullString `db:"referrer_icon"`
	OS                        string
	OSVersion                 string `db:"os_version"`
	Browser                   string
	BrowserVersion            string `db:"browser_version"`
	Desktop                   bool
	Mobile                    bool
	ScreenWidth               int            `db:"screen_width"`
	ScreenHeight              int            `db:"screen_height"`
	ScreenClass               string         `db:"screen_class"`
	UTMSource                 sql.NullString `db:"utm_source"`
	UTMMedium                 sql.NullString `db:"utm_medium"`
	UTMCampaign               sql.NullString `db:"utm_campaign"`
	UTMContent                sql.NullString `db:"utm_content"`
	UTMTerm                   sql.NullString `db:"utm_term"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// ActiveVisitorStats is the result type for active visitor statistics.
type ActiveVisitorStats struct {
	Path     sql.NullString `json:"path"`
	Visitors int            `json:"visitors"`
}

// VisitorStats is the result type for visitor statistics.
type VisitorStats struct {
	Day        time.Time `json:"day"`
	Visitors   int       `json:"visitors"`
	Views      int       `json:"views"`
	Sessions   int       `json:"sessions"`
	Bounces    int       `json:"bounces"`
	BounceRate float64   `db:"bounce_rate" json:"bounce_rate"`
}

// Growth represents the visitors, views, sessions, bounces, and average session duration growth between two time periods.
type Growth struct {
	VisitorsGrowth  float64 `json:"visitors_growth"`
	ViewsGrowth     float64 `json:"views_growth"`
	SessionsGrowth  float64 `json:"sessions_growth"`
	BouncesGrowth   float64 `json:"bounces_growth"`
	TimeSpentGrowth float64 `json:"time_spent_growth"`
}

// VisitorHourStats is the result type for visitor statistics grouped by time of day.
type VisitorHourStats struct {
	Hour     int `json:"hour"`
	Visitors int `json:"visitors"`
}

// PageStats is the result type for page statistics.
type PageStats struct {
	Path                    sql.NullString `json:"path"`
	Visitors                int            `json:"visitors"`
	Views                   int            `json:"views"`
	Sessions                int            `json:"sessions"`
	Bounces                 int            `json:"bounces"`
	RelativeVisitors        float64        `db:"relative_visitors" json:"relative_visitors"`
	RelativeViews           float64        `db:"relative_views" json:"relative_views"`
	BounceRate              float64        `db:"bounce_rate" json:"bounce_rate"`
	AverageTimeSpentSeconds int            `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
}

// ReferrerStats is the result type for referrer statistics.
type ReferrerStats struct {
	Referrer         sql.NullString `json:"referrer"`
	ReferrerName     sql.NullString `db:"referrer_name" json:"referrer_name"`
	ReferrerIcon     sql.NullString `db:"referrer_icon" json:"referrer_icon"`
	Visitors         int            `json:"visitors"`
	RelativeVisitors float64        `db:"relative_visitors" json:"relative_visitors"`
	Bounces          int            `json:"bounces"`
	BounceRate       float64        `db:"bounce_rate" json:"bounce_rate"`
}

// PlatformStats is the result type for platform statistics.
type PlatformStats struct {
	PlatformDesktop         int     `db:"platform_desktop" json:"platform_desktop"`
	PlatformMobile          int     `db:"platform_mobile" json:"platform_mobile"`
	PlatformUnknown         int     `db:"platform_unknown" json:"platform_unknown"`
	RelativePlatformDesktop float64 `db:"relative_platform_desktop" json:"relative_platform_desktop"`
	RelativePlatformMobile  float64 `db:"relative_platform_mobile" json:"relative_platform_mobile"`
	RelativePlatformUnknown float64 `db:"relative_platform_unknown" json:"relative_platform_unknown"`
}

// TimeSpentStats is the result type for average time spent statistics (sessions, time on page).
type TimeSpentStats struct {
	Day                     time.Time      `json:"day"`
	Path                    sql.NullString `json:"path"`
	AverageTimeSpentSeconds int            `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
}

// MetaStats is the base for meta result types (languages, countries, ...).
type MetaStats struct {
	Visitors         int     `json:"visitors"`
	RelativeVisitors float64 `db:"relative_visitors" json:"relative_visitors"`
}

// LanguageStats is the result type for language statistics.
type LanguageStats struct {
	MetaStats
	Language sql.NullString `json:"language"`
}

// CountryStats is the result type for country statistics.
type CountryStats struct {
	MetaStats
	CountryCode sql.NullString `db:"country_code" json:"country_code"`
}

// BrowserStats is the result type for browser statistics.
type BrowserStats struct {
	MetaStats
	Browser sql.NullString `json:"browser"`
}

// OSStats is the result type for operating system statistics.
type OSStats struct {
	MetaStats
	OS sql.NullString `json:"os"`
}

// ScreenClassStats is the result type for screen class statistics.
type ScreenClassStats struct {
	MetaStats
	ScreenClass sql.NullString `db:"screen_class" json:"screen_class"`
}

// UTMSourceStats is the result type for utm source statistics.
type UTMSourceStats struct {
	MetaStats
	UTMSource sql.NullString `db:"utm_source" json:"utm_source"`
}

// UTMMediumStats is the result type for utm medium statistics.
type UTMMediumStats struct {
	MetaStats
	UTMMedium sql.NullString `db:"utm_medium" json:"utm_medium"`
}

// UTMCampaignStats is the result type for utm campaign statistics.
type UTMCampaignStats struct {
	MetaStats
	UTMCampaign sql.NullString `db:"utm_campaign" json:"utm_campaign"`
}

// UTMContentStats is the result type for utm content statistics.
type UTMContentStats struct {
	MetaStats
	UTMContent sql.NullString `db:"utm_content" json:"utm_content"`
}

// UTMTermStats is the result type for utm term statistics.
type UTMTermStats struct {
	MetaStats
	UTMTerm sql.NullString `db:"utm_term" json:"utm_term"`
}
