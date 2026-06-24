package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Hostname is a Dimension.
type Hostname struct{}

// Table implements the Dimension interface.
func (d Hostname) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Hostname) Column(_ string) string {
	return "hostname"
}

// Expression implements the Dimension interface.
func (d Hostname) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Hostname) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Hostname) ScanType() any {
	return new(string)
}
