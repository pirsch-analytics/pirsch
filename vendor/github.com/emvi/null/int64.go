package null

import (
	"database/sql"
	"encoding/json"
)

// Int64 is a nullable int64 type based on sql.NullInt64, that supports parsing to/from JSON.
type Int64 struct {
	sql.NullInt64
}

// NewInt64 returns a new nullable Int64 object.
// This is equivalent to `null.Int64{sql.NullInt64{Int64: i, Valid: valid}}`.
func NewInt64(i int64, valid bool) Int64 {
	return Int64{sql.NullInt64{Int64: i, Valid: valid}}
}

// MarshalJSON implements the encoding json interface.
func (i Int64) MarshalJSON() ([]byte, error) {
	if i.Valid {
		return json.Marshal(i.Int64)
	}

	return json.Marshal(nil)
}

// UnmarshalJSON implements the encoding json interface.
func (i *Int64) UnmarshalJSON(data []byte) error {
	var value *int64

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value != nil {
		i.SetValid(*value)
	} else {
		i.SetNil()
	}

	return nil
}

// SetValid sets the value and valid to true.
func (i *Int64) SetValid(value int64) {
	i.Int64 = value
	i.Valid = true
}

// SetNil sets the value to default and valid to false.
func (i *Int64) SetNil() {
	i.Int64 = 0
	i.Valid = false
}
