package pirsch

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
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
	// Worker sets the number of workers that are used to client hits.
	// Must be greater or equal to 1.
	Worker int

	// WorkerBufferSize is the size of the buffer used to client hits.
	// Must be greater than 0. The hits are stored in batch when the buffer is full.
	WorkerBufferSize int

	// WorkerTimeout sets the timeout used to client hits.
	// This is used to allow the workers to client hits even if the buffer is not full yet.
	// It's recommended to set this to a few seconds.
	// If you leave it 0, the default timeout is used, else it is limited to 60 seconds.
	WorkerTimeout time.Duration

	// ReferrerDomainBlacklist see HitOptions.ReferrerDomainBlacklist.
	ReferrerDomainBlacklist []string

	// ReferrerDomainBlacklistIncludesSubdomains see HitOptions.ReferrerDomainBlacklistIncludesSubdomains.
	ReferrerDomainBlacklistIncludesSubdomains bool

	// SessionCache is the session cache implementation to be used.
	// This will be set to NewSessionCacheMem by default.
	SessionCache SessionCache

	// MaxSessions sets the maximum size for the session cache.
	// If you leave it 0, the default will be used.
	MaxSessions int

	// SessionMaxAge see HitOptions.SessionMaxAge.
	SessionMaxAge time.Duration

	// GeoDB enables/disabled mapping IPs to country codes.
	// Can be set/updated at runtime by calling Tracker.SetGeoDB.
	GeoDB *GeoDB

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

	if config.SessionMaxAge < 0 {
		config.SessionMaxAge = 0
	}

	if config.Logger == nil {
		config.Logger = logger
	}
}

// Tracker provides methods to track requests (hits and events).
// Make sure you call Stop to make sure the hits get stored before shutting down the server.
type Tracker struct {
	store                                     Store
	sessionCache                              SessionCache
	salt                                      string
	hits                                      chan Hit
	events                                    chan Event
	userAgents                                chan UserAgent
	stopped                                   int32
	worker                                    int
	workerBufferSize                          int
	workerTimeout                             time.Duration
	workerCancel                              context.CancelFunc
	workerDone                                chan bool
	referrerDomainBlacklist                   []string
	referrerDomainBlacklistIncludesSubdomains bool
	sessionMaxAge                             time.Duration
	geoDB                                     *GeoDB
	geoDBMutex                                sync.RWMutex
	logger                                    *log.Logger
}

// NewTracker creates a new tracker for given client, salt and config.
// Pass nil for the config to use the defaults. The salt is mandatory.
// It creates the same amount of workers for both, hits and events.
func NewTracker(client Store, salt string, config *TrackerConfig) *Tracker {
	if config == nil {
		config = &TrackerConfig{}
	}

	config.validate()

	if config.SessionCache == nil {
		config.SessionCache = NewSessionCacheMem(client, config.MaxSessions)
	}

	tracker := &Tracker{
		store:                   client,
		sessionCache:            config.SessionCache,
		salt:                    salt,
		hits:                    make(chan Hit, config.Worker*config.WorkerBufferSize),
		events:                  make(chan Event, config.Worker*config.WorkerBufferSize),
		userAgents:              make(chan UserAgent, config.Worker*config.WorkerBufferSize),
		worker:                  config.Worker,
		workerBufferSize:        config.WorkerBufferSize,
		workerTimeout:           config.WorkerTimeout,
		workerDone:              make(chan bool),
		referrerDomainBlacklist: config.ReferrerDomainBlacklist,
		referrerDomainBlacklistIncludesSubdomains: config.ReferrerDomainBlacklistIncludesSubdomains,
		sessionMaxAge: config.SessionMaxAge,
		geoDB:         config.GeoDB,
		logger:        config.Logger,
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
				SessionMaxAge:                             tracker.sessionMaxAge,
			}
		}

		if tracker.geoDB != nil {
			tracker.geoDBMutex.RLock()
			options.geoDB = tracker.geoDB
			tracker.geoDBMutex.RUnlock()
		}

		options.SessionCache = tracker.sessionCache
		hit, ua := HitFromRequest(r, tracker.salt, options)

		if hit != nil {
			tracker.hits <- *hit
		}

		if ua != nil {
			tracker.userAgents <- *ua
		}
	}
}

// Event stores the given request as a new event. The event name in the options must be set, or otherwise the request will be ignored.
// The request might be ignored if it meets certain conditions. The HitOptions, if passed, will overwrite the Tracker configuration.
// It's save (and recommended!) to call this function in its own goroutine.
func (tracker *Tracker) Event(r *http.Request, eventOptions EventOptions, options *HitOptions) {
	if atomic.LoadInt32(&tracker.stopped) > 0 {
		return
	}

	if strings.TrimSpace(eventOptions.Name) != "" && !IgnoreHit(r) {
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

		options.SessionCache = tracker.sessionCache
		metaKeys, metaValues := eventOptions.getMetaData()
		hit, ua := HitFromRequest(r, tracker.salt, options)

		if hit != nil {
			hit.PageViews = 1
			tracker.events <- Event{
				Hit:             *hit,
				Name:            strings.TrimSpace(eventOptions.Name),
				DurationSeconds: eventOptions.Duration,
				MetaKeys:        metaKeys,
				MetaValues:      metaValues,
			}
		}

		if ua != nil {
			tracker.userAgents <- *ua
		}
	}
}

