package report

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
)

// FunnelReport returns the results for a request.FunnelRequest.
type FunnelReport struct {
	// Request is the request.FunnelRequest for this report.
	Request request.FunnelRequest

	// Steps is an ordered list of funnel steps.
	Steps []FunnelStep

	// Meta contains metadata information for this report.
	Meta Meta
}

// FunnelStep is the statistics for a funnel step.
type FunnelStep struct {
	// Step is the number of the step starting at 1.
	Step int

	// Visitors is the unique number of visitors.
	Visitors int

	// RelativeVisitors is the relative number of visitors.
	RelativeVisitors float64

	// PreviousVisitors is the unique number of visitors from the previous step.
	PreviousVisitors int

	// RelativePreviousVisitors is the relative number of visitors from the previous step.
	RelativePreviousVisitors float64

	// Dropped is the unique number of visitors dropped from the previous step.
	Dropped int

	// DropOff is the relative number of visitors dropped from the previous step.
	DropOff float64
}
