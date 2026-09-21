package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// OS is a Dimension.
type OS struct{}

// Table implements the Dimension interface.
func (d OS) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Dimension interface.
func (d OS) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d OS) Column(_ string) string {
	return "os"
}

// ColumnImported implements the Dimension interface.
func (d OS) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d OS) Expression(_ *DimensionExpressionOptions) string {
	return ""
}

// ExpressionImported implements the Dimension interface.
func (d OS) ExpressionImported() string {
	return ""
}

// Args implements the Dimension interface.
func (d OS) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d OS) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d OS) String() string {
	return "os"
}
