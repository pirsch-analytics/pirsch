package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ReferrerName is a Dimension.
type ReferrerName struct{}

// Table implements the Dimension interface.
func (d ReferrerName) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d ReferrerName) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d ReferrerName) Column(_ string) string {
	return "referrer_name"
}

// ColumnImported implements the Dimension interface.
func (d ReferrerName) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d ReferrerName) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d ReferrerName) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d ReferrerName) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d ReferrerName) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d ReferrerName) String() string {
	return "referrer_name"
}
