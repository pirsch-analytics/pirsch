package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// PageViews is a Metric.
type PageViews struct{}

// Table implements the Metric interface.
func (m PageViews) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews}
}

// JoinTable implements the Metric interface.
func (m PageViews) JoinTable() string {
	return ""
}

// Column implements the Metric interface.
func (m PageViews) Column() string {
	return "page_views"
}

// Expression implements the Metric interface.
func (m PageViews) Expression(table string) (string, bool) {
	if table == pkg.TableSessions {
		return "sum(page_views)", false
	}

	return "count(*)", false
}

// ScanType implements the Metric interface.
func (m PageViews) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m PageViews) Zero() any {
	return uint64(0)
}
