package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Visitors is a Metric.
type Visitors struct{}

// Table implements the Metric interface.
func (m Visitors) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Visitors) JoinTable() string {
	return ""
}

// Column implements the Metric interface.
func (m Visitors) Column() string {
	return "visitors"
}

// Expression implements the Metric interface.
func (m Visitors) Expression(_ string) (string, bool) {
	return "uniq(visitor_id)", false
}

// ScanType implements the Metric interface.
func (m Visitors) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Visitors) Zero() any {
	return uint64(0)
}
