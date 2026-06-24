package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// OS is a Dimension.
type OS struct{}

// Table implements the Dimension interface.
func (d OS) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d OS) Column(_ string) string {
	return "os"
}

// Expression implements the Dimension interface.
func (d OS) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d OS) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d OS) ScanType() any {
	return new(string)
}
