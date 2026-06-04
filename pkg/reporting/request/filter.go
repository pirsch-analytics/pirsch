package request

import "github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"

const (
	OperatorAnd Operator = iota
	OperatorOr
	OperatorNot

	OperatorIs
	OperatorIsNot
	OperatorContains
	OperatorContainsNot
	OperatorMatches
	OperatorMatchesNot
)

// Operator is an operator for a Filter.
type Operator int

// Filter filters for a Dimension connected by a logical Operator.
type Filter struct {
	Operator      Operator
	Dimension     dimensions.Dimension
	EventMetaPath string
	Values        []any
	Filter        []Filter
}
