package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Events is a Metric.
type Events struct{}

// Table implements the Metric interface.
func (m Events) Table() []string {
	return []string{pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m Events) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Events) JoinTable() string {
	return pkg.TableEvents
}

// Column implements the Metric interface.
func (m Events) Column() string {
	return "events"
}

// ColumnImported implements the Metric interface.
func (m Events) ColumnImported() string {
	return ""
}

// Expression implements the Metric interface.
func (m Events) Expression(_ string) (string, bool) {
	return "count(*)", false
}

// ExpressionImported implements the Metric interface.
func (m Events) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Events) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Events) Zero() any {
	return uint64(0)
}

// String implements the Metric interface.
func (m Events) String() string {
	return "events"
}
