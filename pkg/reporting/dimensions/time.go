package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Time is a Dimension.
type Time struct{}

// Table implements the Dimension interface.
func (d Time) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Time) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Time) Column(_ string) string {
	return "time"
}

// ColumnImported implements the Dimension interface.
func (d Time) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Time) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return ""
	}

	return fmt.Sprintf(`toTimezone("time", '%s')`, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Time) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d Time) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Time) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Time) String() string {
	return "time"
}
