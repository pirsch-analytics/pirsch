package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// EntryPath is a Dimension.
type EntryPath struct{}

// Table implements the Dimension interface.
func (d EntryPath) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d EntryPath) Column(_ string) string {
	return "entry_path"
}

// Expression implements the Dimension interface.
func (d EntryPath) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d EntryPath) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d EntryPath) ScanType() any {
	return new(string)
}
