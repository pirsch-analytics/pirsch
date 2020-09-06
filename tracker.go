package pirsch

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	defaultWorkerBufferSize = 100
	defaultWorkerTimeout    = time.Second * 10
	maxWorkerTimeout        = time.Second * 60
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
	// If you leave it 0, the default timeout is used, else it is limted to 60 seconds.
	WorkerTimeout time.Duration

	// ReferrerDomainBlacklist see HitOptions.ReferrerDomainBlacklist.
	ReferrerDomainBlacklist []string

	// ReferrerDomainBlacklistIncludesSubdomains see HitOptions.ReferrerDomainBlacklistIncludesSubdomains.
	ReferrerDomainBlacklistIncludesSubdomains bool

	// Logger is the log.Logger used for logging.
	// The default log will be used printing to os.Stdout with "pirsch" in its prefix in case it is not set.
	Logger *log.Logger
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
	} else if config.WorkerTimeout > maxWorkerTimeout {
		config.WorkerTimeout = maxWorkerTimeout
	}

	if config.Logger == nil {
		config.Logger = log.New(os.Stdout, logPrefix, log.LstdFlags)
	}
}

// Tracker is the main component of Pirsch.
// It provides methods to track requests and store them in a data store.
// Make sure you call Stop to make sure the hits get stored before shutting down the server.
type Tracker struct {
	store                                     Store
	salt                                      string
	hits                                      chan Hit
	worker                                    int
	workerBufferSize                          int
	workerTimeout                             time.Duration
	workerCancel                              context.CancelFunc
	workerDone                                chan bool
	referrerDomainBlacklist                   []string
	referrerDomainBlacklistIncludesSubdomains bool
	logger                                    *log.Logger
}

// NewTracker creates a new tracker for given store, salt and config.
// Pass nil for the config to use the defaults.
// The salt is mandatory.
func NewTracker(store Store, salt string, config *TrackerConfig) *Tracker {
	if config == nil {
		config = &TrackerConfig{} // the default values are set by validate
	}

	config.validate()
	tracker := &Tracker{
		store:                   store,
		salt:                    salt,
		hits:                    make(chan Hit, config.Worker*config.WorkerBufferSize),
		worker:                  config.Worker,
		workerBufferSize:        config.WorkerBufferSize,
		workerTimeout:           config.WorkerTimeout,
		workerDone:              make(chan bool),
		referrerDomainBlacklist: config.ReferrerDomainBlacklist,
		referrerDomainBlacklistIncludesSubdomains: config.ReferrerDomainBlacklistIncludesSubdomains,
		logger: config.Logger,
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
					ReferrerDomainBlacklist:                   tracker.referrerDomainBlacklist,
					ReferrerDomainBlacklistIncludesSubdomains: tracker.referrerDomainBlacklistIncludesSubdomains,
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
				if err := tracker.store.SaveHits(hits); err != nil {
					tracker.logger.Printf("error saving hits: %s", err)
				}

				hits = hits[:0]
			}
		case <-time.After(tracker.workerTimeout):
			if len(hits) > 0 {
				if err := tracker.store.SaveHits(hits); err != nil {
					tracker.logger.Printf("error saving hits: %s", err)
				}

				hits = hits[:0]
			}
		case <-ctx.Done():
			if len(hits) > 0 {
				if err := tracker.store.SaveHits(hits); err != nil {
					tracker.logger.Printf("error saving hits: %s", err)
				}
			}

			tracker.workerDone <- true
			return
		}
	}
}
