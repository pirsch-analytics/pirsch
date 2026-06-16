package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// PageViews is a Dimension.
type PageViews struct{}

// Table implements the Dimension interface.
func (d PageViews) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d PageViews) Column(_ string) string {
	return "page_views"
}

// Expression implements the Dimension interface.
func (d PageViews) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d PageViews) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d PageViews) ScanType() any {
	return new(uint16)
}
