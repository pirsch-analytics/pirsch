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
func (d Language) Column() string {
	return "language"
}

// Expression implements the Dimension interface.
func (d Language) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d Language) ScanType() any {
	return new(string)
}
