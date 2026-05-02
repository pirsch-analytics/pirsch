package session

import (
	"math/rand/v2"
	"net/http"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, start := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// time on page
		assert.Equal(t, uint64(0), req.DurationSeconds)

		// the request must have the new session attached
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)

		// check that there is one session with all the relevant fields within the cache
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		firstSession := sessions[0]
		assert.Equal(t, int8(1), firstSession.Sign)
		assert.Equal(t, start, firstSession.Time)
		assert.Equal(t, start, firstSession.Start)
		assert.Equal(t, uint16(1), firstSession.PageViews)
		assert.True(t, firstSession.IsBounce)
		assert.Equal(t, uint64(1), firstSession.ClientID)
		assert.NotZero(t, firstSession.VisitorID)
		assert.NotZero(t, firstSession.SessionID)
		assert.Equal(t, "example.com", firstSession.Hostname)
		assert.Equal(t, "/", firstSession.EntryPath)
		assert.Equal(t, "/", firstSession.ExitPath)
		assert.Equal(t, "Title", firstSession.EntryTitle)
		assert.Equal(t, "Title", firstSession.ExitTitle)
		assert.Equal(t, "https://google.com", firstSession.Referrer)
		assert.Equal(t, "Google", firstSession.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", firstSession.ReferrerIcon)
		assert.Equal(t, "XL", firstSession.ScreenClass)
		assert.Equal(t, "fr", firstSession.Language)
		assert.Equal(t, "fr", firstSession.CountryCode)
		assert.Equal(t, "Auvergne-Rhône-Alpes", firstSession.Region)
		assert.Equal(t, "Lyon", firstSession.City)
		assert.Equal(t, pkg.OSWindows, firstSession.OS)
		assert.Equal(t, "10", firstSession.OSVersion)
		assert.Equal(t, pkg.BrowserChrome, firstSession.Browser)
		assert.Equal(t, "146", firstSession.BrowserVersion)
		assert.True(t, firstSession.Desktop)
		assert.Equal(t, "utm_source", firstSession.UTMSource)
		assert.Equal(t, "utm_medium", firstSession.UTMMedium)
		assert.Equal(t, "utm_campaign", firstSession.UTMCampaign)
		assert.Equal(t, "utm_content", firstSession.UTMContent)
		assert.Equal(t, "utm_term", firstSession.UTMTerm)
		assert.Equal(t, "channel", firstSession.Channel)
		assert.Equal(t, uint32(0), firstSession.DurationSeconds)

		// wait and make a second request
		time.Sleep(time.Second * 23)
		req, now := newSampleRequest()
		req.Path = "/about"
		req.Title = "About"
		req.Referrer = ""
		req.ReferrerName = ""
		req.ReferrerIcon = ""
		req.UTMSource = ""
		req.UTMMedium = ""
		req.UTMCampaign = ""
		req.UTMContent = ""
		req.UTMTerm = ""
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// time on page
		assert.Equal(t, uint64(23), req.DurationSeconds)

		// the request must have the new session and cancelled session attached
		assert.NotNil(t, req.Session)
		assert.NotNil(t, req.CancelSession)
		assert.Equal(t, int8(-1), req.CancelSession.Sign)
		assert.Equal(t, start, req.CancelSession.Time)

		// check that there is one session with all the relevant fields within the cache
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		secondSession := sessions[0]
		assert.Equal(t, now, secondSession.Time)
		assert.Equal(t, start, secondSession.Start)
		assert.Equal(t, uint16(2), secondSession.PageViews)
		assert.False(t, secondSession.IsBounce)
		assert.Equal(t, uint64(1), secondSession.ClientID)
		assert.NotZero(t, secondSession.VisitorID)
		assert.NotZero(t, secondSession.SessionID)
		assert.Equal(t, "example.com", secondSession.Hostname)
		assert.Equal(t, "/", secondSession.EntryPath)
		assert.Equal(t, "/about", secondSession.ExitPath)
		assert.Equal(t, "Title", secondSession.EntryTitle)
		assert.Equal(t, "About", secondSession.ExitTitle)
		assert.Equal(t, "https://google.com", secondSession.Referrer)
		assert.Equal(t, "Google", secondSession.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", secondSession.ReferrerIcon)
		assert.Equal(t, "XL", secondSession.ScreenClass)
		assert.Equal(t, "fr", secondSession.Language)
		assert.Equal(t, "fr", secondSession.CountryCode)
		assert.Equal(t, "Auvergne-Rhône-Alpes", secondSession.Region)
		assert.Equal(t, "Lyon", secondSession.City)
		assert.Equal(t, pkg.OSWindows, secondSession.OS)
		assert.Equal(t, "10", secondSession.OSVersion)
		assert.Equal(t, pkg.BrowserChrome, secondSession.Browser)
		assert.Equal(t, "146", secondSession.BrowserVersion)
		assert.True(t, secondSession.Desktop)
		assert.Equal(t, "utm_source", secondSession.UTMSource)
		assert.Equal(t, "utm_medium", secondSession.UTMMedium)
		assert.Equal(t, "utm_campaign", secondSession.UTMCampaign)
		assert.Equal(t, "utm_content", secondSession.UTMContent)
		assert.Equal(t, "utm_term", secondSession.UTMTerm)
		assert.Equal(t, "channel", secondSession.Channel)
		assert.Equal(t, uint32(23), secondSession.DurationSeconds)

		// wait and make a third request
		time.Sleep(time.Second * 8)
		previousSessionTime := now
		req, now = newSampleRequest()
		req.Path = "/third"
		req.Title = "Third"
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// time on page
		assert.Equal(t, uint64(8), req.DurationSeconds)

		// the request must have the new session and cancelled session attached
		assert.NotNil(t, req.Session)
		assert.NotNil(t, req.CancelSession)
		assert.Equal(t, uint32(31), req.Session.DurationSeconds)
		assert.Equal(t, int8(-1), req.CancelSession.Sign)
		assert.Equal(t, previousSessionTime, req.CancelSession.Time)

		// check that there is one session with all the relevant fields within the cache
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		thirdSession := sessions[0]
		assert.Equal(t, now, thirdSession.Time)
		assert.Equal(t, start, thirdSession.Start)
		assert.Equal(t, uint16(3), thirdSession.PageViews)
		assert.Equal(t, uint32(31), thirdSession.DurationSeconds)
	})
}

