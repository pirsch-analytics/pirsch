package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMContent is a Metic.
type UTMContent struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m UTMContent) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m UTMContent) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m UTMContent) Column() string {
	return "utm_content"
}

// Expression implements the Metic interface.
func (m UTMContent) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(utm_content)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m UTMContent) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m UTMContent) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m UTMContent) String() string {
	return "utm_content"
}
