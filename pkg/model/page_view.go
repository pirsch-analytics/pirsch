package model

import (
	"encoding/json"
	"time"
)

// PageView represents a single page visit.
type PageView struct {
	ClientID        uint64    `db:"client_id" json:"client_id"`
	VisitorID       uint64    `db:"visitor_id" json:"visitor_id"`
	SessionID       uint32    `db:"session_id" json:"session_id"`
	Time            time.Time `json:"time"`
	DurationSeconds uint32    `db:"duration_seconds" json:"duration_seconds"`
	Hostname        string    `json:"hostname"`
	Path            string    `json:"path"`
	Title           string    `json:"title"`
	Language        string    `json:"language"`
	CountryCode     string    `db:"country_code" json:"country_code"`
	Region          string    `json:"region"`
	City            string    `json:"city"`
	Referrer        string    `json:"referrer"`
	ReferrerName    string    `db:"referrer_name" json:"referrer_name"`
	ReferrerIcon    string    `db:"referrer_icon" json:"referrer_icon"`
	OS              string    `json:"os"`
	OSVersion       string    `db:"os_version" json:"os_version"`
	Browser         string    `json:"browser"`
	BrowserVersion  string    `db:"browser_version" json:"browser_version"`
	Desktop         bool      `json:"desktop"`
	Mobile          bool      `json:"mobile"`
	ScreenClass     string    `db:"screen_class" json:"screen_class"`
	UTMSource       string    `db:"utm_source" json:"utm_source"`
	UTMMedium       string    `db:"utm_medium" json:"utm_medium"`
	UTMCampaign     string    `db:"utm_campaign" json:"utm_campaign"`
	UTMContent      string    `db:"utm_content" json:"utm_content"`
	UTMTerm         string    `db:"utm_term" json:"utm_term"`
	Channel         string    `json:"channel"`
	TagKeys         []string  `db:"tag_keys" json:"tag_keys"`
	TagValues       []string  `db:"tag_values" json:"tag_values"`
}

// String implements the Stringer interface.
func (pageView PageView) String() string {
	out, _ := json.Marshal(pageView)
	return string(out)
}
