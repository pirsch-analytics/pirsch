package dimensions

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// EventMetaKey is a Dimension.
// It returns all paths for an event metadata field.
type EventMetaKey struct {
	// Any specifies whether the result set should not be grouped by the metadata keys.
	Any bool
}

// Table implements the Dimension interface.
func (d EventMetaKey) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d EventMetaKey) Column(_ string) string {
	return "meta_data"
}

// Expression implements the Dimension interface.
func (d EventMetaKey) Expression(_ *DimensionExpressionOptions) string {
	if d.Any {
		return "any(JSONAllPaths(meta_data))"
	}

	return "JSONAllPaths(meta_data)"
}

// Args implements the Dimension interface.
func (d EventMetaKey) Args() []any {
	return nil
}

// ScanType implements the Dimension interface.
func (d EventMetaKey) ScanType() any {
	return new([]string)
}

// String implements the Dimension interface.
func (d EventMetaKey) String() string {
	return "event_meta_key"
}
