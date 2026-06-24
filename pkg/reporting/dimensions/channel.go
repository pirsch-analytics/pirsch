package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Channel is a Dimension.
type Channel struct{}

// Table implements the Dimension interface.
func (d Channel) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Channel) Column(_ string) string {
	return "channel"
}

// Expression implements the Dimension interface.
func (d Channel) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Channel) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Channel) ScanType() any {
	return new(string)
}
