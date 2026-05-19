package reporting

import (
	"context"
)

// Request generates a Report.
type Request struct {
	// Ctx can be used to set a timeout or to cancel queries.
	Ctx context.Context

	// SiteID is the site ID for the request.
	SiteID int64

	// Period is the period and timezone for the Report.
	Period Period

	// Metrics is a list of result fields for the Report.
	Metrics []Metric

	// Dimensions is a list of attributes for the Request.
	Dimensions []Dimension

	// Filter filters the results for a Report.
	Filter Filter

	// OrderBy sorts the result fields of a Report.
	OrderBy []OrderBy

	// Pagination limits the number of results for a Report.
	Pagination *Pagination

	// Options are optional fields for the Request.
	Options *Options
}

func (r *Request) validate() {
	// TODO
}
