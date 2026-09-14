package request

import "encoding/json/v2"

// Pagination limits the number of results for a Report.
type Pagination struct {
	// Offset limits the number of results. Offset <= 0 means no offset.
	Offset int

	// Limit limits the number of results. Limit <= 0 means unlimited.
	Limit int
}

// String implements the fmt.Stringer interface.
func (p *Pagination) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}
