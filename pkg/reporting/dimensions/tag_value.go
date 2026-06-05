package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// TagValue is a Dimension.
type TagValue struct {
	Key string
}

// Table implements the Dimension interface.
func (d TagValue) Table() []string {
	return []string{pkg.TablePageViews}
}

// Column implements the Dimension interface.
func (d TagValue) Column(_ string) string {
	if d.Key != "" {
		return "tag_value"
	}

	return "tags"
}

// Expression implements the Dimension interface.
func (d TagValue) Expression() string {
	if d.Key != "" {
		return "tags[?]"
	}

	return "arrayJoin(mapValues(tags))"
}

// Args implements the Dimension interface.
func (d TagValue) Args() []any {
	return []any{d.Key}
}

// ScanType implements the Metric interface.
func (d TagValue) ScanType() any {
	// string, as the ClickHouse driver does not support reading into "any" and we manually need to parse it into JSON
	return new(string)
}
