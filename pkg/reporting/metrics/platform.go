package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Platform is a Metic.
type Platform struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Platform) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Platform) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Platform) Column() string {
	return "platform"
}

// Expression implements the Metic interface.
func (m Platform) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(platform)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Platform) ScanType() any {
	return new(int8)
}

// Zero implements the Metric interface.
func (m Platform) Zero() any {
	return int8(0)
}
