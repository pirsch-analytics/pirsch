package pirsch

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Hit represents a single data point/page visit and is the central entity of Pirsch.
type Hit struct {
	TenantID       sql.NullInt64  `db:"tenant_id" json:"tenant_id,omitempty"`
	Fingerprint    string         `db:"fingerprint" json:"fingerprint"`
	Time           time.Time      `db:"time" json:"time"`
	Session        sql.NullTime   `db:"session" json:"session,omitempty"`
	UserAgent      string         `db:"user_agent" json:"user_agent"`
	Path           string         `db:"path" json:"path"`
	URL            string         `db:"url" json:"url"`
	Language       sql.NullString `db:"language" json:"language,omitempty"`
	CountryCode    sql.NullString `db:"country_code" json:"country_code,omitempty"`
	Referrer       sql.NullString `db:"referrer" json:"referrer,omitempty"`
	ReferrerName   sql.NullString `db:"referrer_name" json:"referrer_name,omitempty"`
	ReferrerIcon   sql.NullString `db:"referrer_icon" json:"referrer_icon,omitempty"`
	OS             sql.NullString `db:"os" json:"os,omitempty"`
	OSVersion      sql.NullString `db:"os_version" json:"os_version,omitempty"`
	Browser        sql.NullString `db:"browser" json:"browser,omitempty"`
	BrowserVersion sql.NullString `db:"browser_version" json:"browser_version,omitempty"`
	Desktop        bool           `db:"desktop" json:"desktop"`
	Mobile         bool           `db:"mobile" json:"mobile"`
	ScreenWidth    int            `db:"screen_width" json:"screen_width"`
	ScreenHeight   int            `db:"screen_height" json:"screen_height"`
	ScreenClass    sql.NullString `db:"screen_class" json:"screen_class,omitempty"`
	UTMSource      sql.NullString `db:"utm_source" json:"utm_source,omitempty"`
	UTMMedium      sql.NullString `db:"utm_medium" json:"utm_medium,omitempty"`
	UTMCampaign    sql.NullString `db:"utm_campaign" json:"utm_campaign,omitempty"`
	UTMContent     sql.NullString `db:"utm_content" json:"utm_content,omitempty"`
	UTMTerm        sql.NullString `db:"utm_term" json:"utm_term,omitempty"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}
