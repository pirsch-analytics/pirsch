package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Events is a Metric.
type Events struct{}

// Table implements the Metric interface.
func (m Events) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Metric interface.
func (m Events) Column() string {
	return "events"
}

// Expression implements the Metric interface.
func (m Events) Expression(_ string) (string, bool) {
	return "count(*)", false
}

// ScanType implements the Metric interface.
func (m Events) ScanType() any {
	return new(uint64)
}
