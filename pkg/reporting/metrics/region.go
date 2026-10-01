package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Region is a Metic.
type Region struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Region) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m Region) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Region) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Region) Column() string {
	return "region"
}

// ColumnImported implements the Metric interface.
func (m Region) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m Region) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(region)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m Region) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Region) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Region) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m Region) String() string {
	return "region"
}
