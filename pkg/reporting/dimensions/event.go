package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

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
func (d Event) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Event) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Event) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Event) String() string {
	return "event"
}
