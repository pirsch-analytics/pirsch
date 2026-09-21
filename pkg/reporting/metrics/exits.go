package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Exits is a Metric.
type Exits struct{}

// Table implements the Metric interface.
func (m Exits) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m Exits) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Exits) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m Exits) Column() string {
	return "exits"
}

// ColumnImported implements the Metric interface.
func (m Exits) ColumnImported() string {
	return ""
}

// Expression implements the Metric interface.
func (m Exits) Expression(_ string) (string, bool) {
	return "uniq(visitor_id, session_id)", false
}

// ExpressionImported implements the Metric interface.
func (m Exits) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Exits) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Exits) Zero() any {
	return uint64(0)
}

// String implements the Metric interface.
func (m Exits) String() string {
	return "exits"
}
