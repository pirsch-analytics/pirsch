package metrics

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Time is a Metic.
type Time struct {
	// Max if set to true, returns the maximum value for this metric.
	Max bool
}

// Table implements the Metic interface.
func (m Time) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m Time) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m Time) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m Time) Column() string {
	if m.Max {
		return "max_time"
	}

	return "time"
}

// ColumnImported implements the Metric interface.
func (m Time) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m Time) Expression(_ string) (string, bool) {
	if m.Max {
		return "max(time)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m Time) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m Time) ScanType() any {
	return new(time.Time)
}

// Zero implements the Metric interface.
func (m Time) Zero() any {
	return time.Time{}
}

// String implements the Metric interface.
func (m Time) String() string {
	return "time"
}
