package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Language is a Dimension.
type Language struct{}

// Table implements the Dimension interface.
func (d Language) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Language) Column(_ string) string {
	return "language"
}

// Expression implements the Dimension interface.
func (d Language) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Language) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Language) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Language) String() string {
	return "language"
}
