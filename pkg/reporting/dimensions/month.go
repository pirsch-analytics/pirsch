package dimensions

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Month is a Dimension.
type Month struct{}

// Table implements the Dimension interface.
func (d Month) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Month) Column(_ string) string {
	return "month"
}

// Expression implements the Dimension interface.
func (d Month) Expression() string {
	return `toMonth("time")`
}

// ScanType implements the Metric interface.
func (d Month) ScanType() any {
	return new(time.Time)
}
