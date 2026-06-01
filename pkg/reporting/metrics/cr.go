package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// CR is a Metric.
type CR struct{}

// Table implements the Metric interface.
func (m CR) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m CR) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m CR) Column() string {
	return "cr"
}

// Expression implements the Metric interface.
func (m CR) Expression(_ string) (string, bool) {
	return `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM "session_v7" %s), 1))`, true
}

// ScanType implements the Metric interface.
func (m CR) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m CR) Zero() any {
	return float64(0)
}
