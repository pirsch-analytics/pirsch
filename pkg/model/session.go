package model

import (
	"encoding/json"
	"time"
)

// Session stores updates to a visitor session.
type Session struct {
	Data

	Sign            int8      `json:"sign" csv:"sign"`
	Version         uint16    `json:"version" csv:"version"`
	Start           time.Time `json:"start" csv:"start"`
	DurationSeconds uint32    `db:"duration_seconds" json:"duration_seconds" csv:"duration_seconds"`
	PageViews       uint16    `db:"page_views" json:"page_views" csv:"page_views"`
	IsBounce        bool      `db:"is_bounce" json:"is_bounce" csv:"is_bounce"`
	EntryPath       string    `db:"entry_path" json:"entry_path" csv:"entry_path"`
	ExitPath        string    `db:"exit_path" json:"exit_path" csv:"exit_path"`
	EntryTitle      string    `db:"entry_title" json:"entry_title" csv:"entry_title"`
	ExitTitle       string    `db:"exit_title" json:"exit_title" csv:"exit_title"`
	Extended        uint16    `json:"extended" csv:"extended"`
}

// String implements the Stringer interface.
func (session Session) String() string {
	out, _ := json.Marshal(session)
	return string(out)
}
