package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// VisitorID is a Dimension.
type VisitorID struct{}

// Table implements the Dimension interface.
func (d VisitorID) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d VisitorID) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d VisitorID) Column(_ string) string {
	return "visitor_id"
}

// ColumnImported implements the Dimension interface.
func (d VisitorID) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d VisitorID) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d VisitorID) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d VisitorID) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d VisitorID) ScanType() any {
	return new(uint64)
}

// String implements the Dimension interface.
func (d VisitorID) String() string {
	return "visitor_id"
}
