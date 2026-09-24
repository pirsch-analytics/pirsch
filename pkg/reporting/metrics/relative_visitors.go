package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// RelativeVisitors is a Metric.
type RelativeVisitors struct{}

// Table implements the Metric interface.
func (m RelativeVisitors) Table() []string {
	return []string{pkg.TableSessions, pkg.TableSessions, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m RelativeVisitors) TableImported() []string {
	return pkg.ImportedTables
}

// JoinTable implements the Metric interface.
func (m RelativeVisitors) JoinTable() string {
	return ""
}

// Column implements the Metric interface.
func (m RelativeVisitors) Column() string {
	return "relative_visitors"
}

// ColumnImported implements the Metric interface.
func (m RelativeVisitors) ColumnImported() string {
	return "relative_visitors"
}

// Expression implements the Metric interface.
func (m RelativeVisitors) Expression(_ string) (string, bool) {
	return `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM "session_v7" %s), 1))`, true
}

// ExpressionImported implements the Metric interface.
func (m RelativeVisitors) ExpressionImported() string {
	return "toFloat64OrDefault(visitors / greatest(sum(visitors), 1))"
}

// ScanType implements the Metric interface.
func (m RelativeVisitors) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m RelativeVisitors) Zero() any {
	return float64(0)
}

// String implements the Metric interface.
func (m RelativeVisitors) String() string {
	return "relative_visitors"
}
