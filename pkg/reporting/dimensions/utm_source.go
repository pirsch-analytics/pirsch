package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMSource is a Dimension.
type UTMSource struct{}

// Table implements the Dimension interface.
func (d UTMSource) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d UTMSource) TableImported() []string {
	return []string{pkg.TableImportedUTMSource}
}

// Column implements the Dimension interface.
func (d UTMSource) Column(_ string) string {
	return "utm_source"
}

// ColumnImported implements the Dimension interface.
func (d UTMSource) ColumnImported() string {
	return "utm_source"
}

// Expression implements the Dimension interface.
func (d UTMSource) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d UTMSource) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d UTMSource) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d UTMSource) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d UTMSource) String() string {
	return "utm_source"
}