func TestSessionBounced(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)

		// check that the session is marked as bounced
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		firstSession := sessions[0]
		assert.True(t, firstSession.IsBounce)

		// wait and make a second request to the same page
		time.Sleep(time.Second * 23)
		req, _ = newSampleRequest()
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.NotNil(t, req.CancelSession)

		// check that there is one session with all the relevant fields within the cache
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		secondSession := sessions[0]
		assert.True(t, secondSession.IsBounce)
	})
}

func TestSessionEventNonInteractive(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, _ := newSampleRequest()
		req.EventName = "Event"
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)

		// check that the session is marked as bounced
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		firstSession := sessions[0]
		assert.True(t, firstSession.IsBounce)

		// wait and make a second request to the same page
		time.Sleep(time.Second * 23)
		req, _ = newSampleRequest()
		req.EventName = "Event"
		req.EventNonInteractive = true
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.NotNil(t, req.CancelSession)

		// check that there is one bounced session
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		secondSession := sessions[0]
		assert.True(t, secondSession.IsBounce)

		// wait and make a third request to the same page
		time.Sleep(time.Second * 23)
		req, _ = newSampleRequest()
		req.EventName = "Event"
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.NotNil(t, req.CancelSession)

		// check that there is one non-bounced session
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		thirdSession := sessions[0]
		assert.False(t, thirdSession.IsBounce)
	})
}

func TestSessionReferrerReset(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)

		// the request must have the referrer attached
		visitorID := req.Session.VisitorID
		sessionID := req.Session.SessionID
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)
		assert.Equal(t, "https://google.com", req.Session.Referrer)
		assert.Equal(t, "Google", req.Session.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", req.Session.ReferrerIcon)

		// check that there is one session with the referrer attached
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		firstSession := sessions[0]
		assert.Equal(t, "https://google.com", firstSession.Referrer)
		assert.Equal(t, "Google", firstSession.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", firstSession.ReferrerIcon)

		// wait and make a second request with a different referrer
		time.Sleep(time.Second * 23)
		req, _ = newSampleRequest()
		req.Referrer = "https://bing.com"
		req.ReferrerName = "Bing"
		req.ReferrerIcon = "https://bing.com/favicon.ico"
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// the request must have the new session and referrer attached
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)
		assert.Equal(t, visitorID, req.Session.VisitorID)
		assert.NotEqual(t, sessionID, req.Session.SessionID)
		assert.Equal(t, "https://bing.com", req.Session.Referrer)
		assert.Equal(t, "Bing", req.Session.ReferrerName)
		assert.Equal(t, "https://bing.com/favicon.ico", req.Session.ReferrerIcon)

		// check that there is one session with the referrer attached
		// (the first session is overwritten for the same visitor ID)
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		secondSession := sessions[0]
		assert.Equal(t, "https://bing.com", secondSession.Referrer)
		assert.Equal(t, "Bing", secondSession.ReferrerName)
		assert.Equal(t, "https://bing.com/favicon.ico", secondSession.ReferrerIcon)
	})
}

