package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// AvgTimeOnPage is a Metric.
type AvgTimeOnPage struct{}

// Table implements the Metric interface.
func (m AvgTimeOnPage) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Metric interface.
func (m AvgTimeOnPage) Column() string {
	return "avg_time_on_page"
}

// Expression implements the Metric interface.
func (m AvgTimeOnPage) Expression(_ string) (string, bool) {
	// TODO
	return "", false
}

// ScanType implements the Metric interface.
func (m AvgTimeOnPage) ScanType() any {
	return new(float64)
}
