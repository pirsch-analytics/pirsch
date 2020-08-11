package pirsch

import (
	"context"
	"net/http"
	"runtime"
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
	// Must be greater than 0. The hits are stored in batch when the buffer is full.
	WorkerBufferSize int

	// WorkerTimeout sets the timeout used to store hits.
	// This is used to allow the workers to store hits even if the buffer is not full yet.
	// It's recommended to set this to a few seconds.
	WorkerTimeout time.Duration

	// RefererDomainBlacklist see HitOptions.RefererDomainBlacklist.
	RefererDomainBlacklist []string

	// RefererDomainBlacklistIncludesSubdomains see HitOptions.RefererDomainBlacklistIncludesSubdomains.
	RefererDomainBlacklistIncludesSubdomains bool
}

func (config *TrackerConfig) validate() {
	if config.Worker < 1 {
		config.Worker = runtime.NumCPU()
	}

	if config.WorkerBufferSize < 1 {
		config.WorkerBufferSize = defaultWorkerBufferSize
	}

	if config.WorkerTimeout <= 0 {
		config.WorkerTimeout = defaultWorkerTimeout
	}
}

// Tracker is the main component of Pirsch.
// It provides methods to track requests and store them in a data store.
// It panics in case it cannot store requests into the configured store.
type Tracker struct {
	store                                    Store
	salt                                     string
	hits                                     chan Hit
	worker                                   int
	workerBufferSize                         int
	workerTimeout                            time.Duration
	workerCancel                             context.CancelFunc
	workerDone                               chan bool
	refererDomainBlacklist                   []string
	refererDomainBlacklistIncludesSubdomains bool
}

// NewTracker creates a new tracker for given store, salt and config.
// Pass nil for the config to use the defaults.
// The salt is mandatory.
func NewTracker(store Store, salt string, config *TrackerConfig) *Tracker {
	if config == nil {
		config = &TrackerConfig{}
	}

	config.validate()
	tracker := &Tracker{
		store:                                    store,
		salt:                                     salt,
		hits:                                     make(chan Hit, config.Worker*config.WorkerBufferSize),
		worker:                                   config.Worker,
		workerBufferSize:                         config.WorkerBufferSize,
		workerTimeout:                            config.WorkerTimeout,
		workerDone:                               make(chan bool),
		refererDomainBlacklist:                   config.RefererDomainBlacklist,
		refererDomainBlacklistIncludesSubdomains: config.RefererDomainBlacklistIncludesSubdomains,
	}
	tracker.startWorker()
	return tracker
}

// Hit stores the given request.
// The request might be ignored if it meets certain conditions. The HitOptions, if passed, will overwrite the Tracker configuration.
// The actions performed within this function run in their own goroutine, so you don't need to create one yourself.
func (tracker *Tracker) Hit(r *http.Request, options *HitOptions) {
	go func() {
		if !IgnoreHit(r) {
			if options == nil {
				options = &HitOptions{
					RefererDomainBlacklist:                   tracker.refererDomainBlacklist,
					RefererDomainBlacklistIncludesSubdomains: tracker.refererDomainBlacklistIncludesSubdomains,
				}
			}

			tracker.hits <- HitFromRequest(r, tracker.salt, options)
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
