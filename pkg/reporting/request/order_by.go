package request

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
)

const (
	DirectionASC Direction = iota
	DirectionDESC
)

// OrderBy is used to sort a Request.
type OrderBy struct {
	// Dimension to sort the results by. Either this field or Metric must be set.
	Dimension *dimensions.Dimension

	// Metric to sor the results by. Either this field or Dimension must be set.
	Metric *metrics.Metric

	// Direction to sort the results by (DirectionASC or DirectionDESC).
	Direction Direction
}

// Direction is a direction to sort results.
type Direction int
