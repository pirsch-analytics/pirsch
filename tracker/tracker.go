package tracker

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/tracker_/geodb"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type data struct {
	session       *model.Session
	cancelSession *model.Session
	pageView      *model.PageView
	event         *model.Event
	ua            *model.UserAgent
}

// Tracker tracks page views, events, and updates sessions.
type Tracker struct {
	config     Config
	data       chan data
	cancel     context.CancelFunc
	done       chan bool
	geoDBMutex sync.RWMutex
	stopped    atomic.Bool
}

// NewTracker creates a new tracker for given client, salt and config.
func NewTracker(config Config) *Tracker {
	config.validate()
	tracker := &Tracker{
		config: config,
		data:   make(chan data, config.WorkerBufferSize),
		done:   make(chan bool),
	}
	tracker.startWorker()
	return tracker
}

// PageView tracks a page view.
func (tracker *Tracker) PageView(r *http.Request) {
	if tracker.stopped.Load() {
		return
	}

	/*if !IgnoreHit(r) {
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
	}*/
}

// Event tracks an event.
func (tracker *Tracker) Event(r *http.Request, eventOptions EventOptions) {
	if tracker.stopped.Load() {
		return
	}

	/*if strings.TrimSpace(eventOptions.Name) != "" && !IgnoreHit(r) {
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
				},
			}
		}
	}*/
}

// ExtendSession extends an existing session.
func (tracker *Tracker) ExtendSession(r *http.Request) {
	/*if options == nil {
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
	})*/
}

// Flush flushes all buffered data.
func (tracker *Tracker) Flush() {
	tracker.stopWorker()
	tracker.startWorker()
}

// Stop flushes and stops all workers.
func (tracker *Tracker) Stop() {
	if !tracker.stopped.Load() {
		tracker.stopped.Store(true)
		tracker.stopWorker()
		tracker.flushData()
	}
}

// SetGeoDB updates the GeoDB.
func (tracker *Tracker) SetGeoDB(geoDB *geodb.GeoDB) {
	tracker.geoDBMutex.Lock()
	defer tracker.geoDBMutex.Unlock()
	tracker.config.GeoDB = geoDB
}

func (tracker *Tracker) startWorker() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	tracker.cancel = cancelFunc

	for i := 0; i < tracker.config.Worker; i++ {
		go tracker.aggregateData(ctx)
	}
}

func (tracker *Tracker) stopWorker() {
	tracker.cancel()

	for i := 0; i < tracker.config.Worker; i++ {
		<-tracker.done
	}
}

func (tracker *Tracker) flushData() {
	bufferSize := tracker.config.WorkerBufferSize
	sessions := make([]model.Session, 0, bufferSize*2)
	pageViews := make([]model.PageView, 0, bufferSize)
	events := make([]model.Event, 0, bufferSize)
	userAgents := make([]model.UserAgent, 0, bufferSize)

	for {
		stop := false

		select {
		case data := <-tracker.data:
			if data.cancelSession != nil {
				sessions = append(sessions, *data.cancelSession)
			}

			if data.session != nil {
				sessions = append(sessions, *data.session)
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

			if len(sessions)+2 >= bufferSize*2 || len(pageViews)+1 >= bufferSize ||
				len(events)+1 >= bufferSize || len(userAgents)+1 >= bufferSize {
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
	bufferSize := tracker.config.WorkerBufferSize
	sessions := make([]model.Session, 0, bufferSize*2)
	pageViews := make([]model.PageView, 0, bufferSize)
	events := make([]model.Event, 0, bufferSize)
	userAgents := make([]model.UserAgent, 0, bufferSize)
	timer := time.NewTimer(tracker.config.WorkerTimeout)
	defer timer.Stop()

	for {
		timer.Reset(tracker.config.WorkerTimeout)

		select {
		case data := <-tracker.data:
			if data.cancelSession != nil {
				sessions = append(sessions, *data.cancelSession)
			}

			if data.session != nil {
				sessions = append(sessions, *data.session)
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

			if len(sessions)+2 >= bufferSize*2 || len(pageViews)+1 >= bufferSize ||
				len(events)+1 >= bufferSize || len(userAgents)+1 >= bufferSize {
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
			tracker.done <- true
			return
		}
	}
}

func (tracker *Tracker) savePageViews(pageViews []model.PageView) {
	if len(pageViews) > 0 {
		if err := tracker.config.Store.SavePageViews(pageViews); err != nil {
			log.Panicf("error saving page views: %s", err)
		}
	}
}

func (tracker *Tracker) saveSessions(sessions []model.Session) {
	if len(sessions) > 0 {
		if err := tracker.config.Store.SaveSessions(sessions); err != nil {
			log.Panicf("error saving sessions: %s", err)
		}
	}
}

func (tracker *Tracker) saveEvents(events []model.Event) {
	if len(events) > 0 {
		if err := tracker.config.Store.SaveEvents(events); err != nil {
			log.Panicf("error saving events: %s", err)
		}
	}
}

func (tracker *Tracker) saveUserAgents(userAgents []model.UserAgent) {
	if len(userAgents) > 0 {
		if err := tracker.config.Store.SaveUserAgents(userAgents); err != nil {
			tracker.config.Logger.Printf("error saving user agents: %s", err)
		}
	}
}
