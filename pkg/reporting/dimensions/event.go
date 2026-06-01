package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Event is a Dimension.
type Event struct{}

// Table implements the Dimension interface.
func (d Event) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Event) Column(_ string) string {
	return "name"
}

// Expression implements the Dimension interface.
func (d Event) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d Event) ScanType() any {
	return new(string)
}
