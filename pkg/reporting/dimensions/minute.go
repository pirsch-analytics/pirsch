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

// Column implements the Dimension interface.
func (d Minute) Column(_ string) string {
	return "minute"
}

// Expression implements the Dimension interface.
func (d Minute) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfMinute("time")`
	}

	return fmt.Sprintf(`toStartOfMinute("time", '%s')`, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Minute) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Minute) ScanType() any {
	return new(time.Time)
}
