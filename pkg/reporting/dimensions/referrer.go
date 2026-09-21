package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Referrer is a Dimension.
type Referrer struct {
	// Any specifies whether the result set should not be grouped by the referrer.
	Any bool
}

// Table implements the Dimension interface.
func (d Referrer) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Referrer) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Referrer) Column(_ string) string {
	return "referrer"
}

// ColumnImported implements the Dimension interface.
func (d Referrer) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Referrer) Expression(_ *DimensionExpressionOptions) string {
	if d.Any {
		return "any(referrer)"
	}

	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Referrer) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d Referrer) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Referrer) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Referrer) String() string {
	return "referrer"
}
