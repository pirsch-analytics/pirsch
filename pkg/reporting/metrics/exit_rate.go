package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// ExitRate is a Metric.
type ExitRate struct{}

// Table implements the Metric interface.
func (m ExitRate) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m ExitRate) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m ExitRate) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m ExitRate) Column() string {
	return "exit_rate"
}

// ColumnImported implements the Metric interface.
func (m ExitRate) ColumnImported() string {
	return ""
}

// Expression implements the Metric interface.
func (m ExitRate) Expression(_ string) (string, bool) {
	return `toFloat64OrDefault(exits / greatest((SELECT uniq(visitor_id, session_id) FROM "session_v7" %s), 1))`, true
}

// ExpressionImported implements the Metric interface.
func (m ExitRate) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m ExitRate) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m ExitRate) Zero() any {
	return float64(0)
}

// String implements the Metric interface.
func (m ExitRate) String() string {
	return "exit_rate"
}
