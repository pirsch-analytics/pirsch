package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// TagValue is a Dimension.
type TagValue struct {
	// Key is the map key to extract the value.
	Key string
}

// Table implements the Dimension interface.
func (d TagValue) Table() []string {
	return []string{pkg.TablePageViews}
}

// TableImported implements the Dimension interface.
func (d TagValue) TableImported() []string {
	return nil
}

// Column implements the Dimension interface.
func (d TagValue) Column(_ string) string {
	if d.Key != "" {
		return "tag_value"
	}

	return "tags"
}

// ColumnImported implements the Dimension interface.
func (d TagValue) ColumnImported() string {
	return ""
}

// Expression implements the Dimension interface.
func (d TagValue) Expression(_ *DimensionExpressionOptions) string {
	if d.Key != "" {
		return "tags[?]"
	}

	return "arrayJoin(mapValues(tags))"
}

// ExpressionImported implements the Dimension interface.
func (d TagValue) ExpressionImported(_ *DimensionExpressionOptions) string {
	return ""
}

// Args implements the Dimension interface.
func (d TagValue) Args() []any {
	return []any{d.Key}
}

// ScanType implements the Dimension interface.
func (d TagValue) ScanType() any {
	// string, as the ClickHouse driver does not support reading into "any" and we manually need to parse it into JSON
	return new(string)
}

// String implements the Dimension interface.
func (d TagValue) String() string {
	return "tag_value"
}
