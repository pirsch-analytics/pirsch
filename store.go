package pirsch

import "log"

// Store defines an interface to persists hits and other data.
type Store interface {
	// Save persists a list of hits.
	Save([]Hit)
}

// DefaultStore implements the Store interface and logs each request.
// The main purpose of this is testing, it is not intended for real world usage.
type DefaultStore struct{}

// Save logs the requests.
func (store *DefaultStore) Save(hits []Hit) {
	for _, hit := range hits {
		log.Println(hit)
	}
}
