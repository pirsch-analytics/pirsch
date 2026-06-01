package dimensions

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Week is a Dimension.
type Week struct{}

// Table implements the Dimension interface.
func (d Week) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Week) Column() string {
	return "week"
}

// Expression implements the Dimension interface.
func (d Week) Expression() string {
	return `toWeek("time")`
}

// ScanType implements the Metric interface.
func (d Week) ScanType() any {
	return new(time.Time)
}
