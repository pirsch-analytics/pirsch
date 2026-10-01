package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// City is a Metic.
type City struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m City) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m City) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m City) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m City) Column() string {
	return "city"
}

// ColumnImported implements the Metric interface.
func (m City) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m City) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(city)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m City) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m City) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m City) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m City) String() string {
	return "city"
}
