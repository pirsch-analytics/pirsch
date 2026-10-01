package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// OSVersion is a Dimension.
type OSVersion struct{}

// Table implements the Dimension interface.
func (d OSVersion) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d OSVersion) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d OSVersion) Column(_ string) string {
	return "os_version"
}

// ColumnImported implements the Dimension interface.
func (d OSVersion) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d OSVersion) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d OSVersion) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d OSVersion) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d OSVersion) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d OSVersion) String() string {
	return "os_version"
}
