package pirsch

import (
	"database/sql"
	"strings"
	"time"
)

// Filter is used to specify the time frame, path and tenant for the Analyzer.
type Filter struct {
	// TenantID is the optional tenant ID used to filter results.
	TenantID sql.NullInt64

	// From is the start of the selection.
	From time.Time

	// To is the end of the selection.
	To time.Time

	// Day is the day for the selection.
	Day time.Time

	// Path is the optional path for the selection.
	Path string
}

// NewFilter returns a new default filter for given tenant.
func NewFilter(tenantID sql.NullInt64) *Filter {
	return &Filter{TenantID: tenantID}
}

// Days returns the number of days covered by the filter.
func (filter *Filter) Days() int {
	return int(filter.To.Sub(filter.From).Hours()) / 24
}

func (filter *Filter) validate() {
	today := Today()

	if !filter.From.IsZero() {
		filter.From = time.Date(filter.From.Year(), filter.From.Month(), filter.From.Day(), 0, 0, 0, 0, time.UTC)
	}

	if !filter.To.IsZero() {
		filter.To = time.Date(filter.To.Year(), filter.To.Month(), filter.To.Day(), 0, 0, 0, 0, time.UTC)
	}

	if !filter.To.IsZero() && filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}

	if !filter.To.IsZero() && filter.To.After(today) {
		filter.To = today
	}

	if !filter.Day.IsZero() {
		filter.Day = time.Date(filter.Day.Year(), filter.Day.Month(), filter.Day.Day(), 0, 0, 0, 0, time.UTC)
	}

	filter.Path = strings.TrimSpace(filter.Path)
}
