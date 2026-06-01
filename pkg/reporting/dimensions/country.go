package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Country is a Dimension.
type Country struct{}

// Table implements the Dimension interface.
func (d Country) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Country) Column(_ string) string {
	return "country_code"
}

// Expression implements the Dimension interface.
func (d Country) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d Country) ScanType() any {
	return new(string)
}
