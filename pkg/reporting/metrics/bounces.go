package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Bounces is a Metric.
type Bounces struct{}

// Table implements the Metric interface.
func (m Bounces) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m Bounces) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m Bounces) Column() string {
	return "bounces"
}

// Expression implements the Metric interface.
func (m Bounces) Expression(_ string) (string, bool) {
	return "sum(is_bounce * sign)", false
}

// ScanType implements the Metric interface.
func (m Bounces) ScanType() any {
	return new(int64)
}

// Zero implements the Metric interface.
func (m Bounces) Zero() any {
	return int64(0)
}
