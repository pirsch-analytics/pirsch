package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// EntryPath is a Metic.
type EntryPath struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m EntryPath) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m EntryPath) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m EntryPath) Column() string {
	return "entry_path"
}

// Expression implements the Metic interface.
func (m EntryPath) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(entry_path)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m EntryPath) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m EntryPath) Zero() any {
	return ""
}
