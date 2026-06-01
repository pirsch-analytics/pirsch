package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ReferrerName is a Dimension.
type ReferrerName struct{}

// Table implements the Dimension interface.
func (d ReferrerName) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d ReferrerName) Column(_ string) string {
	return "referrer_name"
}

// Expression implements the Dimension interface.
func (d ReferrerName) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d ReferrerName) ScanType() any {
	return new(string)
}
