package pirsch

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Hit represents a single data point/page visit and is the central entity of Pirsch.
type Hit struct {
	ClientID       int64          `db:"client_id"`
	Fingerprint    string         `db:"fingerprint"`
	Time           time.Time      `db:"time"`
	Session        sql.NullTime   `db:"session"`
	UserAgent      string         `db:"user_agent"`
	Path           string         `db:"path"`
	URL            string         `db:"url"`
	Language       string         `db:"language"`
	CountryCode    string         `db:"country_code"`
	Referrer       sql.NullString `db:"referrer"`
	ReferrerName   sql.NullString `db:"referrer_name"`
	ReferrerIcon   sql.NullString `db:"referrer_icon"`
	OS             string         `db:"os"`
	OSVersion      string         `db:"os_version"`
	Browser        string         `db:"browser"`
	BrowserVersion string         `db:"browser_version"`
	Desktop        bool           `db:"desktop"`
	Mobile         bool           `db:"mobile"`
	ScreenWidth    int            `db:"screen_width"`
	ScreenHeight   int            `db:"screen_height"`
	ScreenClass    string         `db:"screen_class"`
	UTMSource      sql.NullString `db:"utm_source"`
	UTMMedium      sql.NullString `db:"utm_medium"`
	UTMCampaign    sql.NullString `db:"utm_campaign"`
	UTMContent     sql.NullString `db:"utm_content"`
	UTMTerm        sql.NullString `db:"utm_term"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// Stats is the base entity for all statistics.
type Stats struct {
	Day                     time.Time      `db:"day" json:"day,omitempty"`
	Path                    sql.NullString `db:"path" json:"path,omitempty"`
	Referrer                sql.NullString `db:"referrer" json:"referrer,omitempty"`
	ReferrerName            sql.NullString `db:"referrer_name" json:"referrer_name,omitempty"`
	ReferrerIcon            sql.NullString `db:"referrer_icon" json:"referrer_icon,omitempty"`
	Language                sql.NullString `db:"language" json:"language,omitempty"`
	CountryCode             sql.NullString `db:"country_code" json:"country_code,omitempty"`
	Browser                 sql.NullString `db:"browser" json:"browser,omitempty"`
	OS                      sql.NullString `db:"os" json:"os,omitempty"`
	Views                   int            `db:"views" json:"views,omitempty"`
	Visitors                int            `db:"visitors" json:"visitors,omitempty"`
	Sessions                int            `db:"sessions" json:"sessions,omitempty"`
	Bounces                 int            `db:"bounces" json:"bounces,omitempty"`
	RelativeViews           float64        `db:"relative_views" json:"relative_views,omitempty"`
	RelativeVisitors        float64        `db:"relative_visitors" json:"relative_visitors,omitempty"`
	BounceRate              float64        `db:"bounce_rate" json:"bounce_rate,omitempty"`
	ScreenWidth             int            `db:"screen_width" json:"screen_width,omitempty"`
	ScreenHeight            int            `db:"screen_height" json:"screen_height,omitempty"`
	ScreenClass             sql.NullString `db:"screen_class" json:"screen_class,omitempty"`
	PlatformDesktop         int            `db:"platform_desktop" json:"platform_desktop,omitempty"`
	PlatformMobile          int            `db:"platform_mobile" json:"platform_mobile,omitempty"`
	PlatformUnknown         int            `db:"platform_unknown" json:"platform_unknown,omitempty"`
	RelativePlatformDesktop float64        `db:"relative_platform_desktop" json:"relative_platform_desktop,omitempty"`
	RelativePlatformMobile  float64        `db:"relative_platform_mobile" json:"relative_platform_mobile,omitempty"`
	RelativePlatformUnknown float64        `db:"relative_platform_unknown" json:"relative_platform_unknown,omitempty"`
	AverageTimeSpendSeconds int            `db:"average_time_spend_seconds" json:"average_time_spend_seconds,omitempty"`
}
