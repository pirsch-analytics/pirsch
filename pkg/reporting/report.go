package reporting

// Report returns metrics for a Request.
type Report struct {
	// Request is the Request for this report.
	Request Request

	// Results is an ordered list of query results.
	Results []Result
}

// Result is a result row.
type Result struct {
	// Dimensions is the ordered list of dimensions as in the Request.
	Dimensions []Dimension

	// Metrics is the ordered list of result values as in the Request.
	Metrics []ResultMetric
}

// ResultMetric is the result value in a Report.
type ResultMetric interface{}
