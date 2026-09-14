package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Hour is a Dimension.
type Hour struct{}

// Table implements the Dimension interface.
func (d Hour) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Hour) Column(_ string) string {
	return "hour"
}

// Expression implements the Dimension interface.
func (d Hour) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfHour("time")`
	}

	return fmt.Sprintf(`toStartOfHour("time", '%s')`, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Hour) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Hour) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Hour) String() string {
	return "hour"
}
