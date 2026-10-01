package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Platform is a Dimension.
type Platform struct{}

// Table implements the Dimension interface.
func (d Platform) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Platform) TableImported() []string {
	return []string{pkg.TableImportedDevice}
}

// Column implements the Dimension interface.
func (d Platform) Column(_ string) string {
	return "platform"
}

// ColumnImported implements the Dimension interface.
func (d Platform) ColumnImported() string {
	return "platform"
}

// Expression implements the Dimension interface.
func (d Platform) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Platform) ExpressionImported(_ *DimensionExpressionOptions) string {
	return "category"
}

// Args implements the Dimension interface.
func (d Platform) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Platform) ScanType() any {
	return new(int8)
}

// String implements the Dimension interface.
func (d Platform) String() string {
	return "platform"
}
