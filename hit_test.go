package pirsch

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

func TestHitFromRequest(t *testing.T) {
	cleanupDB()
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar&utm_source=test+source&utm_medium=email&utm_campaign=newsletter&utm_content=signup&utm_term=keywords", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", uaString)
	req.Header.Set("Referer", "http://ref/")
	pageView, sessionState, ua := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ClientID:     42,
		Title:        "title",
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})
	assert.NotNil(t, pageView)
	session := sessionState.State
	assert.Equal(t, 42, int(session.ClientID))
	assert.NotZero(t, session.VisitorID)
	assert.NoError(t, dbClient.SaveSessions([]Session{session}))
	assert.InDelta(t, time.Now().UTC().UnixMilli(), ua.Time.UnixMilli(), 30)
	assert.Equal(t, uaString, ua.UserAgent)

	if pageView.Time.IsZero() ||
		pageView.SessionID == 0 ||
		pageView.DurationSeconds != 0 ||
		pageView.Path != "/test/path" ||
		pageView.Title != "title" ||
		pageView.Language != "de" ||
		pageView.Referrer != "http://ref" ||
		pageView.OS != OSWindows ||
		pageView.OSVersion != "10" ||
		pageView.Browser != BrowserChrome ||
		pageView.BrowserVersion != "84.0" ||
		!pageView.Desktop ||
		pageView.Mobile ||
		pageView.ScreenWidth != 640 ||
		pageView.ScreenHeight != 1024 ||
		pageView.UTMSource != "test source" ||
		pageView.UTMMedium != "email" ||
		pageView.UTMCampaign != "newsletter" ||
		pageView.UTMContent != "signup" ||
		pageView.UTMTerm != "keywords" {
		t.Fatalf("PageView not as expected: %v", pageView)
	}

	if session.Sign != 1 ||
		session.Time.IsZero() ||
		session.SessionID == 0 ||
		session.DurationSeconds != 0 ||
		session.ExitPath != "/test/path" ||
		session.EntryPath != "/test/path" ||
		session.PageViews != 1 ||
		!session.IsBounce ||
		session.EntryTitle != "title" ||
		session.Language != "de" ||
		session.Referrer != "http://ref" ||
		session.OS != OSWindows ||
		session.OSVersion != "10" ||
		session.Browser != BrowserChrome ||
		session.BrowserVersion != "84.0" ||
		!session.Desktop ||
		session.Mobile ||
		session.ScreenWidth != 640 ||
		session.ScreenHeight != 1024 ||
		session.UTMSource != "test source" ||
		session.UTMMedium != "email" ||
		session.UTMCampaign != "newsletter" ||
		session.UTMContent != "signup" ||
		session.UTMTerm != "keywords" {
		t.Fatalf("Session not as expected: %v", session)
	}
}

