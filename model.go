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

// Stats combines statistical data.
type Stats struct {
	Day                     time.Time      `json:"day,omitempty"`
	Hour                    int            `json:"hour,omitempty"`
	Path                    sql.NullString `json:"path,omitempty"`
	Referrer                sql.NullString `json:"referrer,omitempty"`
	ReferrerName            sql.NullString `db:"referrer_name" json:"referrer_name,omitempty"`
	ReferrerIcon            sql.NullString `db:"referrer_icon" json:"referrer_icon,omitempty"`
	Language                sql.NullString `json:"language,omitempty"`
	CountryCode             sql.NullString `db:"country_code" json:"country_code,omitempty"`
	Browser                 sql.NullString `json:"browser,omitempty"`
	OS                      sql.NullString `json:"os,omitempty"`
	Views                   int            `json:"views,omitempty"`
	Visitors                int            `json:"visitors,omitempty"`
	Sessions                int            `json:"sessions,omitempty"`
	Bounces                 int            `json:"bounces,omitempty"`
	RelativeViews           float64        `db:"relative_views" json:"relative_views,omitempty"`
	RelativeVisitors        float64        `db:"relative_visitors" json:"relative_visitors,omitempty"`
	BounceRate              float64        `db:"bounce_rate" json:"bounce_rate,omitempty"`
	ScreenWidth             int            `db:"screen_width" json:"screen_width,omitempty"`
	ScreenHeight            int            `db:"screen_height" json:"screen_height,omitempty"`
	ScreenClass             sql.NullString `db:"screen_class" json:"screen_class,omitempty"`
	UTMSource               sql.NullString `db:"utm_source"`
	UTMMedium               sql.NullString `db:"utm_medium"`
	UTMCampaign             sql.NullString `db:"utm_campaign"`
	UTMContent              sql.NullString `db:"utm_content"`
	UTMTerm                 sql.NullString `db:"utm_term"`
	PlatformDesktop         int            `db:"platform_desktop" json:"platform_desktop,omitempty"`
	PlatformMobile          int            `db:"platform_mobile" json:"platform_mobile,omitempty"`
	PlatformUnknown         int            `db:"platform_unknown" json:"platform_unknown,omitempty"`
	RelativePlatformDesktop float64        `db:"relative_platform_desktop" json:"relative_platform_desktop,omitempty"`
	RelativePlatformMobile  float64        `db:"relative_platform_mobile" json:"relative_platform_mobile,omitempty"`
	RelativePlatformUnknown float64        `db:"relative_platform_unknown" json:"relative_platform_unknown,omitempty"`
	AverageTimeSpentSeconds int            `db:"average_time_spent_seconds" json:"average_time_spent_seconds,omitempty"`
}

// Growth represents the visitors, views, sessions, bounces, and average session duration growth between two time periods.
type Growth struct {
	VisitorsGrowth  float64 `json:"visitors_growth"`
	ViewsGrowth     float64 `json:"views_growth"`
	SessionsGrowth  float64 `json:"sessions_growth"`
	BouncesGrowth   float64 `json:"bounces_growth"`
	TimeSpentGrowth float64 `json:"time_spent_growth"`
}
