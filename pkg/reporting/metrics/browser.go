package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Browser is a Metic.
type Browser struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Browser) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Browser) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Browser) Column() string {
	return "browser"
}

// Expression implements the Metic interface.
func (m Browser) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(browser)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Browser) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Browser) Zero() any {
	return ""
}
