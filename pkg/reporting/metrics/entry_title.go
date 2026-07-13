package metrics

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// EntryTitle is a Metic.
type EntryTitle struct {
	// Any if set to true, any value for this metric is returned.
	Any bool
}

// Table implements the Metic interface.
func (m EntryTitle) Table() []string {
	return []string{pkg.TableSessions}
}

// JoinTable implements the Metric interface.
func (m EntryTitle) JoinTable() string {
	return pkg.TableSessions
}

// Column implements the Metic interface.
func (m EntryTitle) Column() string {
	return "entry_title"
}

// Expression implements the Metic interface.
func (m EntryTitle) Expression(_ string) (string, bool) {
	if m.Any {
		return "any(entry_title)", false
	}

	return "", false
}

// ScanType implements the Metric interface.
func (m EntryTitle) ScanType() any {
	return new(string)
}

// Zero implements the Metric interface.
func (m EntryTitle) Zero() any {
	return ""
}
