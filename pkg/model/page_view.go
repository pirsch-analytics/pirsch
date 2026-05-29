package model

import (
	"encoding/json"
)

// PageView stores page views.
type PageView struct {
	Data

	DurationSeconds uint32            `db:"duration_seconds" json:"duration_seconds" csv:"duration_seconds"`
	Path            string            `json:"path" csv:"path"`
	Title           string            `json:"title" csv:"title"`
	Tags            map[string]string `db:"tags" json:"tags" csv:"-"` // TODO csv
}

// String implements the Stringer interface.
func (pageView PageView) String() string {
	out, _ := json.Marshal(pageView)
	return string(out)
}
