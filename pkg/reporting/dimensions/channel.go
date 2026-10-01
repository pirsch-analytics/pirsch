package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Channel is a Dimension.
type Channel struct{}

// Table implements the Dimension interface.
func (d Channel) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Channel) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Channel) Column(_ string) string {
	return "channel"
}

// ColumnImported implements the Dimension interface.
func (d Channel) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Channel) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Channel) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Channel) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Channel) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Channel) String() string {
	return "channel"
}
