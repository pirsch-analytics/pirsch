package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// ExitPath is a Dimension.
type ExitPath struct{}

// Table implements the Dimension interface.
func (d ExitPath) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d ExitPath) Column(_ string) string {
	return "exit_path"
}

// Expression implements the Dimension interface.
func (d ExitPath) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d ExitPath) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d ExitPath) ScanType() any {
	return new(string)
}
