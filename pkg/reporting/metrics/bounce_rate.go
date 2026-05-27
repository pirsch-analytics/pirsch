package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// BounceRate is a Metric.
type BounceRate struct{}

// Table implements the Metric interface.
func (m BounceRate) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Metric interface.
func (m BounceRate) Column() string {
	return "bounce_rate"
}

// Expression implements the Metric interface.
func (m BounceRate) Expression(_ string) string {
	// TODO filters missing
	return "sum(is_bounce*sign) / IF(uniq(visitor_id, session_id) = 0, 1, uniq(visitor_id, session_id))"
}
