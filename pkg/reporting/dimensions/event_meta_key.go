package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// EventMetaKey is a Dimension.
// It's only really useful as a filter. The dimension will simply return the metadata column.
type EventMetaKey struct{}

// Table implements the Dimension interface.
func (d EventMetaKey) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d EventMetaKey) Column(_ string) string {
	return "meta_data"
}

// Expression implements the Dimension interface.
func (d EventMetaKey) Expression() string {
	return "toString(meta_data)"
}

// ScanType implements the Metric interface.
func (d EventMetaKey) ScanType() any {
	// string, as the ClickHouse driver does not support reading into "any" and we manually need to parse it into JSON
	return new(string)
}
