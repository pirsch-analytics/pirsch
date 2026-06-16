package dimensions

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Start is a Dimension.
type Start struct{}

// Table implements the Dimension interface.
func (d Start) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d Start) Column(_ string) string {
	return "start"
}

// Expression implements the Dimension interface.
func (d Start) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Start) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Start) ScanType() any {
	return new(time.Time)
}
