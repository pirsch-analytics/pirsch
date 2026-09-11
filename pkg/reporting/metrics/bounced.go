package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Bounced is a Metric.
type Bounced struct {
	// Min if set to true, returns the minimum value for this metric.
	Min bool
}

// Table implements the Metic interface.
func (m Bounced) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m Bounced) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m Bounced) Column() string {
	return "is_bounce"
}

// Expression implements the Metic interface.
func (m Bounced) Expression(_ string) (string, bool) {
	if m.Min {
		return "min(is_bounce)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Bounced) ScanType() any {
	return new(bool)
}

// Zero implements the Metric interface.
func (m Bounced) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m Bounced) String() string {
	return "bounced"
}
