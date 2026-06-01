package dimensions

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Hour is a Dimension.
type Hour struct{}

// Table implements the Dimension interface.
func (d Hour) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Hour) Column() string {
	return "hour"
}

// Expression implements the Dimension interface.
func (d Hour) Expression() string {
	return `toHour("time")`
}

// ScanType implements the Metric interface.
func (d Hour) ScanType() any {
	return new(time.Time)
}
