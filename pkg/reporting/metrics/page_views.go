package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// PageViews is a Metric.
type PageViews struct{}

// Table implements the Metric interface.
func (m PageViews) Table() string {
	return pkg.TableSessions
}

// Column implements the Metric interface.
func (m PageViews) Column() string {
	return "page_views"
}

// Expression implements the Metric interface.
func (m PageViews) Expression() string {
	return "sum(page_views)"
}
