package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Day is a Dimension.
type Day struct{}

// Table implements the Dimension interface.
func (d Day) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Day) Column(_ string) string {
	return "day"
}

// Expression implements the Dimension interface.
func (d Day) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toDate("time")`
	}

	return fmt.Sprintf(`toDate("time", '%s')`, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Day) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Day) ScanType() any {
	return new(time.Time)
}
