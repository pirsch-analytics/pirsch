package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// EntryRate is a Metric.
type EntryRate struct{}

// Table implements the Metric interface.
func (m EntryRate) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m EntryRate) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m EntryRate) Column() string {
	return "entry_rate"
}

// Expression implements the Metric interface.
func (m EntryRate) Expression(_ string) (string, bool) {
	return `toFloat64OrDefault(entries / greatest((SELECT uniq(visitor_id, session_id) FROM "session_v7" %s), 1))`, true
}

// ScanType implements the Metric interface.
func (m EntryRate) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m EntryRate) Zero() any {
	return float64(0)
}
