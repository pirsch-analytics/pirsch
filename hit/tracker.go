package hit

import (
	"context"
	"github.com/pirsch-analytics/pirsch/geodb"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultWorkerBufferSize = 100
	defaultWorkerTimeout    = time.Second * 10
	maxWorkerTimeout        = time.Second * 60
)

var logger = log.New(os.Stdout, "[pirsch] ", log.LstdFlags)

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

	// Sessions enables/disables session tracking.
	// It's enabled by default.
	Sessions bool

	// SessionMaxAge is used to define how long a session runs at maximum.
	// Set to 15 minutes by default.
	SessionMaxAge time.Duration

	// SessionCleanupInterval sets the session cache lifetime.
	// If not passed, the default will be used.
	SessionCleanupInterval time.Duration

	// GeoDB enables/disabled mapping IPs to country codes.
	// Can be set/updated at runtime by calling Tracker.SetGeoDB.
	GeoDB *geodb.GeoDB

	// Logger is the log.Logger used for logging.
	// The default log will be used printing to os.Stdout with "pirsch" in its prefix in case it is not set.
	Logger *log.Logger
}

// The default session configuration is set by the session cache.
// The TrackerConfig just passes on the values and overwrites them if required.
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
		config.Logger = logger
	}
}

// Tracker is the main component of Pirsch.
// It provides methods to track requests and store them in a data store.
// Make sure you call Stop to make sure the hits get stored before shutting down the server.
type Tracker struct {
	store                                     Store
	salt                                      string
	hits                                      chan Hit
	stopped                                   int32
	worker                                    int
	workerBufferSize                          int
	workerTimeout                             time.Duration
	workerCancel                              context.CancelFunc
	workerDone                                chan bool
	referrerDomainBlacklist                   []string
	referrerDomainBlacklistIncludesSubdomains bool
	geoDB                                     *geodb.GeoDB
	geoDBMutex                                sync.RWMutex
	sessionCache                              *sessionCache
	logger                                    *log.Logger
}

// NewTracker creates a new tracker for given store, salt and config.
// Pass nil for the config to use the defaults.
// The salt is mandatory.
func NewTracker(store Store, salt string, config *TrackerConfig) *Tracker {
	if config == nil {
		// the other default values are set by validate
		config = &TrackerConfig{
			Sessions: true,
		}
	}

	config.validate()
	var sessionCache *sessionCache

	if config.Sessions {
		sessionCache = newSessionCache(store, &sessionCacheConfig{
			maxAge:          config.SessionMaxAge,
			cleanupInterval: config.SessionCleanupInterval,
		})
	}

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
		sessionCache: sessionCache,
		geoDB:        config.GeoDB,
		logger:       config.Logger,
	}
	tracker.startWorker()
	return tracker
}

// Hit stores the given request.
// The request might be ignored if it meets certain conditions. The HitOptions, if passed, will overwrite the Tracker configuration.
// It's save (and recommended!) to call this function in its own goroutine.
func (tracker *Tracker) Hit(r *http.Request, options *HitOptions) {
	if atomic.LoadInt32(&tracker.stopped) > 0 {
		return
	}

	if !IgnoreHit(r) {
		if options == nil {
			options = &HitOptions{
				ReferrerDomainBlacklist:                   tracker.referrerDomainBlacklist,
				ReferrerDomainBlacklistIncludesSubdomains: tracker.referrerDomainBlacklistIncludesSubdomains,
			}
		}

		if tracker.geoDB != nil {
			tracker.geoDBMutex.RLock()
			options.geoDB = tracker.geoDB
			tracker.geoDBMutex.RUnlock()
		}

		if tracker.sessionCache != nil {
			options.sessionCache = tracker.sessionCache
		}

		tracker.hits <- HitFromRequest(r, tracker.salt, options)
	}
}

// Flush flushes all hits to store that are currently buffered by the workers.
// Call Tracker.Stop to also save hits that are in the queue.
func (tracker *Tracker) Flush() {
	tracker.stopWorker()
	tracker.startWorker()
}

// Stop flushes and stops all workers.
func (tracker *Tracker) Stop() {
	if atomic.LoadInt32(&tracker.stopped) == 0 {
		atomic.StoreInt32(&tracker.stopped, 1)
		tracker.stopWorker()
		tracker.flushHits()
	}
}

// SetGeoDB sets the GeoDB for the Tracker.
// The call to this function is thread safe to enable live updates of the database.
// Pass nil to disable the feature.
func (tracker *Tracker) SetGeoDB(geoDB *geodb.GeoDB) {
	tracker.geoDBMutex.Lock()
	defer tracker.geoDBMutex.Unlock()
	tracker.geoDB = geoDB
}

func (tracker *Tracker) startWorker() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tracker.workerCancel = cancelFunc

	for i := 0; i < tracker.worker; i++ {
		go tracker.aggregate(ctx)
	}
}

func (tracker *Tracker) stopWorker() {
	tracker.workerCancel()

	for i := 0; i < tracker.worker; i++ {
		<-tracker.workerDone
	}
}

func (tracker *Tracker) flushHits() {
	// this function will make sure all dangling hits will be saved in database before shutdown
	// hits are buffered before saving
	hits := make([]Hit, 0, tracker.workerBufferSize)

	for {
		stop := false

		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)

			if len(hits) == tracker.workerBufferSize {
				tracker.saveHits(hits)
				hits = hits[:0]
			}
		default:
			stop = true
		}

		if stop {
			break
		}
	}

	tracker.saveHits(hits)
}

func (tracker *Tracker) aggregate(ctx context.Context) {
	hits := make([]Hit, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)
			tracker.saveHits(hits)
			hits = hits[:0]
		case <-timer.C:
			tracker.saveHits(hits)
			hits = hits[:0]
		case <-ctx.Done():
			tracker.saveHits(hits)
			tracker.workerDone <- true
			return
		}
	}
}

func (tracker *Tracker) saveHits(hits []Hit) {
	if len(hits) > 0 {
		if err := tracker.store.SaveHits(hits); err != nil {
			tracker.logger.Printf("error saving hits: %s", err)
		}
	}
}
