package tracker

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ip"
	"github.com/pirsch-analytics/pirsch/v4/tracker/session"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"log"
	"net"
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
	defaultMinDelayMS       = 50
	defaultIsBotThreshold   = 5 // also defined for the analyzer.Analyzer
	defaultMaxPageViews     = 150
)

// Config is the optional configuration for the Tracker.
type Config struct {
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
	// This will be set to NewMemCache by default.
	SessionCache session.Cache

	// MaxSessions sets the maximum size for the session cache.
	// If you leave it 0, the default will be used.
	MaxSessions int

	// SessionMaxAge see HitOptions.SessionMaxAge.
	SessionMaxAge time.Duration

	// HeaderParser see HitOptions.HeaderParser.
	// Set it to nil to use the DefaultHeaderParser list.
	HeaderParser []ip.HeaderParser

	// AllowedProxySubnets see HitOptions.AllowedProxySubnets.
	// Set it to nil to allow any IP.
	AllowedProxySubnets []net.IPNet

	// MinDelay see HitOptions.MinDelay.
	MinDelay int64

	// IsBotThreshold see HitOptions.IsBotThreshold.
	// Will be set to 5 by default.
	IsBotThreshold uint8

	// MaxPageViews see HitOptions.MaxPageViews.
	// Will be set to 150 by default.
	MaxPageViews uint16

	// DisableFlaggingBots disables MinDelay and IsBotThreshold (otherwise these would be set to their default values).
	DisableFlaggingBots bool

	// GeoDB enables/disabled mapping IPs to country codes.
	// Can be set/updated at runtime by calling Tracker.SetGeoDB.
	GeoDB *geodb.GeoDB

	// Logger is the log.Logger used for logging.
	// The default log will be used printing to os.Stdout with "pirsch" in its prefix in case it is not set.
	Logger *log.Logger
}

func (config *Config) validate() {
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
		config.MaxPageViews = 0
	} else {
		if config.MinDelay <= 0 {
			config.MinDelay = defaultMinDelayMS
		}

		if config.IsBotThreshold == 0 {
			config.IsBotThreshold = defaultIsBotThreshold
		}

		if config.MaxPageViews == 0 {
			config.MaxPageViews = defaultMaxPageViews
		}
	}

	if config.Logger == nil {
		config.Logger = util.GetDefaultLogger()
	}
}

type data struct {
	session  SessionState
	pageView *model.PageView
	event    *model.Event
	ua       *model.UserAgent
}

// Tracker provides methods to track requests.
// Make sure you call Stop to make sure the hits get stored before shutting down the server.
type Tracker struct {
	store                                     db.Store
	sessionCache                              session.Cache
	salt                                      string
	data                                      chan data
	worker                                    int
	workerBufferSize                          int
	workerTimeout                             time.Duration
	workerCancel                              context.CancelFunc
	workerDone                                chan bool
	referrerDomainBlacklist                   []string
	referrerDomainBlacklistIncludesSubdomains bool
	sessionMaxAge                             time.Duration
	headerParser                              []ip.HeaderParser
	allowedProxySubnets                       []net.IPNet
	minDelay                                  int64
	isBotThreshold                            uint8
	maxPageViews                              uint16
	geoDB                                     *geodb.GeoDB
	geoDBMutex                                sync.RWMutex
	logger                                    *log.Logger
	lock                                      atomic.Bool
}

// NewTracker creates a new tracker for given client, salt and config.
// Pass nil for the config to use the defaults. The salt is mandatory.
// It creates the same amount of workers for both, hits and events.
func NewTracker(client db.Store, salt string, config *Config) *Tracker {
	if config == nil {
		config = &Config{}
	}

	config.validate()

	if config.SessionCache == nil {
		config.SessionCache = session.NewMemCache(client, config.MaxSessions)
	}

	if config.HeaderParser == nil {
		config.HeaderParser = ip.DefaultHeaderParser
	}

	tracker := &Tracker{
		store:                   client,
		sessionCache:            config.SessionCache,
		salt:                    salt,
		data:                    make(chan data, config.Worker*config.WorkerBufferSize),
		worker:                  config.Worker,
		workerBufferSize:        config.WorkerBufferSize,
		workerTimeout:           config.WorkerTimeout,
		workerDone:              make(chan bool),
		referrerDomainBlacklist: config.ReferrerDomainBlacklist,
		referrerDomainBlacklistIncludesSubdomains: config.ReferrerDomainBlacklistIncludesSubdomains,
		sessionMaxAge:       config.SessionMaxAge,
		headerParser:        config.HeaderParser,
		allowedProxySubnets: config.AllowedProxySubnets,
		minDelay:            config.MinDelay,
		isBotThreshold:      config.IsBotThreshold,
		maxPageViews:        config.MaxPageViews,
		geoDB:               config.GeoDB,
		logger:              config.Logger,
	}
	tracker.startWorker()
	return tracker
}

