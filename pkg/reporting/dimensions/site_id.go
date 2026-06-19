package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// SiteID is a Dimension.
type SiteID struct{}

// Table implements the Dimension interface.
func (d SiteID) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d SiteID) Column(_ string) string {
	return "site_id"
}

// Expression implements the Dimension interface.
func (d SiteID) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d SiteID) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d SiteID) ScanType() any {
	return new(uint64)
}
