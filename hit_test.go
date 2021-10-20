package pirsch

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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
	pageView, sessions, ua := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ClientID:     42,
		Title:        "title",
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})
	assert.NotNil(t, pageView)
	assert.Len(t, sessions, 1)
	session := sessions[0]
	assert.Equal(t, 42, int(session.ClientID))
	assert.NotZero(t, session.VisitorID)
	assert.NoError(t, dbClient.SaveSessions(sessions))
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
		session.Path != "/test/path" ||
		session.EntryPath != "/test/path" ||
		session.PageViews != 1 ||
		!session.IsBounce ||
		session.Title != "title" ||
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
	pageView1, sessions, ua1 := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	assert.NotNil(t, pageView1)
	assert.Len(t, sessions, 1)
	session1 := sessions[0]
	assert.Equal(t, int8(1), session1.Sign)
	assert.Equal(t, uint64(0), session1.ClientID)
	assert.NotZero(t, session1.VisitorID)
	assert.Equal(t, "/test/path", session1.Path)
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
	assert.Equal(t, "/test/path", session.Path)
	assert.Equal(t, "/test/path", session.EntryPath)
	assert.Equal(t, uint16(1), session.PageViews)
	session.Time = session.Time.Add(-time.Second * 5)   // manipulate the time the hit was created
	session.Start = session.Start.Add(-time.Second * 5) // manipulate the time the session was created
	session.Path = "/different/path"
	sessionCache.sessions[getSessionKey(session1.ClientID, session1.VisitorID)] = session

	pageView2, sessions, ua2 := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
	})
	assert.Len(t, sessions, 2)
	session2 := sessions[1]
	assert.Equal(t, int8(-1), sessions[0].Sign)
	assert.Equal(t, int8(1), session2.Sign)
	assert.Equal(t, uint64(0), session2.ClientID)
	assert.Equal(t, session1.VisitorID, session2.VisitorID)
	assert.Equal(t, "/test/path", session2.Path)
	assert.Equal(t, "/test/path", session2.EntryPath)
	assert.Equal(t, uint32(5), session2.DurationSeconds)
	assert.False(t, session2.IsBounce)
	assert.Nil(t, ua2)
	assert.Equal(t, uint64(0), pageView2.ClientID)
	assert.NotZero(t, pageView2.VisitorID)
	assert.Equal(t, "/test/path", pageView2.Path)
	assert.Equal(t, uint32(5), pageView2.DurationSeconds)
}

func TestHitFromRequestOverwrite(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, session, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		URL:          "http://bar.foo/new/custom/path?query=param&foo=bar#anchor",
	})

	if session[0].Path != "/new/custom/path" {
		t.Fatalf("Session not as expected: %v", session)
	}
}

func TestHitFromRequestOverwritePathAndReferrer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, session, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		URL:          "http://bar.foo/overwrite/this?query=param&foo=bar#anchor",
		Path:         "/new/custom/path",
		Referrer:     "http://custom.ref/",
	})

	if session[0].Path != "/new/custom/path" || session[0].Referrer != "http://custom.ref" {
		t.Fatalf("Session not as expected: %v", session)
	}
}

func TestHitFromRequestScreenSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	_, sessions, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ScreenWidth:  0,
		ScreenHeight: 400,
	})

	if sessions[0].ScreenWidth != 0 || sessions[0].ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", sessions[0].ScreenWidth, sessions[0].ScreenHeight)
	}

	_, sessions, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ScreenWidth:  400,
		ScreenHeight: 0,
	})

	if sessions[0].ScreenWidth != 0 || sessions[0].ScreenHeight != 0 {
		t.Fatalf("Screen size must be 0, but was: %v %v", sessions[0].ScreenWidth, sessions[0].ScreenHeight)
	}

	_, sessions, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: NewSessionCacheMem(dbClient, 100),
		ScreenWidth:  640,
		ScreenHeight: 1024,
	})

	if sessions[0].ScreenWidth != 640 || sessions[0].ScreenHeight != 1024 {
		t.Fatalf("Screen size must be set, but was: %v %v", sessions[0].ScreenWidth, sessions[0].ScreenHeight)
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
	_, sessions, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
		geoDB:        geoDB,
	})
	assert.Equal(t, "gb", sessions[0].CountryCode)
	assert.Equal(t, "London", sessions[0].City)
	req = httptest.NewRequest(http.MethodGet, "http://foo.bar/test/path?query=param&foo=bar#anchor", nil)
	req.RemoteAddr = "127.0.0.1"
	_, sessions, _ = HitFromRequest(req, "salt", &HitOptions{
		SessionCache: sessionCache,
		geoDB:        geoDB,
	})
	assert.Empty(t, sessions[0].CountryCode)
	assert.Empty(t, sessions[0].City)
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
	_, sessions, _ := HitFromRequest(req, "salt", options)
	assert.Len(t, sessions, 1)
	at := sessions[0].Time
	ExtendSession(req, "salt", options)
	session := sessionCache.Get(0, sessions[0].VisitorID, time.Now().UTC().Add(-time.Second))
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
