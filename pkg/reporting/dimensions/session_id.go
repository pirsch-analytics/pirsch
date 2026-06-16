package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// SessionID is a Dimension.
type SessionID struct{}

// Table implements the Dimension interface.
func (d SessionID) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d SessionID) Column(_ string) string {
	return "session_id"
}

// Expression implements the Dimension interface.
func (d SessionID) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d SessionID) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d SessionID) ScanType() any {
	return new(uint32)
}
