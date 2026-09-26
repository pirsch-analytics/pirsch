package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// EntryRate is a Metric.
type EntryRate struct{}

// Table implements the Metric interface.
func (m EntryRate) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m EntryRate) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m EntryRate) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m EntryRate) Column() string {
	return "entry_rate"
}

// ColumnImported implements the Metric interface.
func (m EntryRate) ColumnImported() string {
	return ""
}

// Expression implements the Metric interface.
func (m EntryRate) Expression(_ string) (string, bool) {
	return `toFloat64OrDefault(entries / greatest((SELECT uniq(visitor_id, session_id) FROM "session_v7" %s), 1))`, true
}

// ExpressionImported implements the Metric interface.
func (m EntryRate) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m EntryRate) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m EntryRate) Zero() any {
	return float64(0)
}

// String implements the Metric interface.
func (m EntryRate) String() string {
	return "entry_rate"
}