func TestHitFromRequestSession(t *testing.T) {
	cleanupDB()
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	sessionCache := NewSessionCacheMem(dbClient, 100)
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", uaString)
	req.Header.Set("Referer", "http://ref/")
	pageView1, sessionState, ua1 := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	assert.NotNil(t, pageView1)
	session1 := sessionState.State
	assert.Equal(t, int8(1), session1.Sign)
	assert.Equal(t, uint64(0), session1.ClientID)
	assert.NotZero(t, session1.VisitorID)
	assert.Equal(t, "/test/path", session1.ExitPath)
	assert.Equal(t, "/test/path", session1.EntryPath)
	assert.Equal(t, uint32(0), session1.DurationSeconds)
	assert.True(t, session1.IsBounce)
	assert.InDelta(t, time.Now().UTC().UnixMilli(), ua1.Time.UnixMilli(), 10)
	assert.Equal(t, uaString, ua1.UserAgent)
	assert.Equal(t, uint64(0), pageView1.ClientID)
	assert.NotZero(t, pageView1.VisitorID)
	assert.Equal(t, "/test/path", pageView1.Path)
	assert.Equal(t, uint32(0), pageView1.DurationSeconds)

	session := sessionCache.sessions[getSessionKey(session1.ClientID, session1.VisitorID)]
	assert.False(t, session.Time.IsZero())
	assert.NotEqual(t, uint32(0), session.SessionID)
	assert.Equal(t, "/test/path", session.ExitPath)
	assert.Equal(t, "/test/path", session.EntryPath)
	assert.Equal(t, uint16(1), session.PageViews)
	session.Time = session.Time.Add(-time.Second * 5)   // manipulate the time the hit was created
	session.Start = session.Start.Add(-time.Second * 5) // manipulate the time the session was created
	session.ExitPath = "/different/path"
	sessionCache.sessions[getSessionKey(session1.ClientID, session1.VisitorID)] = session

	pageView2, sessionState2, ua2 := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	session2 := sessionState2.State
	assert.Equal(t, int8(-1), sessionState2.Cancel.Sign)
	assert.Equal(t, int8(1), session2.Sign)
	assert.Equal(t, uint64(0), session2.ClientID)
	assert.Equal(t, session1.VisitorID, session2.VisitorID)
	assert.Equal(t, "/test/path", session2.ExitPath)
	assert.Equal(t, "/test/path", session2.EntryPath)
	assert.Equal(t, uint32(5), session2.DurationSeconds)
	assert.False(t, session2.IsBounce)
	assert.Nil(t, ua2)
	assert.Equal(t, uint64(0), pageView2.ClientID)
	assert.NotZero(t, pageView2.VisitorID)
	assert.Equal(t, "/test/path", pageView2.Path)
	assert.Equal(t, uint32(5), pageView2.DurationSeconds)
}

func TestHitFromRequestBounces(t *testing.T) {
	cleanupDB()
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	sessionCache := NewSessionCacheMem(dbClient, 100)
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", uaString)
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	session1 := sessionState.State
	assert.True(t, session1.IsBounce)

	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	assert.False(t, sessionState.State.IsBounce)
	assert.True(t, sessionState.Cancel.IsBounce)

	req.URL.Path = "/different/path"
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	assert.False(t, sessionState.Cancel.IsBounce)
	assert.False(t, sessionState.State.IsBounce)
}

func TestHitFromRequestOverwrite(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		URL:          "http://bar.foo/new/custom/path?query=param&foo=bar#anchor",
	})

	if sessionState.State.ExitPath != "/new/custom/path" {
		t.Fatalf("Session not as expected: %v", sessionState.State)
	}
}

func TestHitFromRequestOverwritePathAndReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		URL:          "http://bar.foo/overwrite/this?query=param&foo=bar#anchor",
		Path:         "/new/custom/path",
		Referrer:     "http://custom.ref/",
	})

	if sessionState.State.ExitPath != "/new/custom/path" || sessionState.State.Referrer != "http://custom.ref" {
		t.Fatalf("Session not as expected: %v", sessionState.State)
	}
}

func TestHitFromRequestScreenSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ScreenWidth:  0,
		ScreenHeight: 400,
	})

	if sessionState.State.ScreenWidth != 0 || sessionState.State.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", sessionState.State.ScreenWidth, sessionState.State.ScreenHeight)
	}

	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ScreenWidth:  400,
		ScreenHeight: 0,
	})

	if sessionState.State.ScreenWidth != 0 || sessionState.State.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", sessionState.State.ScreenWidth, sessionState.State.ScreenHeight)
	}

	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})

	if sessionState.State.ScreenWidth != 640 || sessionState.State.ScreenHeight != 1024 {
		t.Fatalf("Screen size must be set, but was: %v %v", sessionState.State.ScreenWidth, sessionState.State.ScreenHeight)
	}
}

func TestHitFromRequestCountryCodeCity(t *testing.T) {
	sessionCache := NewSessionCacheMem(dbClient, 100)
	geoDB, err := NewGeoDB(GeoDBConfig{
		File: filepath.Join("geodb/GeoIP2-City-Test.mmdb"),
	})
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	req.RemoteAddr = "81.2.69.142"
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
		geoDB:        geoDB,
	})
	assert.Equal(t, "gb", sessionState.State.CountryCode)
	assert.Equal(t, "London", sessionState.State.City)
	req = httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	req.RemoteAddr = "127.0.0.1"
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
		geoDB:        geoDB,
	})
	assert.Empty(t, sessionState.State.CountryCode)
	assert.Empty(t, sessionState.State.City)
}

