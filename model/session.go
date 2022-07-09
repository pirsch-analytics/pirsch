package model

import (
	"encoding/json"
	"time"
)

// Session represents a single visitor.
type Session struct {
	Sign            int8      `json:"sign"`
	ClientID        uint64    `db:"client_id" json:"client_id"`
	VisitorID       uint64    `db:"visitor_id" json:"visitor_id"`
	SessionID       uint32    `db:"session_id" json:"session_id"`
	Time            time.Time `json:"time"`
	Start           time.Time `json:"start"`
	DurationSeconds uint32    `db:"duration_seconds" json:"duration_seconds"`
	EntryPath       string    `db:"entry_path" json:"entry_path"`
	ExitPath        string    `db:"exit_path" json:"exit_path"`
	PageViews       uint16    `db:"page_views" json:"page_views"`
	IsBounce        bool      `db:"is_bounce" json:"is_bounce"`
	EntryTitle      string    `db:"entry_title" json:"entry_title"`
	ExitTitle       string    `db:"exit_title" json:"exit_title"`
	Language        string    `json:"language"`
	CountryCode     string    `db:"country_code" json:"country_code"`
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
	ScreenWidth     uint16    `db:"screen_width" json:"screen_width"`
	ScreenHeight    uint16    `db:"screen_height" json:"screen_height"`
	ScreenClass     string    `db:"screen_class" json:"screen_class"`
	UTMSource       string    `db:"utm_source" json:"utm_source"`
	UTMMedium       string    `db:"utm_medium" json:"utm_medium"`
	UTMCampaign     string    `db:"utm_campaign" json:"utm_campaign"`
	UTMContent      string    `db:"utm_content" json:"utm_content"`
	UTMTerm         string    `db:"utm_term" json:"utm_term"`
	IsBot           uint8     `db:"is_bot" json:"is_bot"`
}

// String implements the Stringer interface.
func (session Session) String() string {
	out, _ := json.Marshal(session)
	return string(out)
}