func TestSessionUTMReset(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)

		// the request must have the UTM attached
		visitorID := req.Session.VisitorID
		sessionID := req.Session.SessionID
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)
		assert.Equal(t, "utm_source", req.Session.UTMSource)
		assert.Equal(t, "utm_campaign", req.Session.UTMCampaign)
		assert.Equal(t, "utm_medium", req.Session.UTMMedium)
		assert.Equal(t, "utm_content", req.Session.UTMContent)
		assert.Equal(t, "utm_term", req.Session.UTMTerm)

		// check that there is one session with the UTM attached
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		firstSession := sessions[0]
		assert.Equal(t, "utm_source", firstSession.UTMSource)
		assert.Equal(t, "utm_campaign", firstSession.UTMCampaign)
		assert.Equal(t, "utm_medium", firstSession.UTMMedium)
		assert.Equal(t, "utm_content", firstSession.UTMContent)
		assert.Equal(t, "utm_term", firstSession.UTMTerm)

		// wait and make a second request with a different UTM
		time.Sleep(time.Second * 23)
		req, _ = newSampleRequest()
		req.UTMSource = "new_source"
		req.UTMCampaign = "new_campaign"
		req.UTMMedium = "new_medium"
		req.UTMContent = "new_content"
		req.UTMTerm = "new_term"
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// the request must have the new session and UTM attached
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)
		assert.Equal(t, visitorID, req.Session.VisitorID)
		assert.NotEqual(t, sessionID, req.Session.SessionID)
		assert.Equal(t, "new_source", req.Session.UTMSource)
		assert.Equal(t, "new_campaign", req.Session.UTMCampaign)
		assert.Equal(t, "new_medium", req.Session.UTMMedium)
		assert.Equal(t, "new_content", req.Session.UTMContent)
		assert.Equal(t, "new_term", req.Session.UTMTerm)

		// check that there is one session with the UTM attached
		// (the first session is overwritten for the same visitor ID)
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		secondSession := sessions[0]
		assert.Equal(t, "new_source", secondSession.UTMSource)
		assert.Equal(t, "new_campaign", secondSession.UTMCampaign)
		assert.Equal(t, "new_medium", secondSession.UTMMedium)
		assert.Equal(t, "new_content", secondSession.UTMContent)
		assert.Equal(t, "new_term", secondSession.UTMTerm)
	})
}

func TestSessionReferrerHostname(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)

		// the request must have the referrer attached
		visitorID := req.Session.VisitorID
		sessionID := req.Session.SessionID
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)
		assert.Equal(t, "https://google.com", req.Session.Referrer)
		assert.Equal(t, "Google", req.Session.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", req.Session.ReferrerIcon)

		// check that there is one session with the referrer attached
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		firstSession := sessions[0]
		assert.Equal(t, "https://google.com", firstSession.Referrer)
		assert.Equal(t, "Google", firstSession.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", firstSession.ReferrerIcon)

		// wait and make a second request with the hostname as referrer (empty, as it will be set by the referrer step)
		time.Sleep(time.Second * 23)
		req, _ = newSampleRequest()
		req.Referrer = ""
		req.ReferrerName = ""
		req.ReferrerIcon = ""
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// the request must have the same session and referrer attached
		assert.NotNil(t, req.Session)
		assert.NotNil(t, req.CancelSession)
		assert.Equal(t, visitorID, req.Session.VisitorID)
		assert.Equal(t, sessionID, req.Session.SessionID)
		assert.Equal(t, "https://google.com", req.Session.Referrer)
		assert.Equal(t, "Google", req.Session.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", req.Session.ReferrerIcon)

		// check that there is one session with the original referrer attached
		// (the first session is overwritten for the same visitor ID)
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		secondSession := sessions[0]
		assert.Equal(t, "https://google.com", secondSession.Referrer)
		assert.Equal(t, "Google", secondSession.ReferrerName)
		assert.Equal(t, "https://google.com/favicon.ico", secondSession.ReferrerIcon)
	})
}

func TestSessionTimeout(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// wait and make a second request
		visitorID := req.Session.VisitorID
		sessionID := req.Session.SessionID
		time.Sleep(sessionTimeout + time.Minute)
		req, _ = newSampleRequest()
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// the request must have created a new session
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)
		assert.Equal(t, visitorID, req.Session.VisitorID)
		assert.NotEqual(t, sessionID, req.Session.SessionID)
	})
}

func TestSessionMaxAge(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request at 23:45 UTC
		sleepUntil(23, 45)
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// manipulate start time
		for k := range cache.sessions {
			v := cache.sessions[k]
			v.Start = v.Start.Add(-sessionMaxAge - time.Minute)
			cache.sessions[k] = v
		}

		// wait and make a second request at 00:05 UTC
		visitorID := req.Session.VisitorID
		sessionID := req.Session.SessionID
		time.Sleep(time.Minute * 20)
		req, _ = newSampleRequest()
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// the request must have found the previous session
		assert.NotNil(t, req.Session)
		assert.Nil(t, req.CancelSession)
		assert.NotEqual(t, visitorID, req.Session.VisitorID)
		assert.NotEqual(t, sessionID, req.Session.SessionID)
	})
}