func TestHitFromRequestResetSessionReferrer(t *testing.T) {
	cleanupDB()
	cache := NewSessionCacheMem(dbClient, 100)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	req.Header.Set("Referer", "https://referrer-header.com")
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
	assert.Equal(t, "https://referrer-header.com", cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).Referrer)

	// keep session
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
	assert.Equal(t, "https://referrer-header.com", cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).Referrer)
	req.Header.Del("Referer")
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
	assert.Equal(t, "https://referrer-header.com", cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).Referrer)

	// create new session on changing referrer
	req.Header.Set("Referer", "https://new-referrer-header.com")
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
	assert.Equal(t, "https://new-referrer-header.com", cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).Referrer)

	// URL query parameters
	req.Header.Del("Referer")

	for _, param := range referrerQueryParams {
		req.URL, _ = url.Parse(fmt.Sprintf("/test?%s=https://%s.com", param, param))
		_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
		assert.Equal(t, fmt.Sprintf("https://%s.com", param), cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).Referrer)
	}
}

func TestHitFromRequestResetSessionUTM(t *testing.T) {
	cleanupDB()
	cache := NewSessionCacheMem(dbClient, 100)
	req := httptest.NewRequest(http.MethodGet, "/test?utm_source=source&utm_medium=medium&utm_campaign=campaign&utm_content=content&utm_term=term", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
	assert.Equal(t, "source", cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).UTMSource)

	// keep session
	sessionID := sessionState.State.SessionID
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
	assert.Equal(t, sessionID, cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).SessionID)
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
	assert.Equal(t, sessionID, cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).SessionID)

	// create new session on changing utm parameter
	params := [][]string{
		{"new-source", "medium", "campaign", "content", "term"},
		{"new-source", "new-medium", "campaign", "content", "term"},
		{"new-source", "new-medium", "new-campaign", "content", "term"},
		{"new-source", "new-medium", "new-campaign", "new-content", "term"},
		{"new-source", "new-medium", "new-campaign", "new-content", "new-term"},
	}
	sessionID = sessionState.State.SessionID

	for _, p := range params {
		req.URL, _ = url.Parse(fmt.Sprintf("/test?utm_source=%s&utm_medium=%s&utm_campaign=%s&utm_content=%s&utm_term=%s", p[0], p[1], p[2], p[3], p[4]))
		_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
		assert.NotEqual(t, sessionID, cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).SessionID)
		sessionID = sessionState.State.SessionID
	}
}

func TestExtendSession(t *testing.T) {
	cleanupDB()
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	sessionCache := NewSessionCacheMem(dbClient, 100)
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.Header.Set("User-Agent", uaString)
	options := &HitOptions{
		SessionCache: sessionCache,
	}
	_, sessionState, _ := HitFromRequest(req, "salt", options)
	at := sessionState.State.Time
	ExtendSession(req, "salt", options)
	session := sessionCache.Get(0, sessionState.State.VisitorID, time.Now().UTC().Add(-time.Second))
	assert.NotEqual(t, at, session.Time)
	assert.True(t, session.Time.After(at))
}

func TestIgnoreHitPrefetch(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req.Header.Set("X-Moz", "prefetch")

	if !IgnoreHit(req) {
		t.Fatal("Session with X-Moz header must be ignored")
	}

	req.Header.Del("X-Moz")
	req.Header.Set("X-Purpose", "prefetch")

	if !IgnoreHit(req) {
		t.Fatal("Session with X-Purpose header must be ignored")
	}

	req.Header.Set("X-Purpose", "preview")

	if !IgnoreHit(req) {
		t.Fatal("Session with X-Purpose header must be ignored")
	}

	req.Header.Del("X-Purpose")
	req.Header.Set("Purpose", "prefetch")

	if !IgnoreHit(req) {
		t.Fatal("Session with Purpose header must be ignored")
	}

	req.Header.Set("Purpose", "preview")

	if !IgnoreHit(req) {
		t.Fatal("Session with Purpose header must be ignored")
	}

	req.Header.Del("Purpose")

	if IgnoreHit(req) {
		t.Fatal("Session must not be ignored")
	}
}

func TestIgnoreHitUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "This is a bot request")

	if !IgnoreHit(req) {
		t.Fatal("Session with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a crawler request")

	if !IgnoreHit(req) {
		t.Fatal("Session with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a spider request")

	if !IgnoreHit(req) {
		t.Fatal("Session with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Visit http://spam.com!")

	if !IgnoreHit(req) {
		t.Fatal("Session with URL in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Mozilla/123.0")

	if IgnoreHit(req) {
		t.Fatal("Session with regular User-Agent must not be ignored")
	}

	req.Header.Set("User-Agent", "")

	if !IgnoreHit(req) {
		t.Fatal("Session with empty User-Agent must be ignored")
	}
}

func TestIgnoreHitBotUserAgent(t *testing.T) {
	for _, botUserAgent := range userAgentBlacklist {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", botUserAgent)

		if !IgnoreHit(req) {
			t.Fatalf("Session with user agent '%v' must have been ignored", botUserAgent)
		}
	}
}

func TestIgnoreHitReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "ua")
	req.Header.Set("Referer", "2your.site")

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}

	req.Header.Set("Referer", "subdomain.2your.site")

	if !IgnoreHit(req) {
		t.Fatal("Request for subdomain must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/?ref=2your.site", nil)

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}
}

func TestIgnoreHitBrowserVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.4147.135 Safari/537.36")

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if IgnoreHit(req) {
		t.Fatal("Request must not have been ignored")
	}
}

func TestIgnoreHitDoNotTrack(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if IgnoreHit(req) {
		t.Fatal("Request must not have been ignored")
	}

	req.Header.Set("DNT", "1")

	if !IgnoreHit(req) {
		t.Fatal("Request must have been ignored")
	}
}

func TestHitOptionsFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://test.com/my/path", nil)
	options := HitOptionsFromRequest(req)

	if options.ClientID != 0 ||
		options.URL != "" ||
		options.Title != "" ||
		options.Referrer != "" ||
		options.ScreenWidth != 0 ||
		options.ScreenHeight != 0 {
		t.Fatalf("HitOptions not as expected: %v", options)
	}

	req = httptest.NewRequest(http.MethodGet, "http://test.com/my/path?client_id=42&url=http://foo.bar/test&t=title&ref=http://ref/&w=640&h=1024", nil)
	options = HitOptionsFromRequest(req)

	if options.ClientID != 42 ||
		options.URL != "http://foo.bar/test" ||
		options.Title != "title" ||
		options.Referrer != "http://ref/" ||
		options.ScreenWidth != 640 ||
		options.ScreenHeight != 1024 {
		t.Fatalf("HitOptions not as expected: %v", options)
	}
}

func TestShortenString(t *testing.T) {
	out := shortenString("Hello World", 5)

	if out != "Hello" {
		t.Fatalf("String must have been shortened to 'Hello', but was: %v", out)
	}

	out = shortenString("Hello World", 50)

	if out != "Hello World" {
		t.Fatalf("String must not have been shortened, but was: %v", out)
	}
}

func TestGetUInt16QueryParam(t *testing.T) {
	input := []string{
		"",
		"   ",
		"asdf",
		"32asdf",
		"42",
	}
	expected := []uint16{
		0,
		0,
		0,
		0,
		42,
	}

	for i, in := range input {
		if out := getUInt16QueryParam(in); out != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], out)
		}
	}
}

func TestGetUInt64QueryParam(t *testing.T) {
	input := []string{
		"",
		"   ",
		"asdf",
		"32asdf",
		"42",
	}
	expected := []uint64{
		0,
		0,
		0,
		0,
		42,
	}

	for i, in := range input {
		if out := getUInt64QueryParam(in); out != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], out)
		}
	}
}

func TestGetURLQueryParam(t *testing.T) {
	assert.Equal(t, "https://test.com/foo/bar?param=value#anchor", getURLQueryParam("https://test.com/foo/bar?param=value#anchor"))
	assert.Empty(t, getURLQueryParam("test"))
}

func TestMin(t *testing.T) {
	assert.Equal(t, int64(5), min(5, 7))
	assert.Equal(t, int64(19), min(34, 19))
}
