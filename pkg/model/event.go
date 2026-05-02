package model

import (
	"encoding/json"
)

// Event stores custom events.
type Event struct {
	Data

	Name     string         `json:"name"`
	MetaData map[string]any `db:"meta_data" json:"meta_data"`
	Path     string         `json:"path"`
	Title    string         `json:"title"`
}

// String implements the Stringer interface.
func (event Event) String() string {
	out, _ := json.Marshal(event)
	return string(out)
}
