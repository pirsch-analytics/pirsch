package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMMedium is a Metic.
type UTMMedium struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m UTMMedium) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m UTMMedium) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m UTMMedium) Column() string {
	return "utm_medium"
}

// Expression implements the Metic interface.
func (m UTMMedium) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(utm_medium)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m UTMMedium) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m UTMMedium) Zero() any {
	return ""
}
