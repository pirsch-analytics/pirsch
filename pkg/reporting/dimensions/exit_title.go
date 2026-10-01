package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ExitTitle is a Dimension.
type ExitTitle struct{}

// Table implements the Dimension interface.
func (d ExitTitle) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Dimension interface.
func (d ExitTitle) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d ExitTitle) Column(_ string) string {
	return "exit_title"
}

// ColumnImported implements the Dimension interface.
func (d ExitTitle) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d ExitTitle) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d ExitTitle) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d ExitTitle) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d ExitTitle) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d ExitTitle) String() string {
	return "exit_title"
}
