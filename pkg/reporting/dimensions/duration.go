package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Duration is a Dimension.
type Duration struct{}

// Table implements the Dimension interface.
func (d Duration) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d Duration) Column(_ string) string {
	return "duration_seconds"
}

// Expression implements the Dimension interface.
func (d Duration) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Duration) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Duration) ScanType() any {
	return new(uint32)
}
