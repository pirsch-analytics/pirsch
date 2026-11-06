package request

import (
	"encoding/json/v2"
	"fmt"
	"time"
)

var (
	WeekdayMonday WeekdayMode = 1
	WeekdaySunday WeekdayMode = 2
)

// WeekdayMode sets the start day of the week (WeekdayMonday or WeekdaySunday).
type WeekdayMode int

// Period is the start and end date and time for the Request, as well as the timezone.
type Period struct {
	// From is the start date and time for the Period.
	// The value is interpreted as wall clock time in Timezone. The location of the time.Time itself is ignored,
	// as the date (and time) is passed to ClickHouse as a string and converted to Timezone there.
	From time.Time

	// To is the end date and time for the Period (inclusive).
	// The value is interpreted as wall clock time in Timezone, see From.
	To time.Time

	// Timezone is the timezone for the Request and Report.
	// It is set to UTC by default and must be a valid IANA timezone (time.Local is not supported).
	Timezone *time.Location

	// WeekdayMode sets the start day of the week (WeekdayMonday or WeekdaySunday).
	// WeekdayMonday by default.
	WeekdayMode WeekdayMode

	// IncludeTime determines whether the From and To times should be included.
	IncludeTime bool

	// Compare is optional. When set, the query runs for both periods and results are merged.
	Compare *ComparePeriod

	// ImportedUntil includes imported statistics when set up to this date.
	// Native statistics are used from this date onwards.
	ImportedUntil time.Time
}

// ComparePeriod is the start and end date for a previous period for the Request.
type ComparePeriod struct {
	// From is the start date and time for the ComparePeriod.
	From time.Time

	// To is the end date and time for the ComparePeriod.
	To time.Time
}

func (p *Period) validate() error {
	if p.Timezone == nil {
		p.Timezone = time.UTC
	}

	if p.WeekdayMode != WeekdayMonday && p.WeekdayMode != WeekdaySunday {
		p.WeekdayMode = WeekdayMonday
	}

	// only swap if both are set, so that open-ended periods stay open-ended
	if !p.From.IsZero() && !p.To.IsZero() && p.From.After(p.To) {
		p.From, p.To = p.To, p.From
	}

	if p.Compare != nil && !p.Compare.From.IsZero() && !p.Compare.To.IsZero() && p.Compare.From.After(p.Compare.To) {
		p.Compare.From, p.Compare.To = p.Compare.To, p.Compare.From
	}

	return validateTimezone(p.Timezone)
}

func validateTimezone(tz *time.Location) error {
	name := tz.String()

	if name == "" || name == "Local" {
		return fmt.Errorf("timezone '%s' is not a valid IANA timezone name", name)
	}

	if _, err := time.LoadLocation(name); err != nil {
		return fmt.Errorf("timezone '%s' is not a valid IANA timezone name", name)
	}

	return nil
}

// String implements the fmt.Stringer interface.
func (p *Period) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// String implements the fmt.Stringer interface.
func (c *ComparePeriod) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}
