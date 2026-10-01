package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Channel is a Metic.
type Channel struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m Channel) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m Channel) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Channel) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Channel) Column() string {
	return "channel"
}

// ColumnImported implements the Metric interface.
func (m Channel) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m Channel) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(channel)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m Channel) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Channel) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m Channel) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m Channel) String() string {
	return "channel"
}
