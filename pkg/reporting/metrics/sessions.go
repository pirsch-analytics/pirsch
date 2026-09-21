package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Sessions is a Metric.
type Sessions struct{}

// Table implements the Metric interface.
func (m Sessions) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m Sessions) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Sessions) JoinTable() string {
	return ""
}

// Column implements the Metric interface.
func (m Sessions) Column() string {
	return "sessions"
}

// ColumnImported implements the Metric interface.
func (m Sessions) ColumnImported() string {
	return ""
}

// Expression implements the Metric interface.
func (m Sessions) Expression(_ string) (string, bool) {
	return "uniq(session_id)", false
}

// ExpressionImported implements the Metric interface.
func (m Sessions) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Sessions) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Sessions) Zero() any {
	return uint64(0)
}

// String implements the Metric interface.
func (m Sessions) String() string {
	return "sessions"
}
