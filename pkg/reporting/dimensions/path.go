package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Path is a Dimension.
type Path struct{}

// Table implements the Dimension interface.
func (d Path) Table() []string {
	return []string{pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Path) Column() string {
	return "path"
}

// Expression implements the Dimension interface.
func (d Path) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d Path) ScanType() any {
	return new(string)
}
