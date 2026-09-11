package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// TagKey is a Dimension.
type TagKey struct{}

// Table implements the Dimension interface.
func (d TagKey) Table() []string {
	return []string{pkg.TablePageViews}
}

// Column implements the Dimension interface.
func (d TagKey) Column(_ string) string {
	return "tag_keys"
}

// Expression implements the Dimension interface.
func (d TagKey) Expression(_ *DimensionExpressionOptions) string {
	return "arrayJoin(mapKeys(tags))"
}

// Args implements the Dimension interface.
func (d TagKey) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d TagKey) ScanType() any {
	// string, as the ClickHouse driver does not support reading into "any" and we manually need to parse it into JSON
	return new(string)
}

// String implements the Dimension interface.
func (d TagKey) String() string {
	return "tag_key"
}
