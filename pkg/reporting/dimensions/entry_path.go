package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// EntryPath is a Dimension.
type EntryPath struct{}

// Table implements the Dimension interface.
func (d EntryPath) Table() string {
	return pkg.TableSessions
}

// Column implements the Dimension interface.
func (d EntryPath) Column() string {
	return "entry_path"
}

// Expression implements the Dimension interface.
func (d EntryPath) Expression() string {
	return ""
}
