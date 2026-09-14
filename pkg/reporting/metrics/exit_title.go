package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// ExitTitle is a Metic.
type ExitTitle struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m ExitTitle) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m ExitTitle) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m ExitTitle) Column() string {
	return "exit_title"
}

// Expression implements the Metic interface.
func (m ExitTitle) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(exit_title)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m ExitTitle) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m ExitTitle) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m ExitTitle) String() string {
	return "exit_title"
}
