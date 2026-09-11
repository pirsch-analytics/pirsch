package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Referrer is a Metic.
type Referrer struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Referrer) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Referrer) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Referrer) Column() string {
	return "referrer"
}

// Expression implements the Metic interface.
func (m Referrer) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(referrer)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Referrer) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Referrer) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m Referrer) String() string {
	return "referrer"
}
