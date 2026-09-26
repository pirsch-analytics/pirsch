package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// BrowserVersion is a Dimension.
type BrowserVersion struct{}

// Table implements the Dimension interface.
func (d BrowserVersion) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d BrowserVersion) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d BrowserVersion) Column(_ string) string {
	return "browser_version"
}

// ColumnImported implements the Dimension interface.
func (d BrowserVersion) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d BrowserVersion) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d BrowserVersion) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d BrowserVersion) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d BrowserVersion) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d BrowserVersion) String() string {
	return "browser_version"
}
