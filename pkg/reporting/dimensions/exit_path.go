package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// ExitPath is a Dimension.
type ExitPath struct{}

// Table implements the Dimension interface.
func (d ExitPath) Table() string {
	return pkg.TableSessions
}

// Column implements the Dimension interface.
func (d ExitPath) Column() string {
	return "exit_path"
}
