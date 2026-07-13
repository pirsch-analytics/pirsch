package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// OS is a Metic.
type OS struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m OS) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m OS) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m OS) Column() string {
	return "os"
}

// Expression implements the Metic interface.
func (m OS) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(os)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m OS) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m OS) Zero() any {
	return ""
}
