package session

import (
	"math"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/dchest/siphash"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

const (
	sessionTimeout = time.Minute * 30
	sessionMaxAge  = time.Hour * 24
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
	maxAge := request.Time.Add(-sessionTimeout)
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
			if session.Start.Before(request.Time.Add(-sessionMaxAge)) {
				session = nil
			} else {
				request.VisitorID = fingerprintYesterday
			}
		}
	}

	defer m.Unlock()

	// cancel early if we only update the session
	if request.UpdateSession {
		if session != nil {
			s.update(request, session)
			s.cache.Put(request.ClientID, request.VisitorID, session)
		}

		return true, nil
	}

	var cancelSession *model.Session

	if session == nil || s.referrerOrCampaignChanged(request, session) {
		session = s.new(request)
		s.cache.Put(request.ClientID, request.VisitorID, session)
	} else {
		// cancel if the maximum number of page views has been reached
		if s.maxPageViews > 0 && session.PageViews >= s.maxPageViews ||
			s.maxPageViews == 0 && s.maxPageViews > 0 && session.PageViews >= s.maxPageViews {
			return true, nil
		}

		cancelSession = new(*session)
		cancelSession.Sign = -1
		s.update(request, session)
		s.cache.Put(request.ClientID, request.VisitorID, session)
	}

	// set/update the session
	request.Session = session
	request.CancelSession = cancelSession
	return false, nil
}

func (s *Session) new(request *ingest.Request) *model.Session {
	request.SessionID = rand.Uint32()
	return &model.Session{
		Data: model.Data{
			ClientID:       request.ClientID,
			VisitorID:      request.VisitorID,
			SessionID:      request.SessionID,
			Time:           request.Time,
			Hostname:       request.Hostname,
			Language:       request.Language,
			CountryCode:    request.CountryCode,
			Region:         request.Region,
			City:           request.City,
			Referrer:       request.Referrer,
			ReferrerName:   request.ReferrerName,
			ReferrerIcon:   request.ReferrerIcon,
			OS:             request.OS,
			OSVersion:      request.OSVersion,
			Browser:        request.Browser,
			BrowserVersion: request.BrowserVersion,
			Desktop:        request.Desktop,
			Mobile:         request.Mobile,
			ScreenClass:    request.ScreenClass,
			UTMSource:      request.UTMSource,
			UTMMedium:      request.UTMMedium,
			UTMCampaign:    request.UTMCampaign,
			UTMContent:     request.UTMContent,
			UTMTerm:        request.UTMTerm,
			Channel:        request.Channel,
		},
		Sign:       1,
		Version:    1,
		Start:      request.Time,
		EntryPath:  request.Path,
		ExitPath:   request.Path,
		PageViews:  1,
		IsBounce:   true,
		EntryTitle: request.Title,
		ExitTitle:  request.Title,
	}
}

func (s *Session) update(request *ingest.Request, session *model.Session) {
	top := max(request.Time.Unix()-session.Time.Unix(), 0)
	duration := max(request.Time.Unix()-session.Start.Unix(), 0)

	if request.UpdateSession {
		session.Time = request.Time

		if session.Extended < math.MaxUint16-1 {
			session.Extended++
		}
	} else if request.EventName != "" {
		session.Time = session.Time.Add(time.Millisecond)
		session.IsBounce = request.EventNonInteractive && session.IsBounce
	} else {
		session.Time = request.Time
		session.IsBounce = session.IsBounce && request.Path == session.ExitPath
		session.PageViews++
	}

	// update the session
	session.DurationSeconds = uint32(duration)
	session.Sign = 1
	session.Version++
	session.Hostname = request.Hostname
	session.ExitPath = request.Path
	session.ExitTitle = request.Title

	// update the page view/event using the session data, so that it stays consistent across requests
	request.SessionID = session.SessionID
	request.DurationSeconds = uint64(top)
	request.Language = session.Language
	request.CountryCode = session.CountryCode
	request.Region = session.Region
	request.City = session.City
	request.Referrer = session.Referrer
	request.ReferrerName = session.ReferrerName
	request.ReferrerIcon = session.ReferrerIcon
	request.OS = session.OS
	request.OSVersion = session.OSVersion
	request.Browser = session.Browser
	request.BrowserVersion = session.BrowserVersion
	request.Desktop = session.Desktop
	request.Mobile = session.Mobile
	request.ScreenClass = session.ScreenClass
	request.UTMSource = session.UTMSource
	request.UTMMedium = session.UTMMedium
	request.UTMCampaign = session.UTMCampaign
	request.UTMContent = session.UTMContent
	request.UTMTerm = session.UTMTerm
	request.Channel = session.Channel
}

func (s *Session) referrerOrCampaignChanged(request *ingest.Request, session *model.Session) bool {
	if request.Referrer != "" && request.Referrer != session.Referrer ||
		request.ReferrerName != "" && request.ReferrerName != session.ReferrerName {
		return true
	}

	return (request.UTMSource != "" && request.UTMSource != session.UTMSource) ||
		(request.UTMMedium != "" && request.UTMMedium != session.UTMMedium) ||
		(request.UTMCampaign != "" && request.UTMCampaign != session.UTMCampaign) ||
		(request.UTMContent != "" && request.UTMContent != session.UTMContent) ||
		(request.UTMTerm != "" && request.UTMTerm != session.UTMTerm)
}

func (s *Session) fingerprint(ua, ip string, now time.Time) uint64 {
	var sb strings.Builder
	sb.WriteString(ua)
	sb.WriteString(ip)
	sb.WriteString(s.fpSalt)
	sb.WriteString(now.Format("20060102"))
	return siphash.Hash(s.fpKey0, s.fpKey1, []byte(sb.String()))
}
