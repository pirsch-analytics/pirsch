package pirsch

import (
	"context"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultWorkerBufferSize = 100
	defaultWorkerTimeout    = time.Second * 3
	maxWorkerTimeout        = time.Second * 60
	defaultMinDelayMS       = 200
	defaultIsBotThreshold   = 5
)

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

	// MinDelay see HitOptions.MinDelay.
	MinDelay int64

	// IsBotThreshold see HitOptions.IsBotThreshold.
	// Will be set to 5 by default.
	IsBotThreshold uint8

	// DisableFlaggingBots disables MinDelay and IsBotThreshold (otherwise these would be set to their default values).
	DisableFlaggingBots bool

	// GeoDB enables/disabled mapping IPs to country codes.
	// Can be set/updated at runtime by calling Tracker.SetGeoDB.
	GeoDB *GeoDB

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

	if config.SessionMaxAge < 0 {
		config.SessionMaxAge = 0
	}

	if config.DisableFlaggingBots {
		config.MinDelay = 0
		config.IsBotThreshold = 0
	} else {
		if config.MinDelay <= 0 {
			config.MinDelay = defaultMinDelayMS
		}

		if config.IsBotThreshold == 0 {
			config.IsBotThreshold = defaultIsBotThreshold
		}
	}

	if config.Logger == nil {
		config.Logger = logger
	}
}

// Tracker provides methods to track requests.
// Make sure you call Stop to make sure the hits get stored before shutting down the server.
type Tracker struct {
	store                                     Store
	sessionCache                              SessionCache
	salt                                      string
	pageViews                                 chan PageView
	sessions                                  chan SessionState
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
	minDelay                                  int64
	isBotThreshold                            uint8
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
		pageViews:               make(chan PageView, config.Worker*config.WorkerBufferSize),
		sessions:                make(chan SessionState, config.Worker*config.WorkerBufferSize),
		events:                  make(chan Event, config.Worker*config.WorkerBufferSize),
		userAgents:              make(chan UserAgent, config.Worker*config.WorkerBufferSize),
		worker:                  config.Worker,
		workerBufferSize:        config.WorkerBufferSize,
		workerTimeout:           config.WorkerTimeout,
		workerDone:              make(chan bool),
		referrerDomainBlacklist: config.ReferrerDomainBlacklist,
		referrerDomainBlacklistIncludesSubdomains: config.ReferrerDomainBlacklistIncludesSubdomains,
		sessionMaxAge:  config.SessionMaxAge,
		minDelay:       config.MinDelay,
		isBotThreshold: config.IsBotThreshold,
		geoDB:          config.GeoDB,
		logger:         config.Logger,
	}
	tracker.startWorker()
	return tracker
}

// Hit stores the given request.
// The request might be ignored if it meets certain conditions. The HitOptions, if passed, will overwrite the Tracker configuration.
// It's safe (and recommended!) to call this function in its own goroutine.
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
				MinDelay:                                  tracker.minDelay,
				IsBotThreshold:                            tracker.isBotThreshold,
			}
		}

		if tracker.geoDB != nil {
			tracker.geoDBMutex.RLock()
			options.geoDB = tracker.geoDB
			tracker.geoDBMutex.RUnlock()
		}

		options.SessionCache = tracker.sessionCache
		pageView, sessionState, ua := HitFromRequest(r, tracker.salt, options)

		if pageView != nil {
			tracker.pageViews <- *pageView
			tracker.sessions <- sessionState
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
			// HitOptions.MinDelay and HitOptions.IsBotThreshold are ignored for events
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
		options.event = true
		metaKeys, metaValues := eventOptions.getMetaData()
		pageView, sessionState, _ := HitFromRequest(r, tracker.salt, options)

		if pageView != nil {
			tracker.sessions <- sessionState
			tracker.events <- Event{
				ClientID:        pageView.ClientID,
				VisitorID:       pageView.VisitorID,
				Time:            pageView.Time,
				SessionID:       pageView.SessionID,
				DurationSeconds: eventOptions.Duration,
				Name:            strings.TrimSpace(eventOptions.Name),
				MetaKeys:        metaKeys,
				MetaValues:      metaValues,
				Path:            pageView.Path,
				Title:           options.Title,
				Language:        pageView.Language,
				CountryCode:     pageView.CountryCode,
				City:            pageView.City,
				Referrer:        pageView.Referrer,
				ReferrerName:    pageView.ReferrerName,
				ReferrerIcon:    pageView.ReferrerIcon,
				OS:              pageView.OS,
				OSVersion:       pageView.OSVersion,
				Browser:         pageView.Browser,
				BrowserVersion:  pageView.BrowserVersion,
				Desktop:         pageView.Desktop,
				Mobile:          pageView.Mobile,
				ScreenWidth:     pageView.ScreenWidth,
				ScreenHeight:    pageView.ScreenHeight,
				ScreenClass:     pageView.ScreenClass,
				UTMSource:       pageView.UTMSource,
				UTMMedium:       pageView.UTMMedium,
				UTMCampaign:     pageView.UTMCampaign,
				UTMContent:      pageView.UTMContent,
				UTMTerm:         pageView.UTMTerm,
			}
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
		tracker.flushPageViews()
		tracker.flushSessions()
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
		go tracker.aggregatePageViews(ctx)
		go tracker.aggregateSessions(ctx)
		go tracker.aggregateEvents(ctx)
		go tracker.aggregateUserAgents(ctx)
	}
}

