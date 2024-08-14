package null

import (
	"database/sql"
	"encoding/json"
)

// String is a nullable string type based on sql.NullString, that supports parsing to/from JSON.
type String struct {
	sql.NullString
}

// NewString returns a new nullable String object.
// This is equivalent to `null.String{sql.NullString{String: s, Valid: valid}}`.
func NewString(s string, valid bool) String {
	return String{sql.NullString{String: s, Valid: valid}}
}

// MarshalJSON implements the encoding json interface.
func (s String) MarshalJSON() ([]byte, error) {
	if s.Valid {
		return json.Marshal(s.String)
	}

	return json.Marshal(nil)
}

// UnmarshalJSON implements the encoding json interface.
func (s *String) UnmarshalJSON(data []byte) error {
	var value *string

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		s.SetValid(*value)
	} else {
		s.SetNil()
	}

	return nil
}

// SetValid sets the value and valid to true.
func (s *String) SetValid(value string) {
	s.String = value
	s.Valid = true
}

// SetNil sets the value to default and valid to false.
func (s *String) SetNil() {
	s.String = ""
	s.Valid = false
}
