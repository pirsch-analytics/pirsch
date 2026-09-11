package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Week is a Dimension.
type Week struct{}

// Table implements the Dimension interface.
func (d Week) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Week) Column(_ string) string {
	return "week"
}

// Expression implements the Dimension interface.
func (d Week) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfWeek("time")`
	}

	if options.Timezone == nil {
		return fmt.Sprintf(`toStartOfWeek("time", %d)`, options.WeekdayMode)
	}

	return fmt.Sprintf(`toStartOfWeek("time", %d, '%s')`, options.WeekdayMode, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Week) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Week) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Week) String() string {
	return "week"
}
