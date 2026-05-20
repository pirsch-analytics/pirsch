package request

import "github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"

const (
	OperatorIs Operator = iota
	OperatorIsNot
	OperatorContains
	OperatorContainsNot
	OperatorMatches
	OperatorMatchesNot
	OperatorAnd
	OperatorOr
	OperatorNot
)

// Operator is an operator for a Filter.
type Operator int

// Filter filters for a Dimension connected by a logical Operator.
type Filter struct {
	Operator  Operator
	Dimension dimensions.Dimension
	Values    []string
	Filter    []Filter
}