func TestSessionUpdateSession(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// wait and make a second request without updating the session
		now := time.Now()
		time.Sleep(time.Minute * 5)
		req, _ = newSampleRequest()
		req.UpdateSession = true
		cancel, err = s.Step(req)
		assert.True(t, cancel)
		assert.NoError(t, err)

		// the request must not have sessions attached
		assert.Nil(t, req.Session)
		assert.Nil(t, req.CancelSession)

		// the session in cache must have been updated however
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		assert.True(t, sessions[0].Time.After(now))
		assert.Equal(t, uint16(1), sessions[0].Extended)
	})
}

func TestSessionYesterday(t *testing.T) {
	// create an in-memory cache and session step
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)

	synctest.Test(t, func(t *testing.T) {
		// make the first request at 23:45 UTC
		sleepUntil(23, 45)
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// wait and make a second request at 00:05 UTC
		visitorID := req.Session.VisitorID
		sessionID := req.Session.SessionID
		time.Sleep(time.Minute * 20)
		req, _ = newSampleRequest()
		cancel, err = s.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)

		// the request must have found the previous session
		assert.NotNil(t, req.Session)
		assert.NotNil(t, req.CancelSession)
		assert.Equal(t, visitorID, req.Session.VisitorID)
		assert.Equal(t, sessionID, req.Session.SessionID)
	})
}

func TestSessionMaxPageViews(t *testing.T) {
	// create an in-memory cache and session step with a maximum of 10 page views
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 10)

	synctest.Test(t, func(t *testing.T) {
		// make exactly 10 requests
		for range 10 {
			req, _ := newSampleRequest()
			cancel, err := s.Step(req)
			assert.False(t, cancel)
			assert.NoError(t, err)
			time.Sleep(time.Second * time.Duration(rand.IntN(10)+1))
		}

		// there must be one session with 10 page views
		sessions := getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		assert.Equal(t, uint16(10), sessions[0].PageViews)

		// make one more request
		req, _ := newSampleRequest()
		cancel, err := s.Step(req)
		assert.True(t, cancel)
		assert.NoError(t, err)

		// the last request must have been ignored
		sessions = getSessions(cache.Sessions())
		assert.Len(t, sessions, 1)
		assert.Equal(t, uint16(10), sessions[0].PageViews)
	})
}

func TestSessionFingerprint(t *testing.T) {
	cache := NewMemCache(client, 100)
	s := NewSession(1, 2, "salt", cache, 100)
	now := time.Now().UTC()
	fp1 := s.fingerprint("ua", "81.2.69.142", now)
	fp2 := s.fingerprint("ua", "81.2.69.142", now)
	fp3 := s.fingerprint("ua", "2001:9e8:d5d2:b00:ce0a:96e4:ae42:c935", now)
	fp4 := s.fingerprint("ua2", "81.2.69.142", now)
	fp5 := s.fingerprint("ua", "81.2.69.142", now.Add(time.Hour*25))
	assert.Equal(t, fp1, fp2)
	assert.NotEqual(t, fp1, fp3)
	assert.NotEqual(t, fp1, fp4)
	assert.NotEqual(t, fp1, fp5)
}

func newSampleRequest() (*ingest.Request, time.Time) {
	r, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	now := time.Now().UTC()
	return &ingest.Request{
		Request:        r,
		ClientID:       1,
		Time:           now,
		Hostname:       "example.com",
		Path:           "/",
		Title:          "Title",
		Referrer:       "https://google.com",
		Language:       "fr",
		CountryCode:    "fr",
		Region:         "Auvergne-Rhône-Alpes",
		City:           "Lyon",
		ReferrerName:   "Google",
		ReferrerIcon:   "https://google.com/favicon.ico",
		OS:             pkg.OSWindows,
		OSVersion:      "10",
		Browser:        pkg.BrowserChrome,
		BrowserVersion: "146",
		Desktop:        true,
		ScreenClass:    "XL",
		UTMSource:      "utm_source",
		UTMMedium:      "utm_medium",
		UTMCampaign:    "utm_campaign",
		UTMContent:     "utm_content",
		UTMTerm:        "utm_term",
		Channel:        "channel",
	}, now
}

func getSessions(sessions map[string]model.Session) []model.Session {
	list := make([]model.Session, 0, len(sessions))

	for _, session := range sessions {
		list = append(list, session)
	}

	slices.SortFunc(list, func(a, b model.Session) int {
		if a.Time.Before(b.Time) {
			return -1
		} else if a.Time.After(b.Time) {
			return 1
		}

		return 0
	})
	return list
}

func sleepUntil(hour, minute int) {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)

	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}

	time.Sleep(time.Until(next))
}
