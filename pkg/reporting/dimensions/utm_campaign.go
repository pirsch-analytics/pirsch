package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMCampaign is a Dimension.
type UTMCampaign struct{}

// Table implements the Dimension interface.
func (d UTMCampaign) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d UTMCampaign) Column(_ string) string {
	return "utm_campaign"
}

// Expression implements the Dimension interface.
func (d UTMCampaign) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d UTMCampaign) ScanType() any {
	return new(string)
}
