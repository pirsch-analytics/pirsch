package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Extended is a Metic.
type Extended struct {
	// Max if set to true, returns the maximum value for this metric.
	Max bool
}

// Table implements the Metic interface.
func (m Extended) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m Extended) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Extended) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m Extended) Column() string {
	return "extended"
}

// ColumnImported implements the Metric interface.
func (m Extended) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m Extended) Expression(_ string) (string, bool) {
	if m.Max {
		return "max(extended)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m Extended) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Extended) ScanType() any {
	return new(uint16)
}

// Zero implements the Metric interface.
func (m Extended) Zero() any {
	return uint16(0)
}

// String implements the Metric interface.
func (m Extended) String() string {
	return "extended"
}
