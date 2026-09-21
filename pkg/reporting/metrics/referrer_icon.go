package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ReferrerIcon is a Metic.
type ReferrerIcon struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m ReferrerIcon) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m ReferrerIcon) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m ReferrerIcon) JoinTable() string {
	return ""
}

// Column implements the Metic interface.
func (m ReferrerIcon) Column() string {
	return "referrer_icon"
}

// ColumnImported implements the Metric interface.
func (m ReferrerIcon) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m ReferrerIcon) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(referrer_icon)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m ReferrerIcon) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m ReferrerIcon) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m ReferrerIcon) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m ReferrerIcon) String() string {
	return "referrer_icon"
}
