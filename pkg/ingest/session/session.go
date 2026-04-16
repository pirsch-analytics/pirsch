package session

import (
	"strings"
	"time"

	"github.com/dchest/siphash"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

const (
	sessionMaxAge = time.Minute * 30
)

// Session manages visitor sessions.
type Session struct {
	fpKey0, fpKey1 uint64
	fpSalt         string
	cache          Cache
}

// NewSession returns a new Session for the given sipHash parameters and cache.
func NewSession(fpKey0, fpKey1 uint64, fpSalt string, cache Cache) *Session {
	return &Session{
		fpKey0: fpKey0,
		fpKey1: fpKey1,
		fpSalt: fpSalt,
		cache:  cache,
	}
}

// Step implements ingest.PipeStep to process a step.
// It sets the sessions for the visitor and updates it if required.
func (s *Session) Step(request *ingest.Request) (bool, error) {
	// set the visitor ID (fingerprint) first
	request.VisitorID = s.fingerprint(request.UserAgent, request.IP, request.Time)

	// get a lock and data for the session
	m := s.cache.NewMutex(request.ClientID, request.VisitorID)
	m.Lock()
	maxAge := request.Time.Add(-sessionMaxAge)
	session := s.cache.Get(request.ClientID, request.VisitorID, maxAge)

	// if the maximum session age reaches yesterday, we also need to check for the previous day (different fingerprint)
	if session == nil && maxAge.Day() != request.Time.Day() {
		// unlock and try again
		m.Unlock()
		fingerprintYesterday := s.fingerprint(request.UserAgent, request.IP, maxAge)
		m = s.cache.NewMutex(request.ClientID, fingerprintYesterday)
		m.Lock()
		session = s.cache.Get(request.ClientID, fingerprintYesterday, maxAge)

		if session != nil {
			if session.Start.Before(request.Time.Add(-time.Hour * 24)) {
				session = nil
			} else {
				request.VisitorID = fingerprintYesterday
			}
		}
	}

	defer m.Unlock()

	// TODO
	/*if t == sessionUpdate && session == nil {
		return true, nil
	}*/

	var timeOnPage uint32
	var cancelSession *model.Session

	if session == nil || tracker.referrerOrCampaignChanged(r, session, options.Referrer, options.Hostname) {
		session = tracker.newSession(clientID, r, fingerprint, now, ua, ip, options)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	} else {
		if options.MaxPageViews > 0 && session.PageViews >= options.MaxPageViews ||
			options.MaxPageViews == 0 && tracker.config.MaxPageViews > 0 && session.PageViews >= tracker.config.MaxPageViews {
			return nil, nil, 0
		}

		sessionCopy := *session
		cancelSession = &sessionCopy
		cancelSession.Sign = -1
		timeOnPage = tracker.updateSession(t, r, session, now, options.Hostname, options.Path, options.Title, eventNonInteractive)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	}

	// FIXME
	//return session, cancelSession, timeOnPage
	return false, nil
}

/*
func (tracker *Tracker) PageView(r *http.Request, clientID uint64, options Options) bool {
	if tracker.stopped.Load() {
		return false
	}

	now := time.Now().UTC()
	userAgent, ipAddress, ignoreReason := tracker.ignore(r, options)
	options.validate(r)

	if !options.Time.IsZero() {
		now = options.Time
	}

	if ignoreReason == "" {
		session, cancelSession, timeOnPage := tracker.getSession(pageView, clientID, r, now, userAgent, ipAddress, false, options)
		var saveRequest *model.Request

		if session != nil {
			if cancelSession == nil {
				saveRequest = tracker.requestFromSession(session, clientID, ipAddress, userAgent.UserAgent, "")
			}

			tagKeys, tagValues := options.getTags()
			pv := tracker.pageViewFromSession(session, timeOnPage, tagKeys, tagValues)
			tracker.data <- data{
				session:       session,
				cancelSession: cancelSession,
				pageView:      pv,
				request:       saveRequest,
			}
			return true
		}
	} else {
		tracker.captureRequest(now, clientID, r, ipAddress, options.Hostname, options.Path, "", userAgent, ignoreReason)
	}

	return false
}
*/

/*func (s *Session) get(t eventType, clientID uint64, r *http.Request, now time.Time, ua ua.UserAgent, ip string, eventNonInteractive bool, options Options) (*model.Session, *model.Session, uint32) {
	fingerprint := s.fingerprint(s.fpSalt, ua.UserAgent, ip, now)
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
		return nil, nil, 0
	}

	var timeOnPage uint32
	var cancelSession *model.Session

	if session == nil || tracker.referrerOrCampaignChanged(r, session, options.Referrer, options.Hostname) {
		session = tracker.newSession(clientID, r, fingerprint, now, ua, ip, options)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	} else {
		if options.MaxPageViews > 0 && session.PageViews >= options.MaxPageViews ||
			options.MaxPageViews == 0 && tracker.config.MaxPageViews > 0 && session.PageViews >= tracker.config.MaxPageViews {
			return nil, nil, 0
		}

		sessionCopy := *session
		cancelSession = &sessionCopy
		cancelSession.Sign = -1
		timeOnPage = tracker.updateSession(t, r, session, now, options.Hostname, options.Path, options.Title, eventNonInteractive)
		tracker.config.SessionCache.Put(clientID, fingerprint, session)
	}

	return session, cancelSession, timeOnPage
}

func (s *Session) new(clientID uint64, r *http.Request, fingerprint uint64, now time.Time, ua ua.UserAgent, ip string, options Options) *model.Session {
	ua.OS = util.ShortenString(ua.OS, 20)
	ua.OSVersion = util.ShortenString(ua.OSVersion, 20)
	ua.Browser = util.ShortenString(ua.Browser, 20)
	ua.BrowserVersion = util.ShortenString(ua.BrowserVersion, 20)
	lang := util.ShortenString(tracker.getLanguage(r), 10)
	ref, referrerName, referrerIcon := referrer.Get(r, options.Referrer, options.Hostname)
	ref = util.ShortenString(ref, 200)
	referrerName = util.ShortenString(referrerName, 200)
	referrerIcon = util.ShortenString(referrerIcon, 2000)
	screen := tracker.getScreenClass(r, options.ScreenWidth)
	query := r.URL.Query()
	utmSource := strings.TrimSpace(query.Get("utm_source"))
	utmMedium, clickID := tracker.getUTMMedium(r, referrerName)
	utmCampaign := strings.TrimSpace(query.Get("utm_campaign"))
	utmContent := strings.TrimSpace(query.Get("utm_content"))
	utmTerm := strings.TrimSpace(query.Get("utm_term"))
	sourceChannel := channel.Get(ref, referrerName, utmMedium, utmCampaign, utmSource, clickID)
	countryCode, region, city := "", "", ""

	if tracker.config.GeoDB != nil {
		countryCode, region, city = tracker.config.GeoDB.GetLocation(ip)
	}

	hostname := options.Hostname

	if hostname == "" {
		hostname = r.Host
	}

	return &model.Session{
		Sign:           1,
		Version:        1,
		ClientID:       clientID,
		VisitorID:      fingerprint,
		SessionID:      util.RandUint32(),
		Time:           now,
		Start:          now,
		Hostname:       hostname,
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
		ScreenClass:    screen,
		UTMSource:      utmSource,
		UTMMedium:      utmMedium,
		UTMCampaign:    utmCampaign,
		UTMContent:     utmContent,
		UTMTerm:        utmTerm,
		Channel:        sourceChannel,
	}
}

func (s *Session) update(t eventType, r *http.Request, session *model.Session, now time.Time, hostname, path, title string, eventNonInteractive bool) uint32 {
	top := max(now.Unix()-session.Time.Unix(), 0)
	duration := max(now.Unix()-session.Start.Unix(), 0)

	if t == event {
		session.Time = session.Time.Add(time.Millisecond)
		session.IsBounce = eventNonInteractive && session.IsBounce
	} else if t == pageView {
		session.Time = now
		session.IsBounce = session.IsBounce && path == session.ExitPath
		session.PageViews++
	} else if session.Extended < math.MaxUint16-1 {
		session.Time = now
		session.Extended++
	}

	if hostname == "" {
		hostname = r.Host
	}

	session.DurationSeconds = uint32(duration)
	session.Sign = 1
	session.Version++
	session.Hostname = hostname
	session.ExitPath = path
	session.ExitTitle = title
	return uint32(top)
}*/

func (s *Session) fingerprint(ua, ip string, now time.Time) uint64 {
	var sb strings.Builder
	sb.WriteString(ua)
	sb.WriteString(ip)
	sb.WriteString(s.fpSalt)
	sb.WriteString(now.Format("20060102"))
	return siphash.Hash(s.fpKey0, s.fpKey1, []byte(sb.String()))
}
