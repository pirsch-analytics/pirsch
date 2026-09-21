package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// SessionID is a Dimension.
type SessionID struct{}

// Table implements the Dimension interface.
func (d SessionID) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d SessionID) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d SessionID) Column(_ string) string {
	return "session_id"
}

// ColumnImported implements the Dimension interface.
func (d SessionID) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d SessionID) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d SessionID) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d SessionID) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d SessionID) ScanType() any {
	return new(uint32)
}

// String implements the Dimension interface.
func (d SessionID) String() string {
	return "session_id"
}
