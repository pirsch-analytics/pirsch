package tracker

import (
	"context"
	"github.com/dchest/siphash"
	iso6391 "github.com/emvi/iso-639-1"
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ip"
	"github.com/pirsch-analytics/pirsch/v4/tracker/referrer"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ua"
	"github.com/pirsch-analytics/pirsch/v4/tracker_/utm"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	minChromeVersion  = 71 // late 2018
	minFirefoxVersion = 63 // late 2018
	minSafariVersion  = 12 // late 2018
	minOperaVersion   = 57 // late 2018
	minEdgeVersion    = 88 // late 2020
	minIEVersion      = 11 // late 2013

	sessionMaxAge = time.Minute * 30
)

type screenClass struct {
	minWidth uint16
	class    string
}

var screenClasses = []screenClass{
	{5120, "UHD 5K"},
	{3840, "UHD 4K"},
	{2560, "WQHD"},
	{1920, "Full HD"},
	{1280, "HD"},
	{1024, "XL"},
	{800, "L"},
	{600, "M"},
	{415, "S"},
}

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
func (tracker *Tracker) PageView(r *http.Request, clientID uint64, options Options) {
	if tracker.stopped.Load() {
		return
	}

	now := time.Now().UTC()
	userAgent, ipAddress, ignore := tracker.ignore(r)

	if !ignore {
		options.validate(r)
		session, cancelSession, timeOnPage := tracker.getSession(clientID, r, now, userAgent, ipAddress, options)
		data := data{
			session: session,
			pageView: &model.PageView{
				ClientID:        session.ClientID,
				VisitorID:       session.VisitorID,
				SessionID:       session.SessionID,
				Time:            session.Time,
				DurationSeconds: timeOnPage,
				Path:            session.ExitPath,
				Title:           session.ExitTitle,
				Language:        session.Language,
				CountryCode:     session.CountryCode,
				City:            session.City,
				Referrer:        session.Referrer,
				ReferrerName:    session.ReferrerName,
				ReferrerIcon:    session.ReferrerIcon,
				OS:              session.OS,
				OSVersion:       session.OSVersion,
				Browser:         session.Browser,
				BrowserVersion:  session.BrowserVersion,
				Desktop:         session.Desktop,
				Mobile:          session.Mobile,
				ScreenWidth:     session.ScreenWidth,
				ScreenHeight:    session.ScreenHeight,
				ScreenClass:     session.ScreenClass,
				UTMSource:       session.UTMSource,
				UTMMedium:       session.UTMMedium,
				UTMCampaign:     session.UTMCampaign,
				UTMContent:      session.UTMContent,
				UTMTerm:         session.UTMTerm,
			},
			ua: &userAgent,
		}

		if cancelSession.Sign != 0 {
			data.cancelSession = cancelSession
		}

		tracker.data <- data
	}
}

// Event tracks an event.
func (tracker *Tracker) Event(r *http.Request, clientID uint64, eventOptions EventOptions, options Options) {
	if tracker.stopped.Load() {
		return
	}

	now := time.Now().UTC()
	eventOptions.validate()

	if eventOptions.Name != "" {
		userAgent, ipAddress, ignore := tracker.ignore(r)

		if !ignore {
			options.validate(r)
			session, cancelSession, _ := tracker.getSession(clientID, r, now, userAgent, ipAddress, options)
			metaKeys, metaValues := eventOptions.getMetaData()
			data := data{
				session: session,
				event: &model.Event{
					ClientID:        clientID,
					VisitorID:       session.VisitorID,
					Time:            session.Time,
					SessionID:       session.SessionID,
					DurationSeconds: eventOptions.Duration,
					Name:            eventOptions.Name,
					MetaKeys:        metaKeys,
					MetaValues:      metaValues,
					Path:            session.ExitPath,
					Title:           session.ExitTitle,
					Language:        session.Language,
					CountryCode:     session.CountryCode,
					City:            session.City,
					Referrer:        session.Referrer,
					ReferrerName:    session.ReferrerName,
					ReferrerIcon:    session.ReferrerIcon,
					OS:              session.OS,
					OSVersion:       session.OSVersion,
					Browser:         session.Browser,
					BrowserVersion:  session.BrowserVersion,
					Desktop:         session.Desktop,
					Mobile:          session.Mobile,
					ScreenWidth:     session.ScreenWidth,
					ScreenHeight:    session.ScreenHeight,
					ScreenClass:     session.ScreenClass,
					UTMSource:       session.UTMSource,
					UTMMedium:       session.UTMMedium,
					UTMCampaign:     session.UTMCampaign,
					UTMContent:      session.UTMContent,
					UTMTerm:         session.UTMTerm,
				},
			}

			if cancelSession.Sign != 0 {
				data.cancelSession = cancelSession
			}

			tracker.data <- data
		}
	}
}

