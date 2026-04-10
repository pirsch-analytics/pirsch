package ingest

// PipeFunc is a function processing a step of a Pipe.
// The return values indicate if the Pipe should continue processing the Request or if there was an error.
// If the step is supposed to enrich the Request, it can modify the object as a side effect.
type PipeFunc func(*Request) (bool, error)
