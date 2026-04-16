package session

import (
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/dchest/siphash"
	"github.com/pirsch-analytics/pirsch/v7/_pkg/tracker/channel"
	"github.com/pirsch-analytics/pirsch/v7/_pkg/tracker/referrer"
	"github.com/pirsch-analytics/pirsch/v7/_pkg/util"
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
	maxPageViews   uint16
}

// NewSession returns a new Session for the given sipHash parameters, cache, and options.
func NewSession(fpKey0, fpKey1 uint64, fpSalt string, cache Cache, maxPageViews uint16) *Session {
	return &Session{
		fpKey0:       fpKey0,
		fpKey1:       fpKey1,
		fpSalt:       fpSalt,
		cache:        cache,
		maxPageViews: maxPageViews,
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

	if session == nil || s.referrerOrCampaignChanged(request, session) {
		session = s.new(request)
		s.cache.Put(request.ClientID, request.VisitorID, session)
	} else {
		if s.maxPageViews > 0 && session.PageViews >= s.maxPageViews ||
			s.maxPageViews == 0 && s.maxPageViews > 0 && session.PageViews >= s.maxPageViews {
			return true, nil
		}

		sessionCopy := *session
		cancelSession = &sessionCopy
		cancelSession.Sign = -1
		timeOnPage = s.update(request, session)
		s.cache.Put(request.ClientID, request.VisitorID, session)
	}

	// FIXME
	//return session, cancelSession, timeOnPage
	return false, nil
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
		Data: model.Data{
			ClientID:       clientID,
			VisitorID:      fingerprint,
			SessionID:      util.RandUint32(),
			Time:           now,
			Hostname:       hostname,
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
		},
		Sign:       1,
		Version:    1,
		Start:      now,
		EntryPath:  options.Path,
		ExitPath:   options.Path,
		PageViews:  1,
		IsBounce:   true,
		EntryTitle: options.Title,
		ExitTitle:  options.Title,
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
}

func (s *Session) referrerOrCampaignChanged(request *ingest.Request, session *model.Session) bool {
	// TODO
	/*if ref != "" && ref != session.Referrer || refName != "" && refName != session.ReferrerName {
		return true
	}

	return (utmSource != "" && utmSource != session.UTMSource) ||
		(utmMedium != "" && utmMedium != session.UTMMedium) ||
		(utmCampaign != "" && utmCampaign != session.UTMCampaign) ||
		(utmContent != "" && utmContent != session.UTMContent) ||
		(utmTerm != "" && utmTerm != session.UTMTerm)*/
	return false
}

func (s *Session) fingerprint(ua, ip string, now time.Time) uint64 {
	var sb strings.Builder
	sb.WriteString(ua)
	sb.WriteString(ip)
	sb.WriteString(s.fpSalt)
	sb.WriteString(now.Format("20060102"))
	return siphash.Hash(s.fpKey0, s.fpKey1, []byte(sb.String()))
}
