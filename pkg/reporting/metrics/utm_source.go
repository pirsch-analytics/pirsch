package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMSource is a Metic.
type UTMSource struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m UTMSource) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m UTMSource) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m UTMSource) Column() string {
	return "utm_source"
}

// Expression implements the Metic interface.
func (m UTMSource) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(utm_source)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m UTMSource) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m UTMSource) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m UTMSource) String() string {
	return "utm_source"
}
