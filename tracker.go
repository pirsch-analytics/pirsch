package pirsch

import (
	"net/http"
	"runtime"
	"time"
)

const (
	workerBufferSize = 100
	workerTimeout    = time.Second * 10
)

// Tracker is the main component of Pirsch.
// It provides methods to track requests and store them in a data store.
type Tracker struct {
	store Store
	hits  chan Hit
}

func NewTracker(store Store) *Tracker {
	ncpu := runtime.NumCPU()
	tracker := &Tracker{
		store: store,
		hits:  make(chan Hit, ncpu*workerBufferSize),
	}

	for i := 0; i < ncpu; i++ {
		go tracker.aggregate()
	}

	return tracker
}

func (tracker *Tracker) Hit(r *http.Request) {
	go func() {
		tracker.hits <- hitFromRequest(r)
	}()
}

func (tracker *Tracker) aggregate() {
	hits := make([]Hit, 0, workerBufferSize)

	for {
		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)

			if len(hits) > workerBufferSize/2 {
				tracker.store.Save(hits)
				hits = hits[:0]
			}
		case <-time.After(workerTimeout):
			if len(hits) > 0 {
				tracker.store.Save(hits)
				hits = hits[:0]
			}
		}
	}
}
