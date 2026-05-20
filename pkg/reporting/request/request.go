package request

import (
	"context"
	"errors"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
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
	Metrics []metrics.Metric

	// Dimensions is a list of attributes for the Request.
	Dimensions []dimensions.Dimension

	// Filter filters the results for a Report.
	Filter Filter

	// OrderBy sorts the result fields of a Report.
	OrderBy []OrderBy

	// Pagination limits the number of results for a Report.
	Pagination *Pagination

	// Options are optional fields for the Request.
	Options *Options
}

// Validate validates the Request and returns an error if a report.Report cannot be constructed for the specified fields.
func (r *Request) Validate() []error {
	if r.Ctx == nil {
		r.Ctx = context.Background()
	}

	errs := make([]error, 0)

	if r.SiteID == 0 {
		errs = append(errs, errors.New("SiteID is required"))
	}

	if r.Period.Timezone == nil {
		r.Period.Timezone = time.UTC
	}

	if r.Period.WeekdayMode == 0 {
		r.Period.WeekdayMode = WeekdayMonday
	}

	if r.Period.From.After(r.Period.To) {
		r.Period.From, r.Period.To = r.Period.To, r.Period.From
	}

	// TODO check other relevant fields

	if len(errs) > 0 {
		return errs
	}

	// TODO validate metrics, dimensions, filter, ...
	return nil
}
