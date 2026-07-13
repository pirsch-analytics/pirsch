package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Country is a Metic.
type Country struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Country) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Country) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Country) Column() string {
	return "country_code"
}

// Expression implements the Metic interface.
func (m Country) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(country_code)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Country) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Country) Zero() any {
	return ""
}
