package model

import (
	"encoding/json"
	"time"
)

// Data is a shared type for Session, PageView, and Event.
type Data struct {
	SiteID         uint64    `db:"site_id" json:"site_id" csv:"site_id"`
	VisitorID      uint64    `db:"visitor_id" json:"visitor_id" csv:"visitor_id"`
	SessionID      uint32    `db:"session_id" json:"session_id" csv:"session_id"`
	Time           time.Time `json:"time" csv:"time"`
	Hostname       string    `json:"hostname" csv:"hostname"`
	Language       string    `json:"language" csv:"language"`
	CountryCode    string    `db:"country_code" json:"country_code" csv:"country_code"`
	Region         string    `json:"region" csv:"region"`
	City           string    `json:"city" csv:"city"`
	Referrer       string    `json:"referrer" csv:"referrer"`
	ReferrerName   string    `db:"referrer_name" json:"referrer_name" csv:"referrer_name"`
	ReferrerIcon   string    `db:"referrer_icon" json:"referrer_icon" csv:"referrer_icon"`
	OS             string    `json:"os" csv:"os"`
	OSVersion      string    `db:"os_version" json:"os_version" csv:"os_version"`
	Browser        string    `json:"browser" csv:"browser"`
	BrowserVersion string    `db:"browser_version" json:"browser_version" csv:"browser_version"`
	Platform       int8      `json:"platform" csv:"platform"`
	ScreenClass    string    `db:"screen_class" json:"screen_class" csv:"screen_class"`
	UTMSource      string    `db:"utm_source" json:"utm_source" csv:"utm_source"`
	UTMMedium      string    `db:"utm_medium" json:"utm_medium" csv:"utm_medium"`
	UTMCampaign    string    `db:"utm_campaign" json:"utm_campaign" csv:"utm_campaign"`
	UTMContent     string    `db:"utm_content" json:"utm_content" csv:"utm_content"`
	UTMTerm        string    `db:"utm_term" json:"utm_term" csv:"utm_term"`
	Channel        string    `json:"channel" csv:"channel"`
}

// String implements the Stringer interface.
func (data Data) String() string {
	out, _ := json.Marshal(data)
	return string(out)
}
