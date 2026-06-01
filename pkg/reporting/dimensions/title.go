package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Title is a Dimension.
type Title struct{}

// Table implements the Dimension interface.
func (d Title) Table() []string {
	return []string{pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Title) Column(_ string) string {
	return "title"
}

// Expression implements the Dimension interface.
func (d Title) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d Title) ScanType() any {
	return new(string)
}
