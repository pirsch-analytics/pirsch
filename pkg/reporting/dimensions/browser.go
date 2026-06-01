package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Browser is a Dimension.
type Browser struct{}

// Table implements the Dimension interface.
func (d Browser) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Browser) Column() string {
	return "browser"
}

// Expression implements the Dimension interface.
func (d Browser) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d Browser) ScanType() any {
	return new(string)
}
