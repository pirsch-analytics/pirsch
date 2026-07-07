package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Minute is a Dimension.
type Minute struct{}

// Table implements the Dimension interface.
func (d Minute) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Minute) Column(_ string) string {
	return "minute"
}

// Expression implements the Dimension interface.
func (d Minute) Expression() string {
	return `toMinute("time")`
}

// Args implements the Dimension interface.
func (d Minute) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Minute) ScanType() any {
	return new(uint8)
}
