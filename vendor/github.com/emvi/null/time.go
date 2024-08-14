package null

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Time is a nullable time.Time, that supports parsing to/from JSON.
type Time struct {
	sql.NullTime
}

// NewTime returns a new nullable time.Time object.
// This is equivalent to `null.Time{Time: t, Valid: valid}`.
func NewTime(t time.Time, valid bool) Time {
	return Time{sql.NullTime{Time: t, Valid: valid}}
}

// MarshalJSON implements the encoding json interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.Valid {
		return json.Marshal(t.Time)
	}

	return json.Marshal(nil)
}

// UnmarshalJSON implements the encoding json interface.
func (t *Time) UnmarshalJSON(data []byte) error {
	var value time.Time

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if !value.IsZero() {
		t.SetValid(value)
	} else {
		t.SetNil()
	}

	return nil
}

// Scan implements the Scanner interface.
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		t.SetNil()
		return nil
	}

	v, ok := value.(time.Time)

	if !ok {
		return errors.New("unexpected type")
	}

	t.SetValid(v)
	return nil
}

// Value implements the driver Valuer interface.
func (t Time) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}

	return t.Time, nil
}

// SetValid sets the value and valid to true.
func (t *Time) SetValid(value time.Time) {
	t.Time = value
	t.Valid = true
}

// SetNil sets the value to default and valid to false.
func (t *Time) SetNil() {
	t.Time = time.Time{}
	t.Valid = false
}
