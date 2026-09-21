package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ReferrerIcon is a Dimension.
type ReferrerIcon struct{}

// Table implements the Dimension interface.
func (d ReferrerIcon) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d ReferrerIcon) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d ReferrerIcon) Column(_ string) string {
	return "referrer_icon"
}

// ColumnImported implements the Dimension interface.
func (d ReferrerIcon) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d ReferrerIcon) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d ReferrerIcon) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d ReferrerIcon) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d ReferrerIcon) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d ReferrerIcon) String() string {
	return "referrer_icon"
}
