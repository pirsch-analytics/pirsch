package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ExitTitle is a Dimension.
type ExitTitle struct{}

// Table implements the Dimension interface.
func (d ExitTitle) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d ExitTitle) Column() string {
	return "exit_title"
}

// Expression implements the Dimension interface.
func (d ExitTitle) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d ExitTitle) ScanType() any {
	return new(string)
}
