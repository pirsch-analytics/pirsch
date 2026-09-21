package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ScreenClass is a Metic.
type ScreenClass struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m ScreenClass) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m ScreenClass) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m ScreenClass) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m ScreenClass) Column() string {
	return "screen_class"
}

// ColumnImported implements the Metric interface.
func (m ScreenClass) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m ScreenClass) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(screen_class)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m ScreenClass) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m ScreenClass) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m ScreenClass) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m ScreenClass) String() string {
	return "screen_class"
}
