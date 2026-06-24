package request

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// FunnelRequest generates a funnel.
type FunnelRequest struct {
	// Ctx can be used to set a timeout or to cancel queries.
	Ctx context.Context

	// SiteID is the site ID for the request.
	SiteID uint64

	// Period is the period and timezone for the report.FunnelReport.
	Period Period

	// Filter filters the results for a report.FunnelReport per step.
	// Top-level filters are connected using the AND operator by default.
	// To use other operators, they need to be set in the Filter recursively.
	Filter [][]Filter

	// Options are optional fields for the FunnelRequest.
	Options *Options
}

// Validate validates the FunnelRequest and returns an error if a report.FunnelReport cannot be constructed for the specified fields.
func (r *FunnelRequest) Validate() []error {
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

	errs := make([]error, 0)

	if err := validateSiteID(r.SiteID); err != nil {
		errs = append(errs, err)
	}

	if len(r.Filter) < 2 {
		errs = append(errs, errors.New("funnels must have at least two filter steps"))
	} else {
		for i, filter := range r.Filter {
			if len(filter) == 0 {
				errs = append(errs, fmt.Errorf("funnel step %d filters must not be empty", i+1))
			} else {
				for _, f := range filter {
					errs = append(errs, validateFilterValues(f)...)
				}
			}
		}
	}

	// TODO check other relevant fields and filter combinations

	if len(errs) > 0 {
		return errs
	}

	return nil
}
