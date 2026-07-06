package dimensions

import (
	"fmt"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

// EventMetaValue is a Dimension.
type EventMetaValue struct {
	// Path is the JSON path to extract the value.
	Path string

	// Type is the value type for casting.
	Type EventMetaType
}

// Table implements the Dimension interface.
func (d EventMetaValue) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d EventMetaValue) Column(_ string) string {
	return "meta_data_value"
}

// Expression implements the Dimension interface.
func (d EventMetaValue) Expression() string {
	return "toString(meta_data)"
}

// Args implements the Dimension interface.
func (d EventMetaValue) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d EventMetaValue) ScanType() any {
	if d.Type == EventMetaTypeNone {
		// string, as the ClickHouse driver does not support reading into "any" and we manually need to parse it into JSON
		return new(string)
	}

	// unless we perform an aggregation, in which case we know the type
	if d.Type == EventMetaTypeInt {
		return new(int64)
	}

	return new(float64)
}

// Select returns the SQL select expression applying any configured function or type cast.
func (d EventMetaValue) Select(path string) string {
	switch d.Type {
	case EventMetaTypeFloat:
		return fmt.Sprintf("toFloat64OrZero(toString(meta_data%s)) meta_data_value", path)
	case EventMetaTypeInt:
		return fmt.Sprintf("toInt64OrZero(toString(meta_data%s)) meta_data_value", path)
	default:
		return fmt.Sprintf("toString(meta_data%s) meta_data_value", path)
	}
}
