package model

import (
	"encoding/json"
)

// Event stores custom events.
type Event struct {
	Data

	Name     string         `json:"name" csv:"name"`
	MetaData map[string]any `db:"meta_data" json:"meta_data" csv:"-"` // TODO csv
	Path     string         `json:"path" csv:"path"`
	Title    string         `json:"title" csv:"title"`
}

// String implements the Stringer interface.
func (event Event) String() string {
	out, _ := json.Marshal(event)
	return string(out)
}
