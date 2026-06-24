package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Referrer is a Dimension.
type Referrer struct{}

// Table implements the Dimension interface.
func (d Referrer) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Referrer) Column(_ string) string {
	return "referrer"
}

// Expression implements the Dimension interface.
func (d Referrer) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Referrer) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Referrer) ScanType() any {
	return new(string)
}
