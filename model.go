package pirsch

import (
	"encoding/json"
	"time"
)

// Hit represents a single data point/page visit and is the central entity of Pirsch.
type Hit struct {
	ClientID        uint64 `db:"client_id"`
	Fingerprint     string
	Time            time.Time
	SessionID       uint32 `db:"session_id"`
	DurationSeconds uint32 `db:"duration_seconds"`
	UserAgent       string `db:"user_agent"`
	Path            string
	EntryPath       string `db:"entry_path"`
	PageViews       uint16 `db:"page_views"`
	IsBounce        bool   `db:"is_bounce"`
	URL             string
	Title           string
	Language        string
	CountryCode     string `db:"country_code"`
	Referrer        string
	ReferrerName    string `db:"referrer_name"`
	ReferrerIcon    string `db:"referrer_icon"`
	OS              string
	OSVersion       string `db:"os_version"`
	Browser         string
	BrowserVersion  string `db:"browser_version"`
	Desktop         bool
	Mobile          bool
	ScreenWidth     uint16 `db:"screen_width"`
	ScreenHeight    uint16 `db:"screen_height"`
	ScreenClass     string `db:"screen_class"`
	UTMSource       string `db:"utm_source"`
	UTMMedium       string `db:"utm_medium"`
	UTMCampaign     string `db:"utm_campaign"`
	UTMContent      string `db:"utm_content"`
	UTMTerm         string `db:"utm_term"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// Event represents a single data point for custom events.
// It's basically the same as Hit, but with some additional fields (event name, time, and meta fields).
type Event struct {
	Hit
	Name            string   `db:"event_name" json:"name"`
	DurationSeconds int      `db:"event_duration_seconds" json:"duration_seconds"`
	MetaKeys        []string `db:"event_meta_keys" json:"meta_keys"`
	MetaValues      []string `db:"event_meta_values" json:"meta_values"`
}

// String implements the Stringer interface.
func (event Event) String() string {
	out, _ := json.Marshal(event)
	return string(out)
}

// ActiveVisitorStats is the result type for active visitor statistics.
type ActiveVisitorStats struct {
	Path     string `json:"path"`
	Title    string `json:"title"`
	Visitors int    `json:"visitors"`
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
	Path                    string  `json:"path"`
	Title                   string  `json:"title"`
	Visitors                int     `json:"visitors"`
	Views                   int     `json:"views"`
	Sessions                int     `json:"sessions"`
	Bounces                 int     `json:"bounces"`
	RelativeVisitors        float64 `db:"relative_visitors" json:"relative_visitors"`
	RelativeViews           float64 `db:"relative_views" json:"relative_views"`
	BounceRate              float64 `db:"bounce_rate" json:"bounce_rate"`
	AverageTimeSpentSeconds int     `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
}

// EntryStats is the result type for entry page statistics.
type EntryStats struct {
	Path                    string  `json:"path"`
	Title                   string  `json:"title"`
	Visitors                int     `json:"visitors"`
	Entries                 int     `json:"entries"`
	EntryRate               float64 `db:"entry_rate" json:"entry_rate"`
	AverageTimeSpentSeconds int     `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
}

// ExitStats is the result type for exit page statistics.
type ExitStats struct {
	Path     string  `json:"path"`
	Title    string  `json:"title"`
	Visitors int     `json:"visitors"`
	Exits    int     `json:"exits"`
	ExitRate float64 `db:"exit_rate" json:"exit_rate"`
}

// PageConversionsStats is the result type for page conversions.
type PageConversionsStats struct {
	Visitors int     `json:"visitors"`
	Views    int     `json:"views"`
	CR       float64 `json:"cr"`
}

// EventStats is the result type for custom events.
type EventStats struct {
	Name                   string   `db:"event_name" json:"name"`
	Visitors               int      `json:"visitors"`
	Views                  int      `json:"views"`
	CR                     float64  `json:"cr"`
	AverageDurationSeconds int      `db:"average_duration_seconds" json:"average_duration_seconds"`
	MetaKeys               []string `db:"meta_keys" json:"meta_keys"`
	MetaValue              string   `db:"meta_value" json:"meta_value"`
}

// ReferrerStats is the result type for referrer statistics.
type ReferrerStats struct {
	Referrer         string  `json:"referrer"`
	ReferrerName     string  `db:"referrer_name" json:"referrer_name"`
	ReferrerIcon     string  `db:"referrer_icon" json:"referrer_icon"`
	Visitors         int     `json:"visitors"`
	Sessions         int     `json:"sessions"`
	RelativeVisitors float64 `db:"relative_visitors" json:"relative_visitors"`
	Bounces          int     `json:"bounces"`
	BounceRate       float64 `db:"bounce_rate" json:"bounce_rate"`
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
	Day                     time.Time `json:"day"`
	Path                    string    `json:"path"`
	Title                   string    `json:"title"`
	AverageTimeSpentSeconds int       `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
}

// MetaStats is the base for meta result types (languages, countries, ...).
type MetaStats struct {
	Visitors         int     `json:"visitors"`
	RelativeVisitors float64 `db:"relative_visitors" json:"relative_visitors"`
}

// LanguageStats is the result type for language statistics.
type LanguageStats struct {
	MetaStats
	Language string `json:"language"`
}

// CountryStats is the result type for country statistics.
type CountryStats struct {
	MetaStats
	CountryCode string `db:"country_code" json:"country_code"`
}

// BrowserStats is the result type for browser statistics.
type BrowserStats struct {
	MetaStats
	Browser string `json:"browser"`
}

// BrowserVersionStats is the result type for browser version statistics.
type BrowserVersionStats struct {
	MetaStats
	Browser        string `json:"browser"`
	BrowserVersion string `db:"browser_version" json:"browser_version"`
}

// OSStats is the result type for operating system statistics.
type OSStats struct {
	MetaStats
	OS string `json:"os"`
}

// OSVersionStats is the result type for operating system version statistics.
type OSVersionStats struct {
	MetaStats
	OS        string `json:"os"`
	OSVersion string `db:"os_version" json:"os_version"`
}

// ScreenClassStats is the result type for screen class statistics.
type ScreenClassStats struct {
	MetaStats
	ScreenClass string `db:"screen_class" json:"screen_class"`
}

// UTMSourceStats is the result type for utm source statistics.
type UTMSourceStats struct {
	MetaStats
	UTMSource string `db:"utm_source" json:"utm_source"`
}

// UTMMediumStats is the result type for utm medium statistics.
type UTMMediumStats struct {
	MetaStats
	UTMMedium string `db:"utm_medium" json:"utm_medium"`
}

// UTMCampaignStats is the result type for utm campaign statistics.
type UTMCampaignStats struct {
	MetaStats
	UTMCampaign string `db:"utm_campaign" json:"utm_campaign"`
}

// UTMContentStats is the result type for utm content statistics.
type UTMContentStats struct {
	MetaStats
	UTMContent string `db:"utm_content" json:"utm_content"`
}

// UTMTermStats is the result type for utm term statistics.
type UTMTermStats struct {
	MetaStats
	UTMTerm string `db:"utm_term" json:"utm_term"`
}
