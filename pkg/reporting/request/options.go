package request

import "encoding/json/v2"

// Options are optional fields for a Request.
type Options struct {
	// Sample sets the sampling size.
	Sample uint

	// IncludeImportedStatistics sets whether imported statistics should be included in the result set.
	// Certain rules apply when this is set to true:
	//  * Only one dimension (besides the date) that is available from an imported statistics table shall be set
	//  * All metrics must be available from the dimension's imported statistics table
	//  * The period granularity must be one day or greater
	IncludeImportedStatistics bool
}

// String implements the fmt.Stringer interface.
func (o *Options) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}