// ExtendSession extends an existing session.
func (tracker *Tracker) ExtendSession(r *http.Request, clientID uint64) {
	/*
		ExtendSession(r, tracker.salt, &HitOptions{
			ClientID:            options.ClientID,
			SessionCache:        tracker.sessionCache,
			SessionMaxAge:       tracker.sessionMaxAge,
			HeaderParser:        options.HeaderParser,
			AllowedProxySubnets: tracker.allowedProxySubnets,
		})
	*/
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

func (tracker *Tracker) ignore(r *http.Request) (model.UserAgent, string, bool) {
	// respect do not track header
	if r.Header.Get("DNT") == "1" {
		return model.UserAgent{}, "", true
	}

	// empty User-Agents are usually bots
	rawUserAgent := r.UserAgent()
	userAgent := strings.TrimSpace(strings.ToLower(rawUserAgent))

	if userAgent == "" || len(userAgent) < 10 || len(userAgent) > 300 || util.ContainsNonASCIICharacters(userAgent) {
		return model.UserAgent{}, "", true
	}

	// ignore browsers pre-fetching data
	xPurpose := r.Header.Get("X-Purpose")
	purpose := r.Header.Get("Purpose")

	if r.Header.Get("X-Moz") == "prefetch" ||
		xPurpose == "prefetch" ||
		xPurpose == "preview" ||
		purpose == "prefetch" ||
		purpose == "preview" {
		return model.UserAgent{}, "", true
	}

	// filter referrer spammers
	if referrer.Ignore(r) {
		return model.UserAgent{}, "", true
	}

	userAgentResult := ua.Parse(rawUserAgent)

	if tracker.ignoreBrowserVersion(userAgentResult.Browser, userAgentResult.BrowserVersion) {
		return model.UserAgent{}, "", true
	}

	// filter for bot keywords
	for _, botUserAgent := range ua.Blacklist {
		if strings.Contains(userAgent, botUserAgent) {
			return model.UserAgent{}, "", true
		}
	}

	// TODO filter by IP address
	ipAddress := ip.Get(r, tracker.config.HeaderParser, tracker.config.AllowedProxySubnets)

	return userAgentResult, ipAddress, false
}

func (tracker *Tracker) ignoreBrowserVersion(browser, version string) bool {
	return version != "" &&
		browser == pirsch.BrowserChrome && tracker.browserVersionBefore(version, minChromeVersion) ||
		browser == pirsch.BrowserFirefox && tracker.browserVersionBefore(version, minFirefoxVersion) ||
		browser == pirsch.BrowserSafari && tracker.browserVersionBefore(version, minSafariVersion) ||
		browser == pirsch.BrowserOpera && tracker.browserVersionBefore(version, minOperaVersion) ||
		browser == pirsch.BrowserEdge && tracker.browserVersionBefore(version, minEdgeVersion) ||
		browser == pirsch.BrowserIE && tracker.browserVersionBefore(version, minIEVersion)
}

func (tracker *Tracker) browserVersionBefore(version string, min int) bool {
	i := strings.Index(version, ".")

	if i >= 0 {
		version = version[:i]
	}

	v, err := strconv.Atoi(version)

	if err != nil {
		return false
	}

	return v < min
}

func (tracker *Tracker) getSession(clientID uint64, r *http.Request, now time.Time, ua model.UserAgent, ip string, options Options) (*model.Session, *model.Session, uint32) {
	fingerprint := tracker.fingerprint(tracker.config.Salt, ua.UserAgent, ip, now)
	m := tracker.config.SessionCache.NewMutex(clientID, fingerprint)
	maxAge := now.Add(-sessionMaxAge)
	session := tracker.config.SessionCache.Get(clientID, fingerprint, maxAge)

	// if the maximum session age reaches yesterday, we also need to check for the previous day (different fingerprint)
	if session == nil && maxAge.Day() != now.Day() {
		fingerprintYesterday := tracker.fingerprint(tracker.config.Salt, ua.UserAgent, ip, maxAge)
		m = tracker.config.SessionCache.NewMutex(clientID, fingerprintYesterday)
		session = tracker.config.SessionCache.Get(clientID, fingerprintYesterday, maxAge)

		if session != nil {
			if session.Start.Before(now.Add(-time.Hour * 24)) {
				session = nil
			} else {
				fingerprint = fingerprintYesterday
			}
		}
	}

	m.Lock()
	defer m.Unlock()
	var timeOnPage uint32
	var cancelSession model.Session

	if session == nil || tracker.referrerOrCampaignChanged(r, session, options.Referrer) {
		session = tracker.newSession(clientID, r, fingerprint, now, ua, ip, options.Path, options.Title, options.Referrer, options.ScreenWidth, options.ScreenHeight)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	} else {
		if tracker.config.IsBotThreshold > 0 && session.IsBot >= tracker.config.IsBotThreshold {
			return nil, nil, 0
		}

		cancelSession = *session
		cancelSession.Sign = -1
		timeOnPage = tracker.updateSession(session, false, now, options.Path, options.Title)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	}

	return session, &cancelSession, timeOnPage
}

func (tracker *Tracker) newSession(clientID uint64, r *http.Request, fingerprint uint64, now time.Time, ua model.UserAgent, ip, path, title, ref string, screenWidth, screenHeight uint16) *model.Session {
	ua.OS = util.ShortenString(ua.OS, 20)
	ua.OSVersion = util.ShortenString(ua.OSVersion, 20)
	ua.Browser = util.ShortenString(ua.Browser, 20)
	ua.BrowserVersion = util.ShortenString(ua.BrowserVersion, 20)
	lang := util.ShortenString(tracker.getLanguage(r), 10)
	ref, referrerName, referrerIcon := referrer.Get(r, ref, tracker.config.ReferrerDomainBlacklist)
	ref = util.ShortenString(ref, 200)
	referrerName = util.ShortenString(referrerName, 200)
	referrerIcon = util.ShortenString(referrerIcon, 2000)
	screenClass := tracker.getScreenClass(screenWidth)
	query := r.URL.Query()
	utmSource := strings.TrimSpace(query.Get("utm_source"))
	utmMedium := strings.TrimSpace(query.Get("utm_medium"))
	utmCampaign := strings.TrimSpace(query.Get("utm_campaign"))
	utmContent := strings.TrimSpace(query.Get("utm_content"))
	utmTerm := strings.TrimSpace(query.Get("utm_term"))
	countryCode, city := "", ""

	if tracker.config.GeoDB != nil {
		countryCode, city = tracker.config.GeoDB.CountryCodeAndCity(ip)
	}

	if screenWidth <= 0 || screenHeight <= 0 {
		screenWidth = 0
		screenHeight = 0
	}

	return &model.Session{
		Sign:           1,
		ClientID:       clientID,
		VisitorID:      fingerprint,
		SessionID:      util.RandUint32(),
		Time:           now,
		Start:          now,
		EntryPath:      path,
		ExitPath:       path,
		PageViews:      1,
		IsBounce:       true,
		EntryTitle:     title,
		ExitTitle:      title,
		Language:       lang,
		CountryCode:    countryCode,
		City:           city,
		Referrer:       ref,
		ReferrerName:   referrerName,
		ReferrerIcon:   referrerIcon,
		OS:             ua.OS,
		OSVersion:      ua.OSVersion,
		Browser:        ua.Browser,
		BrowserVersion: ua.BrowserVersion,
		Desktop:        ua.IsDesktop(),
		Mobile:         ua.IsMobile(),
		ScreenWidth:    screenWidth,
		ScreenHeight:   screenHeight,
		ScreenClass:    screenClass,
		UTMSource:      utmSource,
		UTMMedium:      utmMedium,
		UTMCampaign:    utmCampaign,
		UTMContent:     utmContent,
		UTMTerm:        utmTerm,
	}
}

func (tracker *Tracker) updateSession(session *model.Session, event bool, now time.Time, path, title string) uint32 {
	if tracker.config.MinDelay > 0 && now.UnixMilli()-session.Time.UnixMilli() < tracker.config.MinDelay {
		session.IsBot++
	}

	top := now.Unix() - session.Time.Unix()

	if top < 0 {
		top = 0
	}

	duration := now.Unix() - session.Start.Unix()

	if duration < 0 {
		duration = 0
	}

	session.DurationSeconds = uint32(duration)
	session.Sign = 1
	session.Time = now

	if event {
		session.IsBounce = false
	} else {
		session.IsBounce = session.IsBounce && path == session.ExitPath
		session.ExitPath = path
		session.ExitTitle = title
		session.PageViews++
	}

	return uint32(top)
}

func (tracker *Tracker) getLanguage(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")

	if lang != "" {
		left, _, _ := strings.Cut(lang, ";")
		left, _, _ = strings.Cut(left, ",")
		left, _, _ = strings.Cut(left, "-")
		code := strings.ToLower(strings.TrimSpace(left))

		if iso6391.ValidCode(code) {
			return code
		}
	}

	return ""
}

func (tracker *Tracker) getScreenClass(width uint16) string {
	if width <= 0 {
		return ""
	}

	for _, class := range screenClasses {
		if width >= class.minWidth {
			return class.class
		}
	}

	return "XS"
}

func (tracker *Tracker) referrerOrCampaignChanged(r *http.Request, session *model.Session, ref string) bool {
	ref, _, _ = referrer.Get(r, ref, tracker.config.ReferrerDomainBlacklist)

	if ref != "" && ref != session.Referrer {
		return true
	}

	utmParams := utm.Get(r)
	return (utmParams.Source != "" && utmParams.Source != session.UTMSource) ||
		(utmParams.Medium != "" && utmParams.Medium != session.UTMMedium) ||
		(utmParams.Campaign != "" && utmParams.Campaign != session.UTMCampaign) ||
		(utmParams.Content != "" && utmParams.Content != session.UTMContent) ||
		(utmParams.Term != "" && utmParams.Term != session.UTMTerm)
}

func (tracker *Tracker) fingerprint(salt, ua, ip string, now time.Time) uint64 {
	var sb strings.Builder
	sb.WriteString(ua)
	sb.WriteString(ip)
	sb.WriteString(salt)
	sb.WriteString(now.Format("20060102"))
	return siphash.Hash(tracker.config.FingerprintKey0, tracker.config.FingerprintKey1, []byte(sb.String()))
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
