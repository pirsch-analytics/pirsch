package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// AvgSessionDuration is a Metric.
type AvgSessionDuration struct{}

// Table implements the Metric interface.
func (m AvgSessionDuration) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m AvgSessionDuration) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m AvgSessionDuration) Column() string {
	return "avg_session_duration"
}

// Expression implements the Metric interface.
func (m AvgSessionDuration) Expression(_ string) (string, bool) {
	return "avgOrDefaultIf(duration_seconds, duration_seconds > 0)", false
}

// ScanType implements the Metric interface.
func (m AvgSessionDuration) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m AvgSessionDuration) Zero() any {
	return float64(0)
}

// String implements the Metric interface.
func (m AvgSessionDuration) String() string {
	return "avg_session_duration"
}
