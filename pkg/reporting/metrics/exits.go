package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Exits is a Metric.
type Exits struct{}

// Table implements the Metric interface.
func (m Exits) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m Exits) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m Exits) Column() string {
	return "exits"
}

// Expression implements the Metric interface.
func (m Exits) Expression(_ string) (string, bool) {
	return "uniq(visitor_id, session_id)", false
}

// ScanType implements the Metric interface.
func (m Exits) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Exits) Zero() any {
	return uint64(0)
}
