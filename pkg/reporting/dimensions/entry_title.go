package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// EntryTitle is a Dimension.
type EntryTitle struct{}

// Table implements the Dimension interface.
func (d EntryTitle) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Dimension interface.
func (d EntryTitle) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d EntryTitle) Column(_ string) string {
	return "entry_title"
}

// ColumnImported implements the Dimension interface.
func (d EntryTitle) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d EntryTitle) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d EntryTitle) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d EntryTitle) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d EntryTitle) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d EntryTitle) String() string {
	return "event_title"
}
