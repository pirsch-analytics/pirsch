package dimensions

import (
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Time is a Dimension.
type Time struct{}

// Table implements the Dimension interface.
func (d Time) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d Time) Column(_ string) string {
	return "time"
}

// Expression implements the Dimension interface.
func (d Time) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Time) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Time) ScanType() any {
	return new(time.Time)
}
