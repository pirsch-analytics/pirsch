package model

import (
	"encoding/json"
	"time"
)

// Bot represents a visitor or event that has been ignored.
// The creation time, User-Agent, path, and event name are stored in the database to find bots.
type Bot struct {
	ClientID  uint64    `db:"client_id" json:"client_id"`
	VisitorID uint64    `db:"visitor_id" json:"visitor_id"`
	Time      time.Time `json:"time"`
	UserAgent string    `db:"user_agent"`
	Path      string    `json:"path"`
	Event     string    `db:"event_name" json:"event"`
}

// String implements the Stringer interface.
func (bot Bot) String() string {
	out, _ := json.Marshal(bot)
	return string(out)
}
