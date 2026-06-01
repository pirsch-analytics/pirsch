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

// Column implements the Dimension interface.
func (d UTMSource) Column() string {
	return "utm_source"
}

// Expression implements the Dimension interface.
func (d UTMSource) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d UTMSource) ScanType() any {
	return new(string)
}
