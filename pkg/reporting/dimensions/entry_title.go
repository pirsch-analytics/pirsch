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

// Column implements the Dimension interface.
func (d EntryTitle) Column(_ string) string {
	return "entry_title"
}

// Expression implements the Dimension interface.
func (d EntryTitle) Expression(_ *DimensionExpressionOptions) string {
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
