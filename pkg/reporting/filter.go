package reporting

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

// Filter filters for a Dimension connected by a logical Operator.
type Filter struct {
	Operator  Operator
	Dimension Dimension
	Values    []string
	Filter    []Filter
}

// Operator is an operator for a Filter.
type Operator int
