package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Duration is a Dimension.
type Duration struct{}

// Table implements the Dimension interface.
func (d Duration) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Dimension interface.
func (d Duration) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Duration) Column(_ string) string {
	return "duration_seconds"
}

// ColumnImported implements the Dimension interface.
func (d Duration) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Duration) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Duration) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Duration) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Duration) ScanType() any {
	return new(uint32)
}

// String implements the Dimension interface.
func (d Duration) String() string {
	return "duration_seconds"
}
