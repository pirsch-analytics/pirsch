package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// OSVersion is a Metic.
type OSVersion struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m OSVersion) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m OSVersion) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m OSVersion) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m OSVersion) Column() string {
	return "os_version"
}

// ColumnImported implements the Metric interface.
func (m OSVersion) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m OSVersion) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(os_version)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m OSVersion) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m OSVersion) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m OSVersion) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m OSVersion) String() string {
	return "os_version"
}
