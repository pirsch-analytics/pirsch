package metrics

// Metric is an (aggregated) result field, like the number of visitors.
type Metric interface {
	// Table returns the valid database tables for the Dimension.
	Table() []string

	// Column returns the database column name for the Metric.
	Column() string

	// Expression returns the SQL expression for aggregation for the given table and if a subquery is required.
	// The subquery filters for the site_id and period.
	Expression(string) (string, bool)

	// ScanType returns a pointer to the type the value for this Metric scans into.
	ScanType() any
}