// Hit stores the given request.
// The request might be ignored if it meets certain conditions. The HitOptions, if passed, will overwrite the Tracker configuration.
// It's safe (and recommended!) to call this function in its own goroutine.
func (tracker *Tracker) Hit(r *http.Request, options *HitOptions) {
	if tracker.lock.Load() {
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
				MaxPageViews:                              tracker.maxPageViews,
			}
		}

		if tracker.geoDB != nil {
			tracker.geoDBMutex.RLock()
			options.geoDB = tracker.geoDB
			tracker.geoDBMutex.RUnlock()
		}

		if options.HeaderParser == nil {
			options.HeaderParser = tracker.headerParser
		}

		options.SessionCache = tracker.sessionCache
		options.AllowedProxySubnets = tracker.allowedProxySubnets
		pageView, sessionState, ua := HitFromRequest(r, tracker.salt, options)

		if pageView != nil {
			tracker.data <- data{
				session:  sessionState,
				pageView: pageView,
				ua:       ua,
			}
		}
	}
}

// Event stores the given request as a new event. The event name in the options must be set, or otherwise the request will be ignored.
// The request might be ignored if it meets certain conditions. The HitOptions, if passed, will overwrite the Tracker configuration.
// It's save (and recommended!) to call this function in its own goroutine.
func (tracker *Tracker) Event(r *http.Request, eventOptions EventOptions, options *HitOptions) {
	if tracker.lock.Load() {
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

		if options.HeaderParser == nil {
			options.HeaderParser = tracker.headerParser
		}

		options.SessionCache = tracker.sessionCache
		options.AllowedProxySubnets = tracker.allowedProxySubnets
		options.event = true
		metaKeys, metaValues := eventOptions.getMetaData()
		pageView, sessionState, _ := HitFromRequest(r, tracker.salt, options)

		if pageView != nil {
			tracker.data <- data{
				session: sessionState,
				event: &model.Event{
					ClientID:        pageView.ClientID,
					VisitorID:       pageView.VisitorID,
					Time:            pageView.Time,
					SessionID:       pageView.SessionID,
					DurationSeconds: eventOptions.Duration,
					Name:            strings.TrimSpace(eventOptions.Name),
					MetaKeys:        metaKeys,
					MetaValues:      metaValues,
					Path:            pageView.Path,
					Title:           pageView.Title,
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
				},
			}
		}
	}
}

