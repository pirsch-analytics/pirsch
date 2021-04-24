package pirsch

import (
	"database/sql"
	"time"
)

// Filter are all fields that can be used to filter the result sets.
type Filter struct {
	// TenantID is the optional.
	TenantID sql.NullInt64

	// From is the start of the selected period.
	From time.Time

	// To is the end of the selected period.
	To time.Time

	// Day is an exact match for the result set ("on this day").
	Day time.Time

	// Path filters for the path.
	Path string
}

// NewFilter creates a new filter for given tenant ID.
func NewFilter(tenantID sql.NullInt64) *Filter {
	return &Filter{
		TenantID: tenantID,
	}
}

func (filter *Filter) validate() {
	if !filter.From.IsZero() {
		filter.From = filter.toUTCDate(filter.From)
	}

	if !filter.To.IsZero() {
		filter.To = filter.toUTCDate(filter.To)
	}

	if !filter.Day.IsZero() {
		filter.Day = filter.toUTCDate(filter.Day)
	}

	if !filter.To.IsZero() && filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}

	today := Today()

	if !filter.To.IsZero() && filter.To.After(today) {
		filter.To = today
	}
}

func (filter *Filter) toUTCDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}
