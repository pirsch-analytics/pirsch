package null

import (
	"database/sql"
	"encoding/json"
)

// Float64 is a nullable float64 type based on sql.NullFloat64, that supports parsing to/from JSON.
type Float64 struct {
	sql.NullFloat64
}

// NewFloat64 returns a new nullable Float64 object.
// This is equivalent to `null.Float64{sql.NullFloat64{Float64: f, Valid: valid}}`.
func NewFloat64(f float64, valid bool) Float64 {
	return Float64{sql.NullFloat64{Float64: f, Valid: valid}}
}

// MarshalJSON implements the encoding json interface.
func (f Float64) MarshalJSON() ([]byte, error) {
	if f.Valid {
		return json.Marshal(f.Float64)
	}

	return json.Marshal(nil)
}

// UnmarshalJSON implements the encoding json interface.
func (f *Float64) UnmarshalJSON(data []byte) error {
	var value *float64

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		f.SetValid(*value)
	} else {
		f.SetNil()
	}

	return nil
}

// SetValid sets the value and valid to true.
func (f *Float64) SetValid(value float64) {
	f.Float64 = value
	f.Valid = true
}

// SetNil sets the value to default and valid to false.
func (f *Float64) SetNil() {
	f.Float64 = 0
	f.Valid = false
}
