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
	if options == nil || options.Timezone == nil {
		return `toDayOfWeek("time")`
	}

	if options.Timezone == nil {
		return fmt.Sprintf(`toDayOfWeek("time", %d)`, options.WeekdayMode)
	}

	return fmt.Sprintf(`toDayOfWeek("time", %d, '%s')`, options.WeekdayMode, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Weekday) ExpressionImported(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toDayOfWeek("date")`
	}

	if options.Timezone == nil {
		return fmt.Sprintf(`toDayOfWeek("date", %d)`, options.WeekdayMode)
	}

	return fmt.Sprintf(`toDayOfWeek("date", %d, '%s')`, options.WeekdayMode, options.Timezone.String())
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
