package ingest

// VisitorIDProvider is an optional interface for a PipeStep that calculates the visitor ID.
// It's used to set the visitor ID for requests that have been canceled before the visitor ID has been set.
type VisitorIDProvider interface {
	VisitorID(*Request) uint64
}
