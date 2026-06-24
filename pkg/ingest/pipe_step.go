package ingest

// PipeStep processes steps for a Pipe.
type PipeStep interface {
	// Step return values indicate if the Pipe should continue processing the Request or if there was an error.
	// If the step is supposed to enrich the Request, it can modify the object as a side effect.
	Step(*Request) (bool, error)
}
