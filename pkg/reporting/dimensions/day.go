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

// TableImported implements the Dimension interface.
func (d Day) TableImported() []string {
	return pkg.ImportedTables
}

// Column implements the Dimension interface.
func (d Day) Column(_ string) string {
	return "day"
}

// ColumnImported implements the Dimension interface.
func (d Day) ColumnImported() string {
	return "day"
}

// Expression implements the Dimension interface.
func (d Day) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toDate("time")`
	}

	return fmt.Sprintf(`toDate("time", '%s')`, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Day) ExpressionImported(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toDate("date")`
	}

	return fmt.Sprintf(`toDate("date", '%s')`, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Day) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Day) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Day) String() string {
	return "day"
}
