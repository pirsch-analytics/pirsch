package reporting

const (
	DirectionASC Direction = iota
	DirectionDESC
)

// OrderBy is used to sort a Request.
type OrderBy struct {
	// Dimension to sort the results by. Either this field or Metric must be set.
	Dimension *Dimension

	// Metric to sor the results by. Either this field or Dimension must be set.
	Metric *Metric

	// Direction to sort the results by (DirectionASC or DirectionDESC).
	Direction Direction
}

// Direction is a direction to sort results.
type Direction int
