package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// EventMeta is a Dimension.
type EventMeta struct {
	Path string
}

// Table implements the Dimension interface.
func (d EventMeta) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d EventMeta) Column(_ string) string {
	return "meta_data"
}

// Expression implements the Dimension interface.
func (d EventMeta) Expression() string {
	return "toString(meta_data)"
}

// ScanType implements the Metric interface.
func (d EventMeta) ScanType() any {
	// string, as the ClickHouse driver does not support reading into "any" and we manually need to parse it into JSON
	return new(string)
}
