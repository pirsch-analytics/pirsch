package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// EntryPath is a Dimension.
type EntryPath struct{}

// Table implements the Dimension interface.
func (d EntryPath) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Dimension interface.
func (d EntryPath) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d EntryPath) Column(_ string) string {
	return "entry_path"
}

// ColumnImported implements the Dimension interface.
func (d EntryPath) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d EntryPath) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d EntryPath) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d EntryPath) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d EntryPath) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d EntryPath) String() string {
	return "entry_path"
}
