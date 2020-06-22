package pirsch

import (
	"net/http"
	"runtime"
	"strings"
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

// NewTracker creates a new tracker for given store.
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

// Hit stores the given request.
// The request might be ignored if it meets certain conditions.
// The actions performed within this function run in their own goroutine, so you don't need to create one yourself.
func (tracker *Tracker) Hit(r *http.Request) {
	go func() {
		if !tracker.ignoreHit(r) {
			tracker.hits <- hitFromRequest(r)
		}
	}()
}

func (tracker *Tracker) ignoreHit(r *http.Request) bool {
	if r.Header.Get("X-Moz") == "prefetch" || // ignore bowser prefetching data
		r.Header.Get("X-Purpose") == "prefetch" ||
		r.Header.Get("X-Purpose") == "preview" ||
		r.Header.Get("Purpose") == "prefetch" ||
		r.Header.Get("Purpose") == "preview" {
		return true
	}

	userAgent := strings.ToLower(r.Header.Get("User-Agent"))

	return strings.Contains(userAgent, "bot") || // words often used in bot names
		strings.Contains(userAgent, "crawler") ||
		strings.Contains(userAgent, "spider") ||
		strings.Contains(userAgent, "://") // URLs
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
