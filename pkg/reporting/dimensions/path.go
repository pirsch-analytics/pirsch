package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Path is a Dimension.
type Path struct{}

// Table implements the Dimension interface.
func (d Path) Table() string {
	return pkg.TablePageViews
}

// Column implements the Dimension interface.
func (d Path) Column() string {
	return "path"
}
