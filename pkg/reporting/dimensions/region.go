package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Region is a Dimension.
type Region struct{}

// Table implements the Dimension interface.
func (d Region) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Region) Column(_ string) string {
	return "region"
}

// Expression implements the Dimension interface.
func (d Region) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d Region) ScanType() any {
	return new(string)
}
