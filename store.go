package pirsch

var (
	store Store
)

// Store defines an interface to store data.
type Store interface {
	// Save stores a batch of hits.
	Save([]Hit)
}

// DefaultStore implements the Store interface and does nothing.
type DefaultStore struct{}

// Save does nothing.
func (store *DefaultStore) Save(hits []Hit) {}

// SetStore sets the store used to save data.
func SetStore(s Store) {
	store = s
}
