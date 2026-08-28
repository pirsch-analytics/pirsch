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

// EventMeta is a Dimension that can act as a metric if the Function is defined.
type EventMeta struct {
	// Path is the JSON path to extract the value.
	Path string

	// Type is the value type for casting.
	Type EventMetaType

	// ColumnName is the column name for the event metadata value to prevent collisions.
	// If not set, it will be set to meta_data_value by default.
	ColumnName string

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
	} else if d.ColumnName != "" {
		return d.ColumnName
	}

	return ""
}

// Expression implements the Dimension interface.
func (d EventMeta) Expression(_ *DimensionExpressionOptions) string {
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
		return d.Expression(nil)
	}

	expression := ""
	castType := ""

	switch d.Type {
	case EventMetaTypeFloat:
		expression = fmt.Sprintf("toFloat64OrZero(toString(meta_data%s))", path)
		castType = "toFloat64"
	case EventMetaTypeInt:
		expression = fmt.Sprintf("toInt64(toFloat64OrZero(toString(meta_data%s)))", path)
		castType = "toInt64"
	default:
		expression = d.Expression(nil)
	}

	switch d.Function {
	case EventMetaFunctionAvg:
		return fmt.Sprintf("%s(avg(%s)) %s", castType, expression, d.Column(""))
	case EventMetaFunctionMedian:
		return fmt.Sprintf("%s(median(%s)) %s", castType, expression, d.Column(""))
	case EventMetaFunctionSum:
		return fmt.Sprintf("%s(sum(%s)) %s", castType, expression, d.Column(""))
	default:
		return fmt.Sprintf("%s %s", expression, d.Column(""))
	}
}
