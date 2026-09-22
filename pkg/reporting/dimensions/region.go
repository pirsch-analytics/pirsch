package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Region is a Dimension.
type Region struct {
	// Any specifies whether the result set should not be grouped by the region.
	Any bool
}

// Table implements the Dimension interface.
func (d Region) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Region) TableImported() []string {
	return []string{pkg.TableImportedRegion}
}

// Column implements the Dimension interface.
func (d Region) Column(_ string) string {
	return "region"
}

// ColumnImported implements the Dimension interface.
func (d Region) ColumnImported() string {
	return "region"
}

// Expression implements the Dimension interface.
func (d Region) Expression(_ *DimensionExpressionOptions) string {
	if d.Any {
		return "any(region)"
	}

	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Region) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Region) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Region) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Region) String() string {
	return "region"
}
