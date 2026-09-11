package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// Country is a Dimension.
type Country struct {
	// Any specifies whether the result set should not be grouped by the country.
	Any bool
}

// Table implements the Dimension interface.
func (d Country) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d Country) Column(_ string) string {
	return "country_code"
}

// Expression implements the Dimension interface.
func (d Country) Expression(_ *DimensionExpressionOptions) string {
	if d.Any {
		return "any(country_code)"
	}

	return ""
}

// Args implements the Dimension interface.
func (d Country) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d Country) ScanType() any {
	return new(string)
}

// String implements the Dimension interface.
func (d Country) String() string {
	return "country"
}
