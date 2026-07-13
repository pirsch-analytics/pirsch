package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Extended is a Metic.
type Extended struct {
	// Max if set to true, returns the maximum value for this metric.
	Max bool
}

// Table implements the Metic interface.
func (m Extended) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m Extended) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m Extended) Column() string {
	return "extended"
}

// Expression implements the Metic interface.
func (m Extended) Expression(_ string) (string, bool) {
	if m.Max {
		return "max(extended)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Extended) ScanType() any {
	return new(uint16)
}

// Zero implements the Metric interface.
func (m Extended) Zero() any {
	return uint16(0)
}
