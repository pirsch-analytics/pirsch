package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Day is a Dimension.
type Day struct{}

// Table implements the Dimension interface.
func (d Day) Table() string {
	return pkg.TableSessions
}

// Column implements the Dimension interface.
func (d Day) Column() string {
	return "day"
}

// Expression implements the Dimension interface.
func (d Day) Expression() string {
	return `toDate("time")`
}
