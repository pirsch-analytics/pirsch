package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Hostname is a Metic.
type Hostname struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Hostname) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Hostname) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Hostname) Column() string {
	return "hostname"
}

// Expression implements the Metic interface.
func (m Hostname) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(hostname)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Hostname) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Hostname) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m Hostname) String() string {
	return "hostname"
}
