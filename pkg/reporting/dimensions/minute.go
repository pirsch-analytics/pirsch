package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Minute is a Dimension.
type Minute struct{}

// Table implements the Dimension interface.
func (d Minute) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Minute) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Minute) Column(_ string) string {
	return "minute"
}

// ColumnImported implements the Dimension interface.
func (d Minute) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Minute) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfMinute("time")`
	}

	return fmt.Sprintf(`toStartOfMinute("time", '%s')`, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Minute) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d Minute) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Minute) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Minute) String() string {
	return "minute"
}
