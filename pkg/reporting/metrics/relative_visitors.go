package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// RelativeVisitors is a Metric.
type RelativeVisitors struct{}

// Table implements the Metric interface.
func (m RelativeVisitors) Table() []string {
	return []string{pkg.TableSessions, pkg.TableSessions, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m RelativeVisitors) JoinTable() string {
	return ""
}

// Column implements the Metric interface.
func (m RelativeVisitors) Column() string {
	return "relative_visitors"
}

// Expression implements the Metric interface.
func (m RelativeVisitors) Expression(_ string) (string, bool) {
	return `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM "session_v7" %s), 1))`, true
}

// ScanType implements the Metric interface.
func (m RelativeVisitors) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m RelativeVisitors) Zero() any {
	return float64(0)
}
