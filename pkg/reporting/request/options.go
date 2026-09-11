package request

import "encoding/json/v2"

// Options are optional fields for a Request.
type Options struct {
	// Sample sets the sampling size.
	Sample uint
}

// String implements the fmt.Stringer interface.
func (o *Options) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}
