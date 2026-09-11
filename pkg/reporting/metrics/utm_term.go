package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// UTMTerm is a Metic.
type UTMTerm struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m UTMTerm) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m UTMTerm) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m UTMTerm) Column() string {
	return "utm_term"
}

// Expression implements the Metic interface.
func (m UTMTerm) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(utm_term)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m UTMTerm) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m UTMTerm) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m UTMTerm) String() string {
	return "utm_term"
}
