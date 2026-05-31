package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Entries is a Metric.
type Entries struct{}

// Table implements the Metric interface.
func (m Entries) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Metric interface.
func (m Entries) Column() string {
	return "entries"
}

// Expression implements the Metric interface.
func (m Entries) Expression(_ string) string {
	return "sum(sign)"
}

// ScanType implements the Metric interface.
func (m Entries) ScanType() any {
	return new(int64)
}
