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

// Column implements the Dimension interface.
func (d City) Column(_ string) string {
	return "city"
}

// Expression implements the Dimension interface.
func (d City) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d City) ScanType() any {
	return new(string)
}
