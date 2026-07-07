package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Weekday is a Dimension.
type Weekday struct{}

// Table implements the Dimension interface.
func (d Weekday) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Weekday) Column(_ string) string {
	return "weekday"
}

// Expression implements the Dimension interface.
func (d Weekday) Expression() string {
	return `toDayOfWeek("time")`
}

// Args implements the Dimension interface.
func (d Weekday) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Weekday) ScanType() any {
	return new(uint8)
}
