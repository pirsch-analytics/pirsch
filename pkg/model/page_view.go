package model

import (
	"encoding/json"
)

// PageView stores page views.
type PageView struct {
	Data

	DurationSeconds uint32            `db:"duration_seconds" json:"duration_seconds"`
	Path            string            `json:"path"`
	Title           string            `json:"title"`
	Tags            map[string]string `db:"tags" json:"tags"`
}

// String implements the Stringer interface.
func (pageView PageView) String() string {
	out, _ := json.Marshal(pageView)
	return string(out)
}
