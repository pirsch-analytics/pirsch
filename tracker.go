package pirsch

import (
	"context"
	"net/http"
	"net/url"
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
	worker           int
	workerBufferSize int
	workerTimeout    time.Duration
	workerCancel     context.CancelFunc
	workerDone       chan bool
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
		worker:           worker,
		workerBufferSize: bufferSize,
		workerTimeout:    timeout,
		workerDone:       make(chan bool),
	}
	tracker.startWorker()
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

// HitPage works like Hit, but sets the path to the given path.
// This can be useful in case you have a single endpoint to track requests that you call from JavaScript for example.
func (tracker *Tracker) HitPage(r *http.Request, path string) {
	go func() {
		if !tracker.ignoreHit(r) {
			u, err := url.Parse(r.RequestURI)

			if err == nil {
				hit := hitFromRequest(r)
				u.Path = path
				hit.Path = path
				hit.URL = u.String()
				tracker.hits <- hit
			}
		}
	}()
}

// Flush flushes all hits to store.
func (tracker *Tracker) Flush() {
	tracker.Stop()
	tracker.startWorker()
}

// Stop flushes and stops all workers.
func (tracker *Tracker) Stop() {
	tracker.workerCancel()

	for i := 0; i < tracker.worker; i++ {
		<-tracker.workerDone
	}
}

func (tracker *Tracker) ignoreHit(r *http.Request) bool {
	if r.Header.Get("X-Moz") == "prefetch" || // ignore browser prefetching data
		r.Header.Get("X-Purpose") == "prefetch" ||
		r.Header.Get("X-Purpose") == "preview" ||
		r.Header.Get("Purpose") == "prefetch" ||
		r.Header.Get("Purpose") == "preview" {
		return true
	}

	userAgent := strings.ToLower(r.Header.Get("User-Agent"))

	if strings.Contains(userAgent, "bot") || // words often used in bot names
		strings.Contains(userAgent, "crawler") ||
		strings.Contains(userAgent, "spider") ||
		strings.Contains(userAgent, "://") { // URLs
		return true
	}

	for _, botUserAgent := range userAgentBotList {
		if strings.Contains(userAgent, botUserAgent) {
			return true
		}
	}

	return false
}

func (tracker *Tracker) startWorker() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tracker.workerCancel = cancelFunc

	for i := 0; i < tracker.worker; i++ {
		go tracker.aggregate(ctx)
	}
}

func (tracker *Tracker) aggregate(ctx context.Context) {
	hits := make([]Hit, 0, tracker.workerBufferSize)

	for {
		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)

			if len(hits) == tracker.workerBufferSize {
				panicOnErr(tracker.store.Save(hits))
				hits = hits[:0]
			}
		case <-time.After(tracker.workerTimeout):
			if len(hits) > 0 {
				panicOnErr(tracker.store.Save(hits))
				hits = hits[:0]
			}
		case <-ctx.Done():
			if len(hits) > 0 {
				panicOnErr(tracker.store.Save(hits))
			}

			tracker.workerDone <- true
			return
		}
	}
}
