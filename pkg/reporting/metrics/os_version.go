package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// OSVersion is a Metic.
type OSVersion struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m OSVersion) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m OSVersion) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m OSVersion) Column() string {
	return "os_version"
}

// Expression implements the Metic interface.
func (m OSVersion) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(os_version)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m OSVersion) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m OSVersion) Zero() any {
	return ""
}
