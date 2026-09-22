package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// SiteID is a Dimension.
type SiteID struct{}

// Table implements the Dimension interface.
func (d SiteID) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d SiteID) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d SiteID) Column(_ string) string {
	return "site_id"
}

// ColumnImported implements the Dimension interface.
func (d SiteID) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d SiteID) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d SiteID) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d SiteID) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d SiteID) ScanType() any {
	return new(uint64)
}

// String implements the Dimension interface.
func (d SiteID) String() string {
	return "site_id"
}
