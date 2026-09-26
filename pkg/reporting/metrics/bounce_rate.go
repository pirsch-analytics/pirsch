package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// BounceRate is a Metric.
type BounceRate struct{}

// Table implements the Metric interface.
func (m BounceRate) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m BounceRate) TableImported() []string {
	return []string{pkg.TableImportedPage, pkg.TableImportedReferrer, pkg.TableImportedVisitors}
}

// JoinTable implements the Metric interface.
func (m BounceRate) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m BounceRate) Column() string {
	return "bounce_rate"
}

// ColumnImported implements the Metric interface.
func (m BounceRate) ColumnImported() string {
	return "bounce_rate"
}

// Expression implements the Metric interface.
func (m BounceRate) Expression(_ string) (string, bool) {
	return "toFloat64OrDefault(bounces / greatest(uniq(visitor_id, session_id), 1))", false
}

// ExpressionImported implements the Metric interface.
func (m BounceRate) ExpressionImported() string {
	return "toFloat64OrDefault(bounces / greatest(sessions, 1))"
}

// ScanType implements the Metric interface.
func (m BounceRate) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m BounceRate) Zero() any {
	return float64(0)
}

// String implements the Metric interface.
func (m BounceRate) String() string {
	return "bounce_rate"
}
