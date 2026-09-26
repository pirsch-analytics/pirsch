package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Duration is a Metic.
type Duration struct {
	// Max if set to true, returns the maximum value for this metric.
	Max bool
}

// Table implements the Metic interface.
func (m Duration) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m Duration) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Duration) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m Duration) Column() string {
	return "duration_seconds"
}

// ColumnImported implements the Metric interface.
func (m Duration) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m Duration) Expression(_ string) (string, bool) {
	if m.Max {
		return "max(duration_seconds)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m Duration) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Duration) ScanType() any {
	return new(uint32)
}

// Zero implements the Metric interface.
func (m Duration) Zero() any {
	return uint32(0)
}

// String implements the Metric interface.
func (m Duration) String() string {
	return "duration"
}
