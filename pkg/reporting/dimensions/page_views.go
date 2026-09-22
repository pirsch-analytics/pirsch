package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// PageViews is a Dimension.
type PageViews struct{}

// Table implements the Dimension interface.
func (d PageViews) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Dimension interface.
func (d PageViews) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d PageViews) Column(_ string) string {
	return "page_views"
}

// ColumnImported implements the Dimension interface.
func (d PageViews) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d PageViews) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d PageViews) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d PageViews) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d PageViews) ScanType() any {
	return new(uint16)
}

// String implements the Dimension interface.
func (d PageViews) String() string {
	return "page_views"
}
