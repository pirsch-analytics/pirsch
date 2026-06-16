package dimensions

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Tags is a Dimension.
type Tags struct{}

// Table implements the Dimension interface.
func (d Tags) Table() []string {
	return []string{pkg.TablePageViews}
}

// Column implements the Dimension interface.
func (d Tags) Column(_ string) string {
	return "tags"
}

// Expression implements the Dimension interface.
func (d Tags) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Tags) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Tags) ScanType() any {
	// string, as the ClickHouse driver does not support reading into "any" and we manually need to parse it into JSON
	return new(map[string]string)
}
