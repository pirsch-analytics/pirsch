package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Exits is a Metric.
type Exits struct{}

// Table implements the Metric interface.
func (m Exits) Table() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m Exits) Column() string {
	return "exits"
}

// Expression implements the Metric interface.
func (m Exits) Expression(_ string) string {
	return "sum(sign)"
}
