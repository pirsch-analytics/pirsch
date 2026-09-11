package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// City is a Metic.
type City struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m City) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m City) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m City) Column() string {
	return "city"
}

// Expression implements the Metic interface.
func (m City) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(city)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m City) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m City) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m City) String() string {
	return "city"
}
