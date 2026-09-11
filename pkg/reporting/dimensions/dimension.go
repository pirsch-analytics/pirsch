package dimensions

import (
	"encoding/json/v2"
	"time"
)

// Dimension is a field results are grouped by.
type Dimension interface {
	// Table returns the valid database tables for the Dimension.
	Table() []string

	// Column returns the database column name for the given table for the Dimension.
	// This also handles "joins" by potentially returning an entirely different column (like entry_path instead of path for bounces).
	Column(string) string

	// Expression returns the SQL aggregation expression for the given options.
	// If empty, the Column name will be used instead.
	Expression(*DimensionExpressionOptions) string

	// Args returns optional arguments for the Expression.
	Args() []any

	// ScanType returns a pointer to the type the value for this Dimension scans into.
	ScanType() any

	// String implements the fmt.Stringer interface.
	String() string
}

// DimensionExpressionOptions are the options for the Dimension.Expression.
type DimensionExpressionOptions struct {
	// Timezone is the timezone.
	Timezone *time.Location

	// WeekdayMode sets the start day of the week.
	WeekdayMode int
}

// String implements the fmt.Stringer interface.
func (o *DimensionExpressionOptions) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}