func (tracker *Tracker) stopWorker() {
	tracker.workerCancel()

	for i := 0; i < tracker.worker*4; i++ {
		<-tracker.workerDone
	}
}

func (tracker *Tracker) flushPageViews() {
	pageViews := make([]PageView, 0, tracker.workerBufferSize)

	for {
		stop := false

		select {
		case pageView := <-tracker.pageViews:
			pageViews = append(pageViews, pageView)

			if len(pageViews) >= tracker.workerBufferSize {
				tracker.savePageViews(pageViews)
				pageViews = pageViews[:0]
			}
		default:
			stop = true
		}

		if stop {
			break
		}
	}

	tracker.savePageViews(pageViews)
}

func (tracker *Tracker) aggregatePageViews(ctx context.Context) {
	pageViews := make([]PageView, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case pageView := <-tracker.pageViews:
			pageViews = append(pageViews, pageView)

			if len(pageViews) >= tracker.workerBufferSize {
				tracker.savePageViews(pageViews)
				pageViews = pageViews[:0]
			}
		case <-timer.C:
			tracker.savePageViews(pageViews)
			pageViews = pageViews[:0]
		case <-ctx.Done():
			tracker.savePageViews(pageViews)
			tracker.workerDone <- true
			return
		}
	}
}

func (tracker *Tracker) savePageViews(pageViews []PageView) {
	if len(pageViews) > 0 {
		if err := tracker.store.SavePageViews(pageViews); err != nil {
			tracker.logger.Printf("error saving page views: %s", err)
		}
	}
}

func (tracker *Tracker) flushSessions() {
	sessions := make([]Session, 0, tracker.workerBufferSize)

	for {
		stop := false

		select {
		case session := <-tracker.sessions:
			if len(sessions)+2 >= tracker.workerBufferSize {
				tracker.saveSessions(sessions)
				sessions = sessions[:0]
			}

			sessions = append(sessions, session.State)

			if session.Cancel != nil {
				sessions = append(sessions, *session.Cancel)
			}
		default:
			stop = true
		}

		if stop {
			break
		}
	}

	tracker.saveSessions(sessions)
}

func (tracker *Tracker) aggregateSessions(ctx context.Context) {
	sessions := make([]Session, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case session := <-tracker.sessions:
			if len(sessions)+2 >= tracker.workerBufferSize {
				tracker.saveSessions(sessions)
				sessions = sessions[:0]
			}

			if session.Cancel != nil {
				sessions = append(sessions, *session.Cancel)
			}

			sessions = append(sessions, session.State)
		case <-timer.C:
			tracker.saveSessions(sessions)
			sessions = sessions[:0]
		case <-ctx.Done():
			tracker.saveSessions(sessions)
			tracker.workerDone <- true
			return
		}
	}
}

func (tracker *Tracker) saveSessions(sessions []Session) {
	if len(sessions) > 0 {
		if err := tracker.store.SaveSessions(sessions); err != nil {
			tracker.logger.Printf("error saving sessions: %s", err)
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

			if len(events) >= tracker.workerBufferSize {
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

			if len(events) >= tracker.workerBufferSize {
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
