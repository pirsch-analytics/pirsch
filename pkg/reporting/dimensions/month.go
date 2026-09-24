package dimensions

import (
	"fmt"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Month is a Dimension.
type Month struct{}

// Table implements the Dimension interface.
func (d Month) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d Month) TableImported() []string {
	return pkg.ImportedTables
}

// Column implements the Dimension interface.
func (d Month) Column(_ string) string {
	return "month"
}

// ColumnImported implements the Dimension interface.
func (d Month) ColumnImported() string {
	return "month"
}

// Expression implements the Dimension interface.
func (d Month) Expression(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfMonth("time")`
	}

	return fmt.Sprintf(`toStartOfMonth("time", '%s')`, options.Timezone.String())
}

// ExpressionImported implements the Dimension interface.
func (d Month) ExpressionImported(options *DimensionExpressionOptions) string {
	if options == nil || options.Timezone == nil {
		return `toStartOfMonth("date")`
	}

	return fmt.Sprintf(`toStartOfMonth("date", '%s')`, options.Timezone.String())
}

// Args implements the Dimension interface.
func (d Month) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Month) ScanType() any {
	return new(time.Time)
}

// String implements the Dimension interface.
func (d Month) String() string {
	return "month"
}
