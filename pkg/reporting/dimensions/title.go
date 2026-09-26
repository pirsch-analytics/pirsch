package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Title is a Dimension.
type Title struct{}

// Table implements the Dimension interface.
func (d Title) Table() []string {
	return []string{pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Title) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Title) Column(_ string) string {
	return "title"
}

// ColumnImported implements the Dimension interface.
func (d Title) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Title) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Title) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Title) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Title) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Title) String() string {
	return "title"
}
