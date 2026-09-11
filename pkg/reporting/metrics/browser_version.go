package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// BrowserVersion is a Metic.
type BrowserVersion struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m BrowserVersion) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m BrowserVersion) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m BrowserVersion) Column() string {
	return "browser_version"
}

// Expression implements the Metic interface.
func (m BrowserVersion) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(browser_version)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m BrowserVersion) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m BrowserVersion) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m BrowserVersion) String() string {
	return "browser_version"
}
