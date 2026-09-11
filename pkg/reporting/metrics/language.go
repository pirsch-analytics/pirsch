package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Language is a Metic.
type Language struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Language) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// JoinTable implements the Metric interface.
func (m Language) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Language) Column() string {
	return "language"
}

// Expression implements the Metic interface.
func (m Language) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(language)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m Language) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Language) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m Language) String() string {
	return "language"
}
