package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// AvgTimeOnPage is a Metric.
type AvgTimeOnPage struct{}

// Table implements the Metric interface.
func (m AvgTimeOnPage) Table() []string {
	// all tables, as this is joined via a special WITH subquery
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m AvgTimeOnPage) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m AvgTimeOnPage) JoinTable() string {
	return ""
}

// Column implements the Metric interface.
func (m AvgTimeOnPage) Column() string {
	return "avg_time_on_page"
}

// ColumnImported implements the Metric interface.
func (m AvgTimeOnPage) ColumnImported() string {
	return ""
}

// Expression implements the Metric interface.
func (m AvgTimeOnPage) Expression(_ string) (string, bool) {
	return "any(avg_time_on_page)", false
}

// ExpressionImported implements the Metric interface.
func (m AvgTimeOnPage) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m AvgTimeOnPage) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m AvgTimeOnPage) Zero() any {
	return float64(0)
}

// String implements the Metric interface.
func (m AvgTimeOnPage) String() string {
	return "avg_time_on_page"
}
