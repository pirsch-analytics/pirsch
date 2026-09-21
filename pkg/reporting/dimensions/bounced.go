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

// TableImported implements the Dimension interface.
func (d Bounced) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Bounced) Column(_ string) string {
	return "is_bounce"
}

// ColumnImported implements the Dimension interface.
func (d Bounced) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Bounced) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d Bounced) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d Bounced) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Bounced) ScanType() any {
	return new(bool)
}

// String implements the Dimension interface.
func (d Bounced) String() string {
	return "bounced"
}
