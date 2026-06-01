package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// RelativeViews is a Metric.
type RelativeViews struct{}

// Table implements the Metric interface.
func (m RelativeViews) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Metric interface.
func (m RelativeViews) Column() string {
	return "relative_views"
}

// Expression implements the Metric interface.
func (m RelativeViews) Expression(_ string) (string, bool) {
	return `toFloat64OrDefault(page_views / greatest((SELECT sum(page_views * sign) FROM "session_v7" %s), 1))`, true
}

// ScanType implements the Metric interface.
func (m RelativeViews) ScanType() any {
	return new(float64)
}
