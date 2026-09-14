package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// EventPath is a Dimension.
type EventPath struct{}

// Table implements the Dimension interface.
func (d EventPath) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d EventPath) Column(table string) string {
	// filter/join on entry path for metrics like the bounce rate
	if table == pkg.TableSessions {
		return "entry_path"
	}

	return "path"
}

// Expression implements the Dimension interface.
func (d EventPath) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d EventPath) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d EventPath) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d EventPath) String() string {
	return "event_path"
}
