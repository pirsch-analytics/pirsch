package dimensions

// Dimension is a field results are grouped by.
type Dimension interface {
	// Table returns the valid database tables for the Dimension.
	Table() []string

	// Column returns the database column name for the given table for the Dimension.
	// This also handles "joins" by potentially returning an entirely different column (like entry_path instead of path for bounces).
	Column(string) string

	// Expression returns the SQL expression for aggregation.
	// If empty, the Column name will be used instead.
	Expression() string

	// Args returns optional arguments for the Expression.
	Args() []any

	// ScanType returns a pointer to the type the value for this Dimension scans into.
	ScanType() any
}

// TODO
/*
	CustomMetricKey string
	CustomMetricType pkg.CustomMetricType
	VisitorID uint64
	SessionID uint32
*/