// ExtendSession looks up and extends the session for given request and client ID (optional).
// This function does not store a hit or event in database.
func (tracker *Tracker) ExtendSession(r *http.Request, clientID uint64) {
	ExtendSession(r, tracker.salt, &HitOptions{
		ClientID:      clientID,
		SessionCache:  tracker.sessionCache,
		SessionMaxAge: tracker.sessionMaxAge,
	})
}

// Flush flushes all hits to client that are currently buffered by the workers.
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
		tracker.flushEvents()
		tracker.flushUserAgents()
	}
}

// SetGeoDB sets the GeoDB for the Tracker.
// The call to this function is thread safe to enable live updates of the database.
// Pass nil to disable the feature.
func (tracker *Tracker) SetGeoDB(geoDB *GeoDB) {
	tracker.geoDBMutex.Lock()
	defer tracker.geoDBMutex.Unlock()
	tracker.geoDB = geoDB
}

// ClearSessionCache clears the session cache.
func (tracker *Tracker) ClearSessionCache() {
	tracker.sessionCache.Clear()
}

func (tracker *Tracker) startWorker() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tracker.workerCancel = cancelFunc

	for i := 0; i < tracker.worker; i++ {
		go tracker.aggregateHits(ctx)
		go tracker.aggregateEvents(ctx)
		go tracker.aggregateUserAgents(ctx)
	}
}

func (tracker *Tracker) stopWorker() {
	tracker.workerCancel()

	for i := 0; i < tracker.worker*3; i++ {
		<-tracker.workerDone
	}
}

// This function will make sure all dangling hits will be saved in database before shutdown.
// Hits are buffered before saving.
// The same method is used for events and user agents.
func (tracker *Tracker) flushHits() {
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

func (tracker *Tracker) aggregateHits(ctx context.Context) {
	hits := make([]Hit, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case hit := <-tracker.hits:
			hits = append(hits, hit)

			if len(hits) == tracker.workerBufferSize {
				tracker.saveHits(hits)
				hits = hits[:0]
			}
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

func (tracker *Tracker) flushEvents() {
	events := make([]Event, 0, tracker.workerBufferSize)

	for {
		stop := false

		select {
		case event := <-tracker.events:
			events = append(events, event)

			if len(events) == tracker.workerBufferSize {
				tracker.saveEvents(events)
				events = events[:0]
			}
		default:
			stop = true
		}

		if stop {
			break
		}
	}

	tracker.saveEvents(events)
}

func (tracker *Tracker) aggregateEvents(ctx context.Context) {
	events := make([]Event, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case event := <-tracker.events:
			events = append(events, event)

			if len(events) == tracker.workerBufferSize {
				tracker.saveEvents(events)
				events = events[:0]
			}
		case <-timer.C:
			tracker.saveEvents(events)
			events = events[:0]
		case <-ctx.Done():
			tracker.saveEvents(events)
			tracker.workerDone <- true
			return
		}
	}
}

func (tracker *Tracker) saveEvents(events []Event) {
	if len(events) > 0 {
		if err := tracker.store.SaveEvents(events); err != nil {
			tracker.logger.Printf("error saving events: %s", err)
		}
	}
}

func (tracker *Tracker) flushUserAgents() {
	userAgents := make([]UserAgent, 0, tracker.workerBufferSize)

	for {
		stop := false

		select {
		case ua := <-tracker.userAgents:
			userAgents = append(userAgents, ua)

			if len(userAgents) == tracker.workerBufferSize {
				tracker.saveUserAgents(userAgents)
				userAgents = userAgents[:0]
			}
		default:
			stop = true
		}

		if stop {
			break
		}
	}

	tracker.saveUserAgents(userAgents)
}

func (tracker *Tracker) aggregateUserAgents(ctx context.Context) {
	userAgents := make([]UserAgent, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case ua := <-tracker.userAgents:
			userAgents = append(userAgents, ua)

			if len(userAgents) == tracker.workerBufferSize {
				tracker.saveUserAgents(userAgents)
				userAgents = userAgents[:0]
			}
		case <-timer.C:
			tracker.saveUserAgents(userAgents)
			userAgents = userAgents[:0]
		case <-ctx.Done():
			tracker.saveUserAgents(userAgents)
			tracker.workerDone <- true
			return
		}
	}
}

func (tracker *Tracker) saveUserAgents(userAgents []UserAgent) {
	if len(userAgents) > 0 {
		if err := tracker.store.SaveUserAgents(userAgents); err != nil {
			tracker.logger.Printf("error saving user agents: %s", err)
		}
	}
}
