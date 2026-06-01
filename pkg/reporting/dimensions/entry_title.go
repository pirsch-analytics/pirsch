package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// EntryTitle is a Dimension.
type EntryTitle struct{}

// Table implements the Dimension interface.
func (d EntryTitle) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d EntryTitle) Column() string {
	return "entry_title"
}

// Expression implements the Dimension interface.
func (d EntryTitle) Expression() string {
	return ""
}

// ScanType implements the Metric interface.
func (d EntryTitle) ScanType() any {
	return new(string)
}
