package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Bounced is a Dimension.
type Bounced struct{}

// Table implements the Dimension interface.
func (d Bounced) Table() []string {
	return []string{pkg.TableSessions}
}

// Column implements the Dimension interface.
func (d Bounced) Column(_ string) string {
	return "is_bounce"
}

// Expression implements the Dimension interface.
func (d Bounced) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Bounced) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Bounced) ScanType() any {
	return new(bool)
}
