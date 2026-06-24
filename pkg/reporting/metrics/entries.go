package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Entries is a Metric.
type Entries struct{}

// Table implements the Metric interface.
func (m Entries) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m Entries) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m Entries) Column() string {
	return "entries"
}

// Expression implements the Metric interface.
func (m Entries) Expression(_ string) (string, bool) {
	return "uniq(visitor_id, session_id)", false
}

// ScanType implements the Metric interface.
func (m Entries) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Entries) Zero() any {
	return uint64(0)
}
