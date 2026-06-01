package dimensions

// Dimension is a field results are grouped by.
type Dimension interface {
	// Table returns the valid database tables for the Dimension.
	Table() []string

	// Column returns the database column name for the Dimension.
	Column() string

	// Expression returns the SQL expression for aggregation.
	// If empty, the Column name will be used instead.
	Expression() string

	// ScanType returns a pointer to the type the value for this Dimension scans into.
	ScanType() any
}

// TODO
/*
	CustomMetricKey string
	CustomMetricType pkg.CustomMetricType
	VisitorID uint64
	SessionID uint32
	Platform string
	Tags map[string]string
	Tag []string
	EventMetaKey []string
	EventMeta map[string]string
*/
