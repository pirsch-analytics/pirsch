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

// Column implements the Dimension interface.
func (d OSVersion) Column() string {
	return "os_version"
}

// Expression implements the Dimension interface.
func (d OSVersion) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d OSVersion) ScanType() any {
	return new(string)
}
