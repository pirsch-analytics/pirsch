package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Region is a Metic.
type Region struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Region) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Region) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Region) Column() string {
	return "region"
}

// Expression implements the Metic interface.
func (m Region) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(region)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Region) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Region) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m Region) String() string {
	return "region"
}
