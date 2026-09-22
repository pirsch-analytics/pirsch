package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMTerm is a Dimension.
type UTMTerm struct{}

// Table implements the Dimension interface.
func (d UTMTerm) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d UTMTerm) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d UTMTerm) Column(_ string) string {
	return "utm_term"
}

// ColumnImported implements the Dimension interface.
func (d UTMTerm) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d UTMTerm) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d UTMTerm) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d UTMTerm) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d UTMTerm) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d UTMTerm) String() string {
	return "utm_term"
}
