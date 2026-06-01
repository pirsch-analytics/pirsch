package dimensions

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Year is a Dimension.
type Year struct{}

// Table implements the Dimension interface.
func (d Year) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Year) Column() string {
	return "year"
}

// Expression implements the Dimension interface.
func (d Year) Expression() string {
	return `toYear("time")`
}

// ScanType implements the Metric interface.
func (d Year) ScanType() any {
	return new(time.Time)
}
