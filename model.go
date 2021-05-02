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
	Day                     time.Time      `json:"day"`
	Hour                    int            `json:"hour"`
	Path                    sql.NullString `json:"path"`
	Referrer                sql.NullString `json:"referrer"`
	ReferrerName            sql.NullString `db:"referrer_name" json:"referrer_name"`
	ReferrerIcon            sql.NullString `db:"referrer_icon" json:"referrer_icon"`
	Language                sql.NullString `json:"language"`
	CountryCode             sql.NullString `db:"country_code" json:"country_code"`
	Browser                 sql.NullString `json:"browser"`
	OS                      sql.NullString `json:"os"`
	Views                   int            `json:"views"`
	Visitors                int            `json:"visitors"`
	Sessions                int            `json:"sessions"`
	Bounces                 int            `json:"bounces"`
	RelativeViews           float64        `db:"relative_views" json:"relative_views"`
	RelativeVisitors        float64        `db:"relative_visitors" json:"relative_visitors"`
	BounceRate              float64        `db:"bounce_rate" json:"bounce_rate"`
	ScreenWidth             int            `db:"screen_width" json:"screen_width"`
	ScreenHeight            int            `db:"screen_height" json:"screen_height"`
	ScreenClass             sql.NullString `db:"screen_class" json:"screen_class"`
	UTMSource               sql.NullString `db:"utm_source" json:"utm_source"`
	UTMMedium               sql.NullString `db:"utm_medium" json:"utm_medium"`
	UTMCampaign             sql.NullString `db:"utm_campaign" json:"utm_campaign"`
	UTMContent              sql.NullString `db:"utm_content" json:"utm_content"`
	UTMTerm                 sql.NullString `db:"utm_term" json:"utm_term"`
	PlatformDesktop         int            `db:"platform_desktop" json:"platform_desktop"`
	PlatformMobile          int            `db:"platform_mobile" json:"platform_mobile"`
	PlatformUnknown         int            `db:"platform_unknown" json:"platform_unknown"`
	RelativePlatformDesktop float64        `db:"relative_platform_desktop" json:"relative_platform_desktop"`
	RelativePlatformMobile  float64        `db:"relative_platform_mobile" json:"relative_platform_mobile"`
	RelativePlatformUnknown float64        `db:"relative_platform_unknown" json:"relative_platform_unknown"`
	AverageTimeSpentSeconds int            `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
}

// Growth represents the visitors, views, sessions, bounces, and average session duration growth between two time periods.
type Growth struct {
	VisitorsGrowth  float64 `json:"visitors_growth"`
	ViewsGrowth     float64 `json:"views_growth"`
	SessionsGrowth  float64 `json:"sessions_growth"`
	BouncesGrowth   float64 `json:"bounces_growth"`
	TimeSpentGrowth float64 `json:"time_spent_growth"`
}
