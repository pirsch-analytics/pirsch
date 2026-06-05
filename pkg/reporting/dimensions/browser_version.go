package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// BrowserVersion is a Dimension.
type BrowserVersion struct{}

// Table implements the Dimension interface.
func (d BrowserVersion) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d BrowserVersion) Column(_ string) string {
	return "browser_version"
}

// Expression implements the Dimension interface.
func (d BrowserVersion) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d BrowserVersion) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d BrowserVersion) ScanType() any {
	return new(string)
}
