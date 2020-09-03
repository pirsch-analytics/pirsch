package pirsch

import (
	"database/sql"
	"time"
)

// Filter is used to specify the time frame, path and tenant for the Analyzer.
type Filter struct {
	// TenantID is the optional tenant ID used to filter results.
	TenantID sql.NullInt64

	// Path is the optional path for the selection.
	Path string

	// From is the start of the selection.
	From time.Time

	// To is the end of the selection.
	To time.Time
}

// NewFilter returns a new default filter for given tenant and the past week.
func NewFilter(tenantID sql.NullInt64) *Filter {
	today := today()
	return &Filter{
		TenantID: tenantID,
		From:     today.Add(-time.Hour * 24 * 6), // 7 including today
		To:       today,
	}
}

// Days returns the number of days covered by the filter.
func (filter *Filter) Days() int {
	return int(filter.To.Sub(filter.From).Hours()) / 24
}

func (filter *Filter) validate() {
	today := today()

	if filter.From.IsZero() && filter.To.IsZero() {
		filter.From = today.Add(-time.Hour * 24 * 6) // 7 including today
		filter.To = today
	} else {
		filter.From = time.Date(filter.From.Year(), filter.From.Month(), filter.From.Day(), 0, 0, 0, 0, time.UTC)
		filter.To = time.Date(filter.To.Year(), filter.To.Month(), filter.To.Day(), 0, 0, 0, 0, time.UTC)
	}

	if filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}
}
