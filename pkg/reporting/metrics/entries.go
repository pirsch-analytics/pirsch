package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Entries is a Metric.
type Entries struct{}

// Table implements the Metric interface.
func (m Entries) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m Entries) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Entries) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m Entries) Column() string {
	return "entries"
}

// ColumnImported implements the Metric interface.
func (m Entries) ColumnImported() string {
	return ""
}

// Expression implements the Metric interface.
func (m Entries) Expression(_ string) (string, bool) {
	return "uniq(visitor_id, session_id)", false
}

// ExpressionImported implements the Metric interface.
func (m Entries) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Entries) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Entries) Zero() any {
	return uint64(0)
}

// String implements the Metric interface.
func (m Entries) String() string {
	return "entries"
}
