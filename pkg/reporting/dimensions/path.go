package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Path is a Dimension.
type Path struct{}

// Table implements the Dimension interface.
func (d Path) Table() []string {
	return []string{pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Path) Column(table string) string {
	if table == pkg.TableSessions {
		return "entry_path"
	}

	return "path"
}

// Expression implements the Dimension interface.
func (d Path) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Path) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Path) ScanType() any {
	return new(string)
}
