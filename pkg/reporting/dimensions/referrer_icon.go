package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ReferrerIcon is a Dimension.
type ReferrerIcon struct{}

// Table implements the Dimension interface.
func (d ReferrerIcon) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d ReferrerIcon) Column(_ string) string {
	return "referrer_icon"
}

// Expression implements the Dimension interface.
func (d ReferrerIcon) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d ReferrerIcon) ScanType() any {
	return new(string)
}
