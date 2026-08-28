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

// Column implements the Dimension interface.
func (d Time) Column(_ string) string {
	return "time"
}

// Expression implements the Dimension interface.
func (d Time) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return ""
	}

	return fmt.Sprintf(`toTimezone("time", '%s')`, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Time) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Time) ScanType() any {
	return new(time.Time)
}
