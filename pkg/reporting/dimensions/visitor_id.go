package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// VisitorID is a Dimension.
type VisitorID struct{}

// Table implements the Dimension interface.
func (d VisitorID) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d VisitorID) Column(_ string) string {
	return "visitor_id"
}

// Expression implements the Dimension interface.
func (d VisitorID) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d VisitorID) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d VisitorID) ScanType() any {
	return new(uint64)
}
