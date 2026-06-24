package request

import (
	"context"
	"regexp"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
)

var (
	metaKeyValidKeySegment = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Request generates a report.Report.
type Request struct {
	// Ctx can be used to set a timeout or to cancel queries.
	Ctx context.Context

	// SiteID is the site ID for the request.
	SiteID uint64

	// Period is the period and timezone for the report.Report.
	Period Period

	// Metrics is a list of result fields for the report.Report.
	Metrics []metrics.Metric

	// Dimensions is a list of attributes for the Request.
	Dimensions []dimensions.Dimension

	// Filter filters the results for a report.Report.
	// Top-level filters are connected using the AND operator by default.
	// To use other operators, they need to be set in the Filter recursively.
	Filter []Filter

	// OrderBy sorts the result fields of a report.Report.
	OrderBy []OrderBy

	// Pagination limits the number of results for a report.Report.
	Pagination *Pagination

	// Options are optional fields for the Request.
	Options *Options
}

// Validate validates the Request and returns an error if a report.Report cannot be constructed for the specified fields.
func (r *Request) Validate() []error {
	if r.Ctx == nil {
		r.Ctx = context.Background()
	}

	if r.Options == nil {
		r.Options = new(Options)
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

	if r.Period.Compare != nil && r.Period.Compare.From.After(r.Period.Compare.To) {
		r.Period.Compare.From, r.Period.Compare.To = r.Period.Compare.To, r.Period.Compare.From
	}

	errs := make([]error, 0)

	if err := validateSiteID(r.SiteID); err != nil {
		errs = append(errs, err)
	}

	for _, f := range r.Filter {
		errs = append(errs, validateFilterValues(f)...)
	}

	errs = append(errs, validateOrderBy(r.OrderBy, r.Dimensions, r.Metrics)...)
	// TODO check other relevant fields and filter combinations

	if len(errs) > 0 {
		return errs
	}

	return nil
}
