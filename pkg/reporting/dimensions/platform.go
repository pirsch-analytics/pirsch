package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Platform is a Dimension.
type Platform struct{}

// Table implements the Dimension interface.
func (d Platform) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Platform) Column(_ string) string {
	return "platform"
}

// Expression implements the Dimension interface.
func (d Platform) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Platform) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Platform) ScanType() any {
	return new(int8)
}
