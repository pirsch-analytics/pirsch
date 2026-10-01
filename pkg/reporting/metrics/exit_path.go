package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// ExitPath is a Metic.
type ExitPath struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m ExitPath) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Metric interface.
func (m ExitPath) TableImported() []string {
	return nil
}

// JoinTable implements the Metric interface.
func (m ExitPath) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m ExitPath) Column() string {
	return "exit_path"
}

// ColumnImported implements the Metric interface.
func (m ExitPath) ColumnImported() string {
	return ""
}

// Expression implements the Metic interface.
func (m ExitPath) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(exit_path)", false
	}

	return "", false
}

// ExpressionImported implements the Metric interface.
func (m ExitPath) ExpressionImported() string {
	return ""
}

// ScanType implements the Metric interface.
func (m ExitPath) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m ExitPath) Zero() any {
	return ""
}

// String implements the Metric interface.
func (m ExitPath) String() string {
	return "exit_path"
}
