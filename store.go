package pirsch

import "log"

// Store defines an interface to persists hits.
type Store interface {
	// Save stores a batch of hits.
	Save([]Hit)
}

// DefaultStore implements the Store interface and logs each request.
// The main purpose for this is for testing, not for usage in a real application.
type DefaultStore struct{}

// Save logs the requests.
func (store *DefaultStore) Save(hits []Hit) {
	for _, hit := range hits {
		log.Println(hit)
	}
}