// ExtendSession looks up and extends the session for given request and client ID (optional).
// This function does not Store a hit or event in database.
func (tracker *Tracker) ExtendSession(r *http.Request, options *HitOptions) {
	if options == nil {
		options = &HitOptions{}
	}

	if options.HeaderParser == nil {
		options.HeaderParser = tracker.headerParser
	}

	ExtendSession(r, tracker.salt, &HitOptions{
		ClientID:            options.ClientID,
		SessionCache:        tracker.sessionCache,
		SessionMaxAge:       tracker.sessionMaxAge,
		HeaderParser:        options.HeaderParser,
		AllowedProxySubnets: tracker.allowedProxySubnets,
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
	if !tracker.lock.Load() {
		tracker.lock.Store(true)
		tracker.stopWorker()
		tracker.flushData()
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

// ClearSessionCache clears the session cache.
func (tracker *Tracker) ClearSessionCache() {
	tracker.sessionCache.Clear()
}

func (tracker *Tracker) startWorker() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tracker.workerCancel = cancelFunc

	for i := 0; i < tracker.worker; i++ {
		go tracker.aggregateData(ctx)
	}
}

func (tracker *Tracker) stopWorker() {
	tracker.workerCancel()

	for i := 0; i < tracker.worker; i++ {
		<-tracker.workerDone
	}
}

func (tracker *Tracker) flushData() {
	sessions := make([]model.Session, 0, tracker.workerBufferSize*2)
	pageViews := make([]model.PageView, 0, tracker.workerBufferSize)
	events := make([]model.Event, 0, tracker.workerBufferSize)
	userAgents := make([]model.UserAgent, 0, tracker.workerBufferSize)

	for {
		stop := false

		select {
		case data := <-tracker.data:
			sessions = append(sessions, data.session.State)

			if data.session.Cancel != nil {
				sessions = append(sessions, data.session.State)
			}

			if data.pageView != nil {
				pageViews = append(pageViews, *data.pageView)
			}

			if data.event != nil {
				events = append(events, *data.event)
			}

			if data.ua != nil {
				userAgents = append(userAgents, *data.ua)
			}

			if len(sessions)+1 >= tracker.workerBufferSize ||
				len(pageViews)+2 >= tracker.workerBufferSize*2 ||
				len(events)+1 >= tracker.workerBufferSize ||
				len(userAgents)+1 >= tracker.workerBufferSize {
				tracker.saveSessions(sessions)
				tracker.savePageViews(pageViews)
				tracker.saveEvents(events)
				tracker.saveUserAgents(userAgents)
				sessions = sessions[:0]
				pageViews = pageViews[:0]
				events = events[:0]
				userAgents = userAgents[:0]
			}
		default:
			stop = true
		}

		if stop {
			break
		}
	}

	tracker.saveSessions(sessions)
	tracker.savePageViews(pageViews)
	tracker.saveEvents(events)
	tracker.saveUserAgents(userAgents)
}

func (tracker *Tracker) aggregateData(ctx context.Context) {
	sessions := make([]model.Session, 0, tracker.workerBufferSize*2)
	pageViews := make([]model.PageView, 0, tracker.workerBufferSize)
	events := make([]model.Event, 0, tracker.workerBufferSize)
	userAgents := make([]model.UserAgent, 0, tracker.workerBufferSize)
	timer := time.NewTimer(tracker.workerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.workerTimeout)

		select {
		case data := <-tracker.data:
			if data.session.Cancel != nil {
				sessions = append(sessions, *data.session.Cancel)
			}

			sessions = append(sessions, data.session.State)

			if data.pageView != nil {
				pageViews = append(pageViews, *data.pageView)
			}

			if data.event != nil {
				events = append(events, *data.event)
			}

			if data.ua != nil {
				userAgents = append(userAgents, *data.ua)
			}

			if len(sessions)+1 >= tracker.workerBufferSize ||
				len(pageViews)+2 >= tracker.workerBufferSize*2 ||
				len(events)+1 >= tracker.workerBufferSize ||
				len(userAgents)+1 >= tracker.workerBufferSize {
				tracker.saveSessions(sessions)
				tracker.savePageViews(pageViews)
				tracker.saveEvents(events)
				tracker.saveUserAgents(userAgents)
				sessions = sessions[:0]
				pageViews = pageViews[:0]
				events = events[:0]
				userAgents = userAgents[:0]
			}
		case <-timer.C:
			tracker.saveSessions(sessions)
			tracker.savePageViews(pageViews)
			tracker.saveEvents(events)
			tracker.saveUserAgents(userAgents)
			sessions = sessions[:0]
			pageViews = pageViews[:0]
			events = events[:0]
			userAgents = userAgents[:0]
		case <-ctx.Done():
			tracker.saveSessions(sessions)
			tracker.savePageViews(pageViews)
			tracker.saveEvents(events)
			tracker.saveUserAgents(userAgents)
			tracker.workerDone <- true
			return
		}
	}
}

func (tracker *Tracker) savePageViews(pageViews []model.PageView) {
	if len(pageViews) > 0 {
		if err := tracker.store.SavePageViews(pageViews); err != nil {
			log.Panicf("error saving page views: %s", err)
		}
	}
}

func (tracker *Tracker) saveSessions(sessions []model.Session) {
	if len(sessions) > 0 {
		if err := tracker.store.SaveSessions(sessions); err != nil {
			log.Panicf("error saving sessions: %s", err)
		}
	}
}

func (tracker *Tracker) saveEvents(events []model.Event) {
	if len(events) > 0 {
		if err := tracker.store.SaveEvents(events); err != nil {
			log.Panicf("error saving events: %s", err)
		}
	}
}

func (tracker *Tracker) saveUserAgents(userAgents []model.UserAgent) {
	if len(userAgents) > 0 {
		if err := tracker.store.SaveUserAgents(userAgents); err != nil {
			tracker.logger.Printf("error saving user agents: %s", err)
		}
	}
}
