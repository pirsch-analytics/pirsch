package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ExitPath is a Dimension.
type ExitPath struct{}

// Table implements the Dimension interface.
func (d ExitPath) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Dimension interface.
func (d ExitPath) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d ExitPath) Column(_ string) string {
	return "exit_path"
}

// ColumnImported implements the Dimension interface.
func (d ExitPath) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d ExitPath) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d ExitPath) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d ExitPath) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d ExitPath) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d ExitPath) String() string {
	return "exit_path"
}
