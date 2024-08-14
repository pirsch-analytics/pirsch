package null

import (
	"database/sql"
	"encoding/json"
)

// Int32 is a nullable int32 type based on sql.NullInt32, that supports parsing to/from JSON.
type Int32 struct {
	sql.NullInt32
}

// NewInt32 returns a new nullable Int32 object.
// This is equivalent to `null.Int32{sql.NullInt32{Int32: i, Valid: valid}}`.
func NewInt32(i int32, valid bool) Int32 {
	return Int32{sql.NullInt32{Int32: i, Valid: valid}}
}

// MarshalJSON implements the encoding json interface.
func (i Int32) MarshalJSON() ([]byte, error) {
	if i.Valid {
		return json.Marshal(i.Int32)
	}

	return json.Marshal(nil)
}

// UnmarshalJSON implements the encoding json interface.
func (i *Int32) UnmarshalJSON(data []byte) error {
	var value *int32

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
func (i *Int32) SetValid(value int32) {
	i.Int32 = value
	i.Valid = true
}

// SetNil sets the value to default and valid to false.
func (i *Int32) SetNil() {
	i.Int32 = 0
	i.Valid = false
}
