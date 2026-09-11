package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ReferrerName is a Metic.
type ReferrerName struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m ReferrerName) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m ReferrerName) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m ReferrerName) Column() string {
	return "referrer_name"
}

// Expression implements the Metic interface.
func (m ReferrerName) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(referrer_name)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m ReferrerName) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m ReferrerName) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m ReferrerName) String() string {
	return "referrer_name"
}
