package tracker

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v4/tracker/referrer"
	"github.com/pirsch-analytics/pirsch/v4/tracker/session"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ua"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

func TestHitFromRequest(t *testing.T) {
	db.CleanupDB(t, dbClient)
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar&utm_source=test+source&utm_medium=email&utm_campaign=newsletter&utm_content=signup&utm_term=keywords", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", uaString)
	req.Header.Set("Referer", "http://ref/")
	pageView, sessionState, userAgent := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: session.NewMemCache(dbClient, 100),
		ClientID:     42,
		Title:        "title",
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})
	assert.NotNil(t, pageView)
	state := sessionState.State
	assert.Equal(t, 42, int(state.ClientID))
	assert.NotZero(t, state.VisitorID)
	assert.NoError(t, dbClient.SaveSessions([]model.Session{state}))
	assert.InDelta(t, time.Now().UTC().UnixMilli(), userAgent.Time.UnixMilli(), 30)
	assert.Equal(t, uaString, userAgent.UserAgent)

	if pageView.Time.IsZero() ||
		pageView.SessionID == 0 ||
		pageView.DurationSeconds != 0 ||
		pageView.Path != "/test/path" ||
		pageView.Title != "title" ||
		pageView.Language != "de" ||
		pageView.Referrer != "http://ref" ||
		pageView.OS != pirsch.OSWindows ||
		pageView.OSVersion != "10" ||
		pageView.Browser != pirsch.BrowserChrome ||
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

	if state.Sign != 1 ||
		state.Time.IsZero() ||
		state.SessionID == 0 ||
		state.DurationSeconds != 0 ||
		state.ExitPath != "/test/path" ||
		state.EntryPath != "/test/path" ||
		state.PageViews != 1 ||
		!state.IsBounce ||
		state.EntryTitle != "title" ||
		state.Language != "de" ||
		state.Referrer != "http://ref" ||
		state.OS != pirsch.OSWindows ||
		state.OSVersion != "10" ||
		state.Browser != pirsch.BrowserChrome ||
		state.BrowserVersion != "84.0" ||
		!state.Desktop ||
		state.Mobile ||
		state.ScreenWidth != 640 ||
		state.ScreenHeight != 1024 ||
		state.UTMSource != "test source" ||
		state.UTMMedium != "email" ||
		state.UTMCampaign != "newsletter" ||
		state.UTMContent != "signup" ||
		state.UTMTerm != "keywords" {
		t.Fatalf("Session not as expected: %v", state)
	}
}

func TestHitFromRequestSession(t *testing.T) {
	db.CleanupDB(t, dbClient)
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	sessionCache := session.NewMemCache(dbClient, 100)
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

	state := sessionCache.Sessions()[fmt.Sprintf("%d_%d", session1.ClientID, session1.VisitorID)]
	assert.False(t, state.Time.IsZero())
	assert.NotEqual(t, uint32(0), state.SessionID)
	assert.Equal(t, "/test/path", state.ExitPath)
	assert.Equal(t, "/test/path", state.EntryPath)
	assert.Equal(t, uint16(1), state.PageViews)
	state.Time = state.Time.Add(-time.Second * 5)   // manipulate the time the hit was created
	state.Start = state.Start.Add(-time.Second * 5) // manipulate the time the session was created
	state.ExitPath = "/different/path"
	sessionCache.Sessions()[fmt.Sprintf("%d_%d", session1.ClientID, session1.VisitorID)] = state

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
	db.CleanupDB(t, dbClient)
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	sessionCache := session.NewMemCache(dbClient, 100)
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
	assert.True(t, sessionState.State.IsBounce)
	assert.True(t, sessionState.Cancel.IsBounce)

	req.URL.Path = "/different/path"
	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	assert.True(t, sessionState.Cancel.IsBounce)
	assert.False(t, sessionState.State.IsBounce)
}

func TestHitFromRequestOverwrite(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: session.NewMemCache(dbClient, 100),
		URL:          "http://bar.foo/new/custom/path?query=param&foo=bar#anchor",
	})

	if sessionState.State.ExitPath != "/new/custom/path" {
		t.Fatalf("Session not as expected: %v", sessionState.State)
	}
}

func TestHitFromRequestOverwritePathAndReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, sessionState, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: session.NewMemCache(dbClient, 100),
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
		SessionCache: session.NewMemCache(dbClient, 100),
		ScreenWidth:  0,
		ScreenHeight: 400,
	})

	if sessionState.State.ScreenWidth != 0 || sessionState.State.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", sessionState.State.ScreenWidth, sessionState.State.ScreenHeight)
	}

	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: session.NewMemCache(dbClient, 100),
		ScreenWidth:  400,
		ScreenHeight: 0,
	})

	if sessionState.State.ScreenWidth != 0 || sessionState.State.ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", sessionState.State.ScreenWidth, sessionState.State.ScreenHeight)
	}

	_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: session.NewMemCache(dbClient, 100),
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})

	if sessionState.State.ScreenWidth != 640 || sessionState.State.ScreenHeight != 1024 {
		t.Fatalf("Screen size must be set, but was: %v %v", sessionState.State.ScreenWidth, sessionState.State.ScreenHeight)
	}
}

func TestHitFromRequestCountryCodeCity(t *testing.T) {
	sessionCache := session.NewMemCache(dbClient, 100)
	geoDB, err := geodb.NewGeoDB(geodb.Config{
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
	db.CleanupDB(t, dbClient)
	cache := session.NewMemCache(dbClient, 100)
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

	for _, param := range referrer.QueryParams {
		req.URL, _ = url.Parse(fmt.Sprintf("/test?%s=https://%s.com", param, param))
		_, sessionState, _ = HitFromRequest(req, "salt", &HitOptions{SessionCache: cache})
		assert.Equal(t, fmt.Sprintf("https://%s.com", param), cache.Get(0, sessionState.State.VisitorID, time.Now().Add(-time.Second)).Referrer)
	}
}

func TestHitFromRequestResetSessionUTM(t *testing.T) {
	db.CleanupDB(t, dbClient)
	cache := session.NewMemCache(dbClient, 100)
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

func TestHitFromRequestIsBot(t *testing.T) {
	db.CleanupDB(t, dbClient)
	cache := session.NewMemCache(dbClient, 100)
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	options := &HitOptions{
		SessionCache:   cache,
		ClientID:       42,
		MinDelay:       20,
		IsBotThreshold: 3,
	}

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/test/path/%d", i), nil)
		req.Header.Set("User-Agent", uaString)
		pageView, _, _ := HitFromRequest(req, "salt", options)
		assert.NotNil(t, pageView)

		for key := range cache.Sessions() {
			assert.Equal(t, uint8(i), cache.Sessions()[key].IsBot)
		}
	}

	// the last one must be ignored
	req := httptest.NewRequest(http.MethodGet, "/last/path", nil)
	req.Header.Set("User-Agent", uaString)
	pageView, _, _ := HitFromRequest(req, "salt", options)
	assert.Nil(t, pageView)

	for key := range cache.Sessions() {
		assert.Equal(t, uint8(3), cache.Sessions()[key].IsBot)
	}
}

func TestExtendSession(t *testing.T) {
	db.CleanupDB(t, dbClient)
	sessionCache := session.NewMemCache(dbClient, 100)
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")
	options := &HitOptions{
		SessionCache: sessionCache,
	}
	_, sessionState, _ := HitFromRequest(req, "salt", options)
	at := sessionState.State.Time
	time.Sleep(time.Millisecond * 1050)
	s := ExtendSession(req, "salt", options)
	state := sessionCache.Get(0, sessionState.State.VisitorID, time.Now().UTC().Add(-time.Second))
	assert.NotEqual(t, at, state.Time)
	assert.True(t, state.Time.After(at))
	assert.NotNil(t, s.Cancel)
	assert.Equal(t, state.Time, s.State.Time)
	assert.Equal(t, int8(-1), s.Cancel.Sign)
	assert.Equal(t, int8(1), s.State.Sign)
	assert.True(t, s.State.DurationSeconds > sessionState.State.DurationSeconds)
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
	for _, botUserAgent := range ua.Blacklist {
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

func TestGetIntQueryParam(t *testing.T) {
	input := []string{
		"",
		"   ",
		"asdf",
		"32asdf",
		"42",
	}
	expectedUInt64 := []uint64{
		0,
		0,
		0,
		0,
		42,
	}
	expectedUInt16 := []uint16{
		0,
		0,
		0,
		0,
		42,
	}

	for i, in := range input {
		if out := getIntQueryParam[uint64](in); out != expectedUInt64[i] {
			t.Fatalf("Expected '%v', but was: %v", expectedUInt64[i], out)
		}
	}

	for i, in := range input {
		if out := getIntQueryParam[uint16](in); out != expectedUInt16[i] {
			t.Fatalf("Expected '%v', but was: %v", expectedUInt64[i], out)
		}
	}
}

func TestGetURLQueryParam(t *testing.T) {
	assert.Equal(t, "https://test.com/foo/bar?param=value#anchor", getURLQueryParam("https://test.com/foo/bar?param=value#anchor"))
	assert.Empty(t, getURLQueryParam("test"))
}
