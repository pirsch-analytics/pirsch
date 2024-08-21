package tracker

import (
	"context"
	"github.com/dchest/siphash"
	"github.com/emvi/iso-639-1"
	"github.com/google/uuid"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ip"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/referrer"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ua"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	minUserAgentLength = 16
	maxUserAgentLength = 500

	minChromeVersion  = 70 // late 2019
	minFirefoxVersion = 68 // mid 2019
	minSafariVersion  = 12 // late 2018
	minOperaVersion   = 65 // late 2019
	minEdgeVersion    = 88 // late 2020
	minIEVersion      = 11 // late 2013

	sessionMaxAge = time.Minute * 30

	pageView = eventType(iota)
	event
	sessionUpdate
)

type eventType int

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
	request       *model.Request
}

// Tracker tracks page views, events, and updates sessions.
type Tracker struct {
	config  Config
	data    chan data
	cancel  context.CancelFunc
	done    chan bool
	stopped atomic.Bool
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
// Returns true if the page view has been accepted and false otherwise.
func (tracker *Tracker) PageView(r *http.Request, clientID uint64, options Options) bool {
	if tracker.stopped.Load() {
		return false
	}

	now := time.Now().UTC()
	userAgent, ipAddress, ignoreReason := tracker.ignore(r)
	options.validate(r)

	if !options.Time.IsZero() {
		now = options.Time
	}

	if ignoreReason == "" {
		session, cancelSession, timeOnPage, bounced := tracker.getSession(pageView, clientID, r, now, userAgent, ipAddress, options)
		var saveRequest *model.Request

		if session != nil {
			if cancelSession == nil {
				saveRequest = tracker.requestFromSession(session, clientID, ipAddress, userAgent.UserAgent, "")
			}

			var pv *model.PageView

			if !bounced {
				tagKeys, tagValues := options.getTags()
				pv = tracker.pageViewFromSession(session, timeOnPage, tagKeys, tagValues)
			}

			tracker.data <- data{
				session:       session,
				cancelSession: cancelSession,
				pageView:      pv,
				request:       saveRequest,
			}
			return true
		}
	} else {
		tracker.captureRequest(now, clientID, r, ipAddress, options.Path, "", userAgent, ignoreReason)
	}

	return false
}

// Event tracks an event.
// Returns true if the event has been accepted and false otherwise.
func (tracker *Tracker) Event(r *http.Request, clientID uint64, eventOptions EventOptions, options Options) bool {
	if tracker.stopped.Load() {
		return false
	}

	now := time.Now().UTC()
	eventOptions.validate()

	if eventOptions.Name != "" {
		userAgent, ipAddress, ignoreReason := tracker.ignore(r)
		options.validate(r)

		if !options.Time.IsZero() {
			now = options.Time
		}

		if ignoreReason == "" {
			session, cancelSession, timeOnPage, _ := tracker.getSession(event, clientID, r, now, userAgent, ipAddress, options)
			var saveRequest *model.Request

			if session != nil {
				if cancelSession == nil {
					saveRequest = tracker.requestFromSession(session, clientID, ipAddress, userAgent.UserAgent, eventOptions.Name)
				}

				var pv *model.PageView
				tagKeys, tagValues := options.getTags()

				if cancelSession == nil || cancelSession.PageViews < session.PageViews {
					pv = tracker.pageViewFromSession(session, timeOnPage, tagKeys, tagValues)
				}

				metaKeys, metaValues := eventOptions.getMetaData(tagKeys, tagValues)
				tracker.data <- data{
					session:       session,
					cancelSession: cancelSession,
					pageView:      pv,
					event:         tracker.eventFromSession(session, clientID, eventOptions.Duration, eventOptions.Name, metaKeys, metaValues),
					request:       saveRequest,
				}
				return true
			}
		} else {
			tracker.captureRequest(now, clientID, r, ipAddress, options.Path, eventOptions.Name, userAgent, ignoreReason)
		}
	}

	return false
}

// ExtendSession extends an existing session.
// Returns true if the session has been extended and false otherwise.
func (tracker *Tracker) ExtendSession(r *http.Request, clientID uint64, options Options) bool {
	if tracker.stopped.Load() {
		return false
	}

	now := time.Now().UTC()
	userAgent, ipAddress, ignoreReason := tracker.ignore(r)

	if ignoreReason == "" {
		options.validate(r)

		if !options.Time.IsZero() {
			now = options.Time
		}

		session, cancelSession, _, _ := tracker.getSession(sessionUpdate, clientID, r, now, userAgent, ipAddress, options)

		if session != nil {
			tracker.data <- data{
				session:       session,
				cancelSession: cancelSession,
			}
			return true
		}
	}

	return false
}

// Flush flushes all buffered data.
func (tracker *Tracker) Flush() {
	tracker.stopWorker()
	tracker.flushData()
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

func (tracker *Tracker) pageViewFromSession(session *model.Session, timeOnPage uint32, tagKeys, tagValues []string) *model.PageView {
	return &model.PageView{
		ClientID:        session.ClientID,
		VisitorID:       session.VisitorID,
		SessionID:       session.SessionID,
		Time:            session.Time,
		DurationSeconds: timeOnPage,
		Path:            session.ExitPath,
		Title:           session.ExitTitle,
		Language:        session.Language,
		CountryCode:     session.CountryCode,
		Region:          session.Region,
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
		ScreenClass:     session.ScreenClass,
		UTMSource:       session.UTMSource,
		UTMMedium:       session.UTMMedium,
		UTMCampaign:     session.UTMCampaign,
		UTMContent:      session.UTMContent,
		UTMTerm:         session.UTMTerm,
		TagKeys:         tagKeys,
		TagValues:       tagValues,
	}
}

func (tracker *Tracker) eventFromSession(session *model.Session, clientID uint64, duration uint32, name string, metaKeys, metaValues []string) *model.Event {
	return &model.Event{
		ClientID:        clientID,
		VisitorID:       session.VisitorID,
		Time:            session.Time,
		SessionID:       session.SessionID,
		DurationSeconds: duration,
		Name:            name,
		MetaKeys:        metaKeys,
		MetaValues:      metaValues,
		Path:            session.ExitPath,
		Title:           session.ExitTitle,
		Language:        session.Language,
		CountryCode:     session.CountryCode,
		Region:          session.Region,
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
		ScreenClass:     session.ScreenClass,
		UTMSource:       session.UTMSource,
		UTMMedium:       session.UTMMedium,
		UTMCampaign:     session.UTMCampaign,
		UTMContent:      session.UTMContent,
		UTMTerm:         session.UTMTerm,
	}
}

func (tracker *Tracker) requestFromSession(session *model.Session, clientID uint64, ipAddress, userAgent, event string) *model.Request {
	logIP := ""

	if tracker.config.LogIP {
		logIP = ipAddress
	}

	return &model.Request{
		ClientID:    clientID,
		VisitorID:   session.VisitorID,
		Time:        session.Time,
		IP:          logIP,
		UserAgent:   userAgent,
		Path:        session.ExitPath,
		Event:       event,
		Referrer:    session.Referrer,
		UTMSource:   session.UTMSource,
		UTMMedium:   session.UTMMedium,
		UTMCampaign: session.UTMCampaign,
	}
}

func (tracker *Tracker) captureRequest(now time.Time, clientID uint64, r *http.Request, ipAddress, path, event string, userAgent ua.UserAgent, botReason string) {
	logIP := ""

	if tracker.config.LogIP {
		logIP = ipAddress
	}

	query := r.URL.Query()
	tracker.data <- data{
		request: &model.Request{
			ClientID:    clientID,
			VisitorID:   tracker.fingerprint(tracker.config.Salt, userAgent.UserAgent, ipAddress, now),
			Time:        now,
			IP:          logIP,
			UserAgent:   r.UserAgent(),
			Path:        path,
			Event:       event,
			Referrer:    r.Referer(),
			UTMSource:   strings.TrimSpace(query.Get("utm_source")),
			UTMMedium:   strings.TrimSpace(query.Get("utm_medium")),
			UTMCampaign: strings.TrimSpace(query.Get("utm_campaign")),
			Bot:         true,
			BotReason:   botReason,
		},
	}
}

func (tracker *Tracker) ignore(r *http.Request) (ua.UserAgent, string, string) {
	ipAddress := ip.Get(r, tracker.config.HeaderParser, tracker.config.AllowedProxySubnets)

	// ignore browsers pre-fetching data
	xPurpose := r.Header.Get("X-Purpose")
	purpose := r.Header.Get("Purpose")

	if r.Header.Get("X-Moz") == "prefetch" ||
		xPurpose == "prefetch" ||
		xPurpose == "preview" ||
		purpose == "prefetch" ||
		purpose == "preview" {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "prefetch"
	}

	// empty User-Agents are usually bots
	rawUserAgent := r.UserAgent()
	userAgent := strings.TrimSpace(strings.ToLower(rawUserAgent))

	if userAgent == "" || len(userAgent) <= minUserAgentLength || len(userAgent) > maxUserAgentLength || util.ContainsNonASCIICharacters(userAgent) {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "ua-chars"
	}

	// ignore User-Agents that are an IP address
	host := rawUserAgent

	if net.ParseIP(host) != nil {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "ua-ip"
	}

	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(rawUserAgent)
	}

	if net.ParseIP(host) != nil {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "ua-ip"
	}

	// filter UUIDs
	if _, err := uuid.Parse(rawUserAgent); err == nil {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "ua-uuid"
	}

	// filter referrer spammers
	if referrer.Ignore(r) {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "referrer"
	}

	userAgentResult := ua.Parse(r)

	if tracker.ignoreBrowserVersion(userAgentResult.Browser, userAgentResult.BrowserVersion) {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "browser"
	}

	// filter for bot keywords
	for _, botUserAgent := range ua.Blacklist {
		if strings.Contains(userAgent, botUserAgent) {
			return ua.UserAgent{
				UserAgent: r.UserAgent(),
			}, ipAddress, "ua-keyword"
		}
	}

	if tracker.config.IPFilter != nil && tracker.config.IPFilter.Ignore(ipAddress) {
		return ua.UserAgent{
			UserAgent: r.UserAgent(),
		}, ipAddress, "ip"
	}

	return userAgentResult, ipAddress, ""
}

func (tracker *Tracker) ignoreBrowserVersion(browser, version string) bool {
	return version != "" &&
		browser == pkg.BrowserChrome && tracker.browserVersionBefore(version, minChromeVersion) ||
		browser == pkg.BrowserFirefox && tracker.browserVersionBefore(version, minFirefoxVersion) ||
		browser == pkg.BrowserSafari && tracker.browserVersionBefore(version, minSafariVersion) ||
		browser == pkg.BrowserOpera && tracker.browserVersionBefore(version, minOperaVersion) ||
		browser == pkg.BrowserEdge && tracker.browserVersionBefore(version, minEdgeVersion) ||
		browser == pkg.BrowserIE && tracker.browserVersionBefore(version, minIEVersion)
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

func (tracker *Tracker) getSession(t eventType, clientID uint64, r *http.Request, now time.Time, ua ua.UserAgent, ip string, options Options) (*model.Session, *model.Session, uint32, bool) {
	fingerprint := tracker.fingerprint(tracker.config.Salt, ua.UserAgent, ip, now)
	m := tracker.config.SessionCache.NewMutex(clientID, fingerprint)
	m.Lock()
	maxAge := now.Add(-sessionMaxAge)
	session := tracker.config.SessionCache.Get(clientID, fingerprint, maxAge)

	// if the maximum session age reaches yesterday, we also need to check for the previous day (different fingerprint)
	if session == nil && maxAge.Day() != now.Day() {
		m.Unlock()
		fingerprintYesterday := tracker.fingerprint(tracker.config.Salt, ua.UserAgent, ip, maxAge)
		m = tracker.config.SessionCache.NewMutex(clientID, fingerprintYesterday)
		m.Lock()
		session = tracker.config.SessionCache.Get(clientID, fingerprintYesterday, maxAge)

		if session != nil {
			if session.Start.Before(now.Add(-time.Hour * 24)) {
				session = nil
			} else {
				fingerprint = fingerprintYesterday
			}
		}
	}

	defer m.Unlock()

	if t == sessionUpdate && session == nil {
		return nil, nil, 0, false
	}

	var timeOnPage uint32
	bounced := false // bounced not including session creation
	var cancelSession *model.Session

	if session == nil || tracker.referrerOrCampaignChanged(r, session, options.Referrer, options.Hostname) {
		session = tracker.newSession(clientID, r, fingerprint, now, ua, ip, options)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	} else {
		if tracker.config.MaxPageViews > 0 && session.PageViews >= tracker.config.MaxPageViews {
			return nil, nil, 0, false
		}

		sessionCopy := *session
		cancelSession = &sessionCopy
		cancelSession.Sign = -1
		timeOnPage, bounced = tracker.updateSession(t, session, now, options.Path, options.Title)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	}

	return session, cancelSession, timeOnPage, bounced
}

func (tracker *Tracker) newSession(clientID uint64, r *http.Request, fingerprint uint64, now time.Time, ua ua.UserAgent, ip string, options Options) *model.Session {
	ua.OS = util.ShortenString(ua.OS, 20)
	ua.OSVersion = util.ShortenString(ua.OSVersion, 20)
	ua.Browser = util.ShortenString(ua.Browser, 20)
	ua.BrowserVersion = util.ShortenString(ua.BrowserVersion, 20)
	lang := util.ShortenString(tracker.getLanguage(r), 10)
	ref, referrerName, referrerIcon := referrer.Get(r, options.Referrer, options.Hostname)
	ref = util.ShortenString(ref, 200)
	referrerName = util.ShortenString(referrerName, 200)
	referrerIcon = util.ShortenString(referrerIcon, 2000)
	screenClass := tracker.getScreenClass(r, options.ScreenWidth)
	query := r.URL.Query()
	utmSource := strings.TrimSpace(query.Get("utm_source"))
	utmMedium := strings.TrimSpace(query.Get("utm_medium"))
	utmCampaign := strings.TrimSpace(query.Get("utm_campaign"))
	utmContent := strings.TrimSpace(query.Get("utm_content"))
	utmTerm := strings.TrimSpace(query.Get("utm_term"))
	countryCode, region, city := "", "", ""

	if tracker.config.GeoDB != nil {
		countryCode, region, city = tracker.config.GeoDB.GetLocation(ip)
	}

	return &model.Session{
		Sign:           1,
		ClientID:       clientID,
		VisitorID:      fingerprint,
		SessionID:      util.RandUint32(),
		Time:           now,
		Start:          now,
		EntryPath:      options.Path,
		ExitPath:       options.Path,
		PageViews:      1,
		IsBounce:       true,
		EntryTitle:     options.Title,
		ExitTitle:      options.Title,
		Language:       lang,
		CountryCode:    countryCode,
		Region:         region,
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
		ScreenClass:    screenClass,
		UTMSource:      utmSource,
		UTMMedium:      utmMedium,
		UTMCampaign:    utmCampaign,
		UTMContent:     utmContent,
		UTMTerm:        utmTerm,
	}
}

func (tracker *Tracker) updateSession(t eventType, session *model.Session, now time.Time, path, title string) (uint32, bool) {
	top := now.Unix() - session.Time.Unix()

	if top < 0 {
		top = 0
	}

	duration := now.Unix() - session.Start.Unix()

	if duration < 0 {
		duration = 0
	}

	if t == event {
		session.Time = session.Time.Add(time.Millisecond)
		session.IsBounce = false

		if session.ExitPath != path {
			session.PageViews++
		}
	} else if t == pageView {
		session.Time = now
		session.IsBounce = session.IsBounce && path == session.ExitPath

		if !session.IsBounce {
			session.PageViews++
		}
	} else if session.Extended < math.MaxUint16-1 {
		session.Time = now
		session.Extended++
	}

	session.DurationSeconds = uint32(duration)
	session.Sign = 1
	session.ExitPath = path
	session.ExitTitle = title
	return uint32(top), session.IsBounce
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

func (tracker *Tracker) getScreenClass(r *http.Request, width uint16) string {
	if width == 0 {
		width = tracker.getScreenWidthFromHeader(r, "Sec-CH-Width")

		if width == 0 {
			width = tracker.getScreenWidthFromHeader(r, "Sec-CH-Viewport-Width")
		}

		if width == 0 {
			return ""
		}
	}

	for _, class := range screenClasses {
		if width >= class.minWidth {
			return class.class
		}
	}

	return "XS"
}

func (tracker *Tracker) getScreenWidthFromHeader(r *http.Request, header string) uint16 {
	h := r.Header.Get(header)

	if h != "" {
		w, err := strconv.Atoi(h)

		if err == nil && w > 0 {
			return uint16(w)
		}
	}

	return 0
}

func (tracker *Tracker) referrerOrCampaignChanged(r *http.Request, session *model.Session, ref, hostname string) bool {
	ref, refName, _ := referrer.Get(r, ref, hostname)

	if ref != "" && ref != session.Referrer || refName != "" && refName != session.ReferrerName {
		return true
	}

	query := r.URL.Query()
	utmSource := strings.TrimSpace(query.Get("utm_source"))
	utmMedium := strings.TrimSpace(query.Get("utm_medium"))
	utmCampaign := strings.TrimSpace(query.Get("utm_campaign"))
	utmContent := strings.TrimSpace(query.Get("utm_content"))
	utmTerm := strings.TrimSpace(query.Get("utm_term"))
	return (utmSource != "" && utmSource != session.UTMSource) ||
		(utmMedium != "" && utmMedium != session.UTMMedium) ||
		(utmCampaign != "" && utmCampaign != session.UTMCampaign) ||
		(utmContent != "" && utmContent != session.UTMContent) ||
		(utmTerm != "" && utmTerm != session.UTMTerm)
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
	requests := make([]model.Request, 0, bufferSize)

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

			if data.request != nil {
				requests = append(requests, *data.request)
			}

			if len(sessions)+2 >= bufferSize*2 ||
				len(pageViews)+1 >= bufferSize ||
				len(events)+1 >= bufferSize ||
				len(requests)+1 >= bufferSize {
				tracker.saveSessions(sessions)
				tracker.savePageViews(pageViews)
				tracker.saveEvents(events)
				tracker.saveRequests(requests)
				sessions = sessions[:0]
				pageViews = pageViews[:0]
				events = events[:0]
				requests = requests[:0]
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
	tracker.saveRequests(requests)
}

func (tracker *Tracker) aggregateData(ctx context.Context) {
	bufferSize := tracker.config.WorkerBufferSize
	sessions := make([]model.Session, 0, bufferSize*2)
	pageViews := make([]model.PageView, 0, bufferSize)
	events := make([]model.Event, 0, bufferSize)
	requests := make([]model.Request, 0, bufferSize)
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

			if data.request != nil {
				requests = append(requests, *data.request)
			}

			if len(sessions)+2 >= bufferSize*2 ||
				len(pageViews)+1 >= bufferSize ||
				len(events)+1 >= bufferSize ||
				len(requests)+1 >= bufferSize {
				tracker.saveSessions(sessions)
				tracker.savePageViews(pageViews)
				tracker.saveEvents(events)
				tracker.saveRequests(requests)
				sessions = sessions[:0]
				pageViews = pageViews[:0]
				events = events[:0]
				requests = requests[:0]
			}
		case <-timer.C:
			tracker.saveSessions(sessions)
			tracker.savePageViews(pageViews)
			tracker.saveEvents(events)
			tracker.saveRequests(requests)
			sessions = sessions[:0]
			pageViews = pageViews[:0]
			events = events[:0]
			requests = requests[:0]
		case <-ctx.Done():
			tracker.saveSessions(sessions)
			tracker.savePageViews(pageViews)
			tracker.saveEvents(events)
			tracker.saveRequests(requests)
			tracker.done <- true
			return
		}
	}
}

func (tracker *Tracker) savePageViews(pageViews []model.PageView) {
	if len(pageViews) > 0 {
		for retries := 5; retries > -1; retries-- {
			if err := tracker.config.Store.SavePageViews(pageViews); err != nil {
				if retries > 0 {
					tracker.config.Logger.Error("error saving page views", "err", err, "retry", retries)
					time.Sleep(time.Second * time.Duration(5-retries) * 10)
				} else {
					log.Panicf("error saving page views: %s", err)
				}
			} else {
				break
			}
		}
	}
}

func (tracker *Tracker) saveSessions(sessions []model.Session) {
	if len(sessions) > 0 {
		for retries := 5; retries > -1; retries-- {
			if err := tracker.config.Store.SaveSessions(sessions); err != nil {
				if retries > 0 {
					tracker.config.Logger.Error("error saving sessions", "err", err, "retry", retries)
					time.Sleep(time.Second * time.Duration(5-retries) * 10)
				} else {
					log.Panicf("error saving sessions: %s", err)
				}
			} else {
				break
			}
		}
	}
}

func (tracker *Tracker) saveEvents(events []model.Event) {
	if len(events) > 0 {
		for retries := 5; retries > -1; retries-- {
			if err := tracker.config.Store.SaveEvents(events); err != nil {
				if retries > 0 {
					tracker.config.Logger.Error("error saving events", "err", err, "retry", retries)
					time.Sleep(time.Second * time.Duration(5-retries) * 10)
				} else {
					log.Panicf("error saving events: %s", err)
				}
			} else {
				break
			}
		}
	}
}

func (tracker *Tracker) saveRequests(requests []model.Request) {
	if len(requests) > 0 {
		for retries := 5; retries > -1; retries-- {
			if err := tracker.config.Store.SaveRequests(requests); err != nil {
				if retries > 0 {
					tracker.config.Logger.Error("error saving requests", "err", err, "retry", retries)
					time.Sleep(time.Second * time.Duration(5-retries) * 10)
				} else {
					log.Panicf("error saving requests: %s", err)
				}
			} else {
				break
			}
		}
	}
}
