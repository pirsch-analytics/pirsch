package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Hostname is a Dimension.
type Hostname struct{}

// Table implements the Dimension interface.
func (d Hostname) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Hostname) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Hostname) Column(_ string) string {
	return "hostname"
}

// ColumnImported implements the Dimension interface.
func (d Hostname) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Hostname) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Hostname) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d Hostname) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Hostname) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Hostname) String() string {
	return "hostname"
}
