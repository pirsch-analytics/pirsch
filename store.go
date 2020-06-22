package pirsch

// Store defines an interface to persists hits and other data.
type Store interface {
	// Save persists a list of hits.
	Save([]Hit)
}
