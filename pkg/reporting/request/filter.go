package request

import "github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"

const (
	OperatorGroupAnd Operator = iota
	OperatorGroupOr
	OperatorGroupNot

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
	Operator  Operator
	Dimension dimensions.Dimension
	Values    []any
	Filter    []Filter
}
