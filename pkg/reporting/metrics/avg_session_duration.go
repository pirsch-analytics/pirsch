package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// AvgSessionDuration is a Metric.
type AvgSessionDuration struct{}

// Table implements the Metric interface.
func (m AvgSessionDuration) Table() []string {
	return []string{pkg.TablePageViews}
}

// JoinTable implements the Metric interface.
func (m AvgSessionDuration) JoinTable() string {
	return pkg.TablePageViews
}

// Column implements the Metric interface.
func (m AvgSessionDuration) Column() string {
	return "avg_session_duration"
}

// Expression implements the Metric interface.
func (m AvgSessionDuration) Expression(_ string) (string, bool) {
	return "avg(duration_seconds)", false
}

// ScanType implements the Metric interface.
func (m AvgSessionDuration) ScanType() any {
	return new(float64)
}

// Zero implements the Metric interface.
func (m AvgSessionDuration) Zero() any {
	return float64(0)
}
