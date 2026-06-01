package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Sessions is a Metric.
type Sessions struct{}

// Table implements the Metric interface.
func (m Sessions) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Metric interface.
func (m Sessions) Column() string {
	return "sessions"
}

// Expression implements the Metric interface.
func (m Sessions) Expression(_ string) (string, bool) {
	return "uniq(session_id)", false
}

// ScanType implements the Metric interface.
func (m Sessions) ScanType() any {
	return new(uint64)
}
