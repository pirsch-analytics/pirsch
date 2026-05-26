package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Referrer is a Dimension.
type Referrer struct{}

// Table implements the Dimension interface.
func (d Referrer) Table() string {
	return pkg.TableSessions
}

// Column implements the Dimension interface.
func (d Referrer) Column() string {
	return "referrer"
}

// Expression implements the Dimension interface.
func (d Referrer) Expression() string {
	return ""
}
