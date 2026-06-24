package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMContent is a Dimension.
type UTMContent struct{}

// Table implements the Dimension interface.
func (d UTMContent) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d UTMContent) Column(_ string) string {
	return "utm_content"
}

// Expression implements the Dimension interface.
func (d UTMContent) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d UTMContent) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d UTMContent) ScanType() any {
	return new(string)
}
