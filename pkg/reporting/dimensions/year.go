package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Year is a Dimension.
type Year struct{}

// Table implements the Dimension interface.
func (d Year) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Year) Column(_ string) string {
	return "year"
}

// Expression implements the Dimension interface.
func (d Year) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfYear("time")`
	}

	return fmt.Sprintf(`toStartOfYear("time", '%s')`, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Year) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Year) ScanType() any {
	return new(time.Time)
}
