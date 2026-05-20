package report

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
)

// Report returns metrics for a Request.
type Report struct {
	// Request is the Request for this report.
	Request request.Request

	// Results is an ordered list of query results.
	Results []Result

	// Meta contains metadata information for this report.
	Meta Meta
}

// Result is a result row.
type Result struct {
	// Dimensions is the ordered list of dimensions as in the Request.
	Dimensions []dimensions.Dimension

	// Metrics is the ordered list of result values as in the Request.
	Metrics []ResultMetric
}

// ResultMetric is the result value in a Report.
type ResultMetric interface{}

// Meta contains metadata information for the Report.
type Meta struct {
	// Errors is a list of errors.
	Errors []error
}
