package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ScreenClass is a Dimension.
type ScreenClass struct{}

// Table implements the Dimension interface.
func (d ScreenClass) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d ScreenClass) Column(_ string) string {
	return "screen_class"
}

// Expression implements the Dimension interface.
func (d ScreenClass) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d ScreenClass) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d ScreenClass) ScanType() any {
	return new(string)
}
