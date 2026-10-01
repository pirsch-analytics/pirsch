package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Start is a Dimension.
type Start struct{}

// Table implements the Dimension interface.
func (d Start) Table() []string {
	return []string{pkg.TableSessions}
}

// TableImported implements the Dimension interface.
func (d Start) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Start) Column(_ string) string {
	return "start"
}

// ColumnImported implements the Dimension interface.
func (d Start) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Start) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return ""
	}

	return fmt.Sprintf(`toTimezone("start", '%s')`, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Start) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Start) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Start) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Start) String() string {
	return "start"
}
