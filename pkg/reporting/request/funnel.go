package request

import (
	"context"
	"errors"
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

	if len(r.Filter) < 2 {
		return []error{errors.New("funnels must have at least two filter steps")}
	}

	for _, filter := range r.Filter {
		if len(filter) == 0 {
			return []error{errors.New("funnel step filters must not be empty")}
		}
	}

	// TODO

	return nil
}
