package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ScreenClass is a Dimension.
type ScreenClass struct{}

// Table implements the Dimension interface.
func (d ScreenClass) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d ScreenClass) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d ScreenClass) Column(_ string) string {
	return "screen_class"
}

// ColumnImported implements the Dimension interface.
func (d ScreenClass) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d ScreenClass) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d ScreenClass) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d ScreenClass) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d ScreenClass) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d ScreenClass) String() string {
	return "screen_class"
}
