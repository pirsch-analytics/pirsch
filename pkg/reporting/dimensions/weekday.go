package dimensions

import (
	"fmt"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Weekday is a Dimension.
type Weekday struct{}

// Table implements the Dimension interface.
func (d Weekday) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Weekday) TableImported() []string {
	return pkg.ImportedTables
}

// Column implements the Dimension interface.
func (d Weekday) Column(_ string) string {
	return "weekday"
}

// ColumnImported implements the Dimension interface.
func (d Weekday) ColumnImported() string {
	return "weekday"
}

// Expression implements the Dimension interface.
func (d Weekday) Expression(options *DimensionExpressionOptions) string {
	mode := weekdayMode(options)

	if options == nil || options.Timezone == nil {
		return fmt.Sprintf(`toDayOfWeek("time", %d)`, mode)
	}

	return fmt.Sprintf(`toDayOfWeek("time", %d, '%s')`, mode, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Weekday) ExpressionImported(options *DimensionExpressionOptions) string {
	return fmt.Sprintf(`toDayOfWeek("date", %d)`, weekdayMode(options))
}

// Args implements the Dimension interface.
func (d Weekday) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Weekday) ScanType() any {
	return new(uint8)
}

// String implements the Dimension interface.
func (d Weekday) String() string {
	return "weekday"
}

func weekdayMode(options *DimensionExpressionOptions) int {
	if options == nil || options.WeekdayMode == 0 {
		return 1
	}

	return options.WeekdayMode
}
