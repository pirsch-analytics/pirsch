package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// City is a Dimension.
type City struct{}

// Table implements the Dimension interface.
func (d City) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d City) TableImported() []string {
	return []string{pkg.TableImportedCity}
}

// Column implements the Dimension interface.
func (d City) Column(_ string) string {
	return "city"
}

// ColumnImported implements the Dimension interface.
func (d City) ColumnImported() string {
	return "city"
}

// Expression implements the Dimension interface.
func (d City) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d City) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d City) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d City) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d City) String() string {
	return "city"
}
