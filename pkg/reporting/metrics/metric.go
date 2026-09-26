package metrics

// Metric is an (aggregated) result field, like the number of visitors.
type Metric interface {
	// Table returns the valid database tables for the Metric.
	Table() []string

	// TableImported returns the valid database tables for the Metric's imported statistics.
	TableImported() []string

	// JoinTable returns the secondary tables to query if the Metric cannot be calculated from the primary table.
	JoinTable() string

	// Column returns the database column name for the Metric.
	Column() string

	// ColumnImported returns the database column name for the Metric's imported statistics.
	ColumnImported() string

	// Expression returns the SQL expression for aggregation for the given table and if a subquery is required.
	// The subquery filters for the site_id and period. Returning true indicates that this Metric requires a subquery.
	Expression(string) (string, bool)

	// ExpressionImported returns the SQL expression for aggregation for the imported statistics table.
	ExpressionImported() string

	// ScanType returns a pointer to the type the value for this Metric scans into.
	ScanType() any

	// Zero returns the zero value for this Metric.
	Zero() any

	// String implements the fmt.Stringer interface.
	String() string
}
