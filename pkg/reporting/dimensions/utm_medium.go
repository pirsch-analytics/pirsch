package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMMedium is a Dimension.
type UTMMedium struct{}

// Table implements the Dimension interface.
func (d UTMMedium) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d UTMMedium) Column(_ string) string {
	return "utm_medium"
}

// Expression implements the Dimension interface.
func (d UTMMedium) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d UTMMedium) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d UTMMedium) ScanType() any {
	return new(string)
}
