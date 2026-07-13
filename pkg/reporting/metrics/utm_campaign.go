package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMCampaign is a Metic.
type UTMCampaign struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m UTMCampaign) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m UTMCampaign) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m UTMCampaign) Column() string {
	return "utm_campaign"
}

// Expression implements the Metic interface.
func (m UTMCampaign) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(utm_campaign)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m UTMCampaign) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m UTMCampaign) Zero() any {
	return ""
}
