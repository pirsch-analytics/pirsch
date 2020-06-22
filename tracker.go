package pirsch

import (
	"net/http"
	"runtime"
	"strings"
	"time"
)

const (
	defaultWorkerBufferSize = 100
	defaultWorkerTimeout    = time.Second * 10
)

// TrackerConfig is the optional configuration for the Tracker.
type TrackerConfig struct {
	// Worker sets the number of workers that are used to store hits.
	// Must be greater or equal to 1.
	Worker int

	// WorkerBufferSize is the size of the buffer used to store hits.
	// Must be greater or equal to 2. The hits are stored when the buffer size reaches half of its maximum.
	WorkerBufferSize int

	// WorkerTimeout sets the timeout used to store hits.
	// This is used to allow the workers to store hits even if the buffer is not full.
	// It's recommended to set this to a few seconds.
	WorkerTimeout time.Duration
}

// Tracker is the main component of Pirsch.
// It provides methods to track requests and store them in a data store.
// In case of an error it will panic.
type Tracker struct {
	store            Store
	hits             chan Hit
	flush            chan bool
	worker           int
	workerBufferSize int
	workerTimeout    time.Duration
}

// NewTracker creates a new tracker for given store and config.
// Pass nil for the config to use the defaults.
func NewTracker(store Store, config *TrackerConfig) *Tracker {
	worker := runtime.NumCPU()
	bufferSize := defaultWorkerBufferSize
	timeout := defaultWorkerTimeout

	if config != nil {
		if config.Worker > 0 {
			worker = config.Worker
		}

		if config.WorkerBufferSize > 1 {
			bufferSize = config.WorkerBufferSize
		}

		if config.WorkerTimeout > 0 {
			timeout = config.WorkerTimeout
		}
	}

	tracker := &Tracker{
		store:            store,
		hits:             make(chan Hit, worker*bufferSize),
		flush:            make(chan bool),
		worker:           worker,
		workerBufferSize: bufferSize,
		workerTimeout:    timeout,
	}

	for i := 0; i < worker; i++ {
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

// Flush flushes all hits to store and stops recording hits.
func (tracker *Tracker) Flush() {
	for i := 0; i < tracker.worker; i++ {
		tracker.flush <- true
		<-tracker.flush
	}
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
	hits := make([]Hit, 0, tracker.workerBufferSize)

	for {
		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)

			if len(hits) > tracker.workerBufferSize/2 {
				panicOnErr(tracker.store.Save(hits))
				hits = hits[:0]
			}
		case <-time.After(tracker.workerTimeout):
			if len(hits) > 0 {
				panicOnErr(tracker.store.Save(hits))
				hits = hits[:0]
			}
		case <-tracker.flush:
			if len(hits) > 0 {
				panicOnErr(tracker.store.Save(hits))
			}

			// signal that we're done and close worker
			tracker.flush <- true
			break
		}
	}
}
