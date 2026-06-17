package request

import "time"

var (
	WeekdayMonday WeekdayMode = 1
	WeekdaySunday WeekdayMode = 2
)

// WeekdayMode sets the start day of the week (WeekdayMonday or WeekdaySunday).
type WeekdayMode int

// Period is the start and end date and time for the Request, as well as the timezone.
type Period struct {
	// From is the start date and time for the Period.
	From time.Time

	// To is the end date and time for the Period.
	To time.Time

	// Timezone is the timezone for the Request and Report.
	// It is set to UTC by default.
	Timezone *time.Location

	// WeekdayMode sets the start day of the week (WeekdayMonday or WeekdaySunday).
	// WeekdayMonday by default.
	WeekdayMode WeekdayMode

	// IncludeTime determines whether the From and To times should be included.
	IncludeTime bool

	// Compare is optional. When set, the query runs for both periods and results are merged.
	Compare *ComparePeriod
}

// ComparePeriod is the start and end date for a previous period for the Request.
type ComparePeriod struct {
	// From is the start date and time for the ComparePeriod.
	From time.Time

	// To is the end date and time for the ComparePeriod.
	To time.Time
}

func (p *Period) validate() {
	if p.From.After(p.To) {
		p.From, p.To = p.To, p.From
	}

	if p.Timezone == nil {
		p.Timezone = time.UTC
	}

	if p.WeekdayMode != WeekdayMonday && p.WeekdayMode != WeekdaySunday {
		p.WeekdayMode = WeekdayMonday
	}
}
