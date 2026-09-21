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

// TableImported implements the Dimension interface.
func (d Year) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d Year) Column(_ string) string {
	return "year"
}

// ColumnImported implements the Dimension interface.
func (d Year) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d Year) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfYear("time")`
	}

	return fmt.Sprintf(`toStartOfYear("time", '%s')`, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Year) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d Year) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Year) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Year) String() string {
	return "year"
}
