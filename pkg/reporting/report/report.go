package report

import (
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
	// DimensionValues is the ordered list of values of dimensions as in the Request.
	DimensionValues []any

	// MetricValues is the ordered list of result values as in the Request.
	// These can be strings, int, or float.
	MetricValues []any
}

// Meta contains metadata information for the Report.
type Meta struct {
	// Errors is a list of errors.
	Errors []error
}
