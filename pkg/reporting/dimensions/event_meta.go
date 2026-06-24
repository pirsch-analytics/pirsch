package dimensions

import (
	"fmt"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

const (
	EventMetaTypeNone = EventMetaType(iota)
	EventMetaTypeFloat
	EventMetaTypeInt
)

const (
	EventMetaFunctionNone = EventMetaFunction(iota)
	EventMetaFunctionSum
	EventMetaFunctionAvg
	EventMetaFunctionMedian
)

// EventMetaType is the cast type for an event meta value.
type EventMetaType uint8

// EventMetaFunction is the function used to calculate an event meta value.
type EventMetaFunction uint8

// EventMeta is a Dimension.
type EventMeta struct {
	// Path is the JSON path to extract the value.
	Path string

	// Type is the value type for casting.
	Type EventMetaType

	// Function is the function used for calculations.
	Function EventMetaFunction
}

// Table implements the Dimension interface.
func (d EventMeta) Table() []string {
	return []string{pkg.TableEvents}
}

// Column implements the Dimension interface.
func (d EventMeta) Column(_ string) string {
	if d.Type == EventMetaTypeNone && d.Function == EventMetaFunctionNone {
		return "meta_data"
	} else if d.Function == EventMetaFunctionNone {
		return "meta_data_value"
	}

	return ""
}

// Expression implements the Dimension interface.
func (d EventMeta) Expression() string {
	return "toString(meta_data)"
}

// Args implements the Dimension interface.
func (d EventMeta) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d EventMeta) ScanType() any {
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
func (d EventMeta) Select(path string) string {
	if d.Type == EventMetaTypeNone && d.Function == EventMetaFunctionNone {
		return d.Expression()
	}

	expression := ""

	switch d.Type {
	case EventMetaTypeFloat:
		expression = fmt.Sprintf("toFloat64OrZero(toString(meta_data%s))", path)
	case EventMetaTypeInt:
		expression = fmt.Sprintf("toInt64OrZero(toString(meta_data%s))", path)
	default:
		expression = d.Expression()
	}

	switch d.Function {
	case EventMetaFunctionAvg:
		return fmt.Sprintf("avg(%s) meta_data_value", expression)
	case EventMetaFunctionMedian:
		return fmt.Sprintf("median(%s) meta_data_value", expression)
	case EventMetaFunctionSum:
		return fmt.Sprintf("sum(%s) meta_data_value", expression)
	default:
		return fmt.Sprintf("%s meta_data_value", expression)
	}
}
