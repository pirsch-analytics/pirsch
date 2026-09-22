package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMMedium is a Dimension.
type UTMMedium struct{}

// Table implements the Dimension interface.
func (d UTMMedium) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d UTMMedium) TableImported() []string {
	return []string{pkg.TableImportedUTMMedium}
}

// Column implements the Dimension interface.
func (d UTMMedium) Column(_ string) string {
	return "utm_medium"
}

// ColumnImported implements the Dimension interface.
func (d UTMMedium) ColumnImported() string {
	return "utm_medium"
}

// Expression implements the Dimension interface.
func (d UTMMedium) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d UTMMedium) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d UTMMedium) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d UTMMedium) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d UTMMedium) String() string {
	return "utm_medium"
}
