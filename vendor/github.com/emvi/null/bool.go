package null

import (
	"database/sql"
	"encoding/json"
)

// Bool is a nullable boolean type based on sql.NullBool, that supports parsing to/from JSON.
type Bool struct {
	sql.NullBool
}

// NewBool returns a new nullable Bool object.
// This is equivalent to `null.Bool{sql.NullBool{Bool: b, Valid: valid}}`.
func NewBool(b, valid bool) Bool {
	return Bool{sql.NullBool{Bool: b, Valid: valid}}
}

// MarshalJSON implements the encoding json interface.
func (b Bool) MarshalJSON() ([]byte, error) {
	if b.Valid {
		return json.Marshal(b.Bool)
	}

	return json.Marshal(nil)
}

// UnmarshalJSON implements the encoding json interface.
func (b *Bool) UnmarshalJSON(data []byte) error {
	var value *bool

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		b.SetValid(*value)
	} else {
		b.SetNil()
	}

	return nil
}

// SetValid sets the value and valid to true.
func (b *Bool) SetValid(value bool) {
	b.Bool = value
	b.Valid = true
}

// SetNil sets the value to default and valid to false.
func (b *Bool) SetNil() {
	b.Bool = false
	b.Valid = false
}
