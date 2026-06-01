package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMTerm is a Dimension.
type UTMTerm struct{}

// Table implements the Dimension interface.
func (d UTMTerm) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d UTMTerm) Column() string {
	return "utm_term"
}

// Expression implements the Dimension interface.
func (d UTMTerm) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d UTMTerm) ScanType() any {
	return new(string)
}
