package metrics

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Start is a Metic.
type Start struct {
	// Min if set to true, returns the minimum value for this metric.
	Min bool
}

// Table implements the Metic interface.
func (m Start) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m Start) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m Start) Column() string {
	return "start"
}

// Expression implements the Metic interface.
func (m Start) Expression(_ string) (string, bool) {
	if m.Min {
		return "min(start)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Start) ScanType() any {
	return new(time.Time)
}

// Zero implements the Metric interface.
func (m Start) Zero() any {
	return time.Time{}
}

// String implements the Metric interface.
func (m Start) String() string {
	return "start"
}
