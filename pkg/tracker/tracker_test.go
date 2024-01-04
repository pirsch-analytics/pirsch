package tracker

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/analyzer"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ip"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/session"
	"github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ua"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"
)

const (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:105.0) Gecko/20100101 Firefox/105.0"
)

func TestTracker(t *testing.T) {
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Add("User-Agent", userAgent)
	tracker.PageView(req, 123, Options{})
	time.Sleep(time.Second)
	tracker.Event(req, 123, EventOptions{Name: "test"}, Options{})
	time.Sleep(time.Second * 2)
	req = httptest.NewRequest(http.MethodGet, "/bar", nil)
	req.Header.Add("User-Agent", userAgent)
	tracker.PageView(req, 123, Options{})
	time.Sleep(time.Second)
	tracker.ExtendSession(req, 123, Options{})
	tracker.Stop()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	events := client.GetEvents()
	assert.Len(t, sessions, 7)
	assert.Len(t, pageViews, 2)
	assert.Len(t, events, 1)
	assert.Equal(t, sessions[6].VisitorID, pageViews[0].VisitorID)
	assert.Equal(t, sessions[6].VisitorID, pageViews[1].VisitorID)
	assert.Equal(t, sessions[6].VisitorID, events[0].VisitorID)
	assert.Equal(t, sessions[6].SessionID, pageViews[0].SessionID)
	assert.Equal(t, sessions[6].SessionID, pageViews[1].SessionID)
	assert.Equal(t, sessions[6].SessionID, events[0].SessionID)

	assert.Equal(t, uint16(2), sessions[6].PageViews)
	assert.Equal(t, "/foo", sessions[6].EntryPath)
	assert.Equal(t, "/bar", sessions[6].ExitPath)
	assert.Equal(t, uint32(4), sessions[6].DurationSeconds)

	assert.Equal(t, "/foo", pageViews[0].Path)
	assert.Equal(t, "/bar", pageViews[1].Path)

	assert.Equal(t, "test", events[0].Name)
	assert.Equal(t, "/foo", events[0].Path)
}

func TestTracker_PageView(t *testing.T) {
	now := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../test/GeoIP2-City-Test.mmdb"))
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
	})
	tracker.PageView(req, 123, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		Tags: map[string]string{
			"author": "John",
			"type":   "blog_post",
		},
	})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 1)
	assert.Len(t, pageViews, 1)
	assert.Equal(t, sessions[0].VisitorID, pageViews[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, pageViews[0].SessionID)

	assert.Equal(t, int8(1), sessions[0].Sign)
	assert.Equal(t, uint64(123), sessions[0].ClientID)
	assert.True(t, sessions[0].Time.After(now))
	assert.True(t, sessions[0].Start.After(now))
	assert.Equal(t, uint32(0), sessions[0].DurationSeconds)
	assert.Equal(t, "/foo/bar", sessions[0].EntryPath)
	assert.Equal(t, "Foo", sessions[0].EntryTitle)
	assert.Equal(t, "/foo/bar", sessions[0].ExitPath)
	assert.Equal(t, "Foo", sessions[0].ExitTitle)
	assert.Equal(t, uint16(1), sessions[0].PageViews)
	assert.True(t, sessions[0].IsBounce)
	assert.Equal(t, "fr", sessions[0].Language)
	assert.Equal(t, "gb", sessions[0].CountryCode)
	assert.Equal(t, "London", sessions[0].City)
	assert.Equal(t, "https://google.com", sessions[0].Referrer)
	assert.Equal(t, "Google", sessions[0].ReferrerName)
	assert.Equal(t, pkg.OSLinux, sessions[0].OS)
	assert.Empty(t, sessions[0].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, sessions[0].Browser)
	assert.Equal(t, "105.0", sessions[0].BrowserVersion)
	assert.True(t, sessions[0].Desktop)
	assert.False(t, sessions[0].Mobile)
	assert.Equal(t, "Full HD", sessions[0].ScreenClass)
	assert.Equal(t, "Source", sessions[0].UTMSource)
	assert.Equal(t, "Medium", sessions[0].UTMMedium)
	assert.Equal(t, "Campaign", sessions[0].UTMCampaign)
	assert.Equal(t, "Content", sessions[0].UTMContent)
	assert.Equal(t, "Term", sessions[0].UTMTerm)

	assert.Equal(t, uint64(123), pageViews[0].ClientID)
	assert.True(t, pageViews[0].Time.After(now))
	assert.Equal(t, uint32(0), pageViews[0].DurationSeconds)
	assert.Equal(t, "/foo/bar", pageViews[0].Path)
	assert.Equal(t, "Foo", pageViews[0].Title)
	assert.Equal(t, "fr", pageViews[0].Language)
	assert.Equal(t, "gb", pageViews[0].CountryCode)
	assert.Equal(t, "London", pageViews[0].City)
	assert.Equal(t, "https://google.com", pageViews[0].Referrer)
	assert.Equal(t, "Google", pageViews[0].ReferrerName)
	assert.Equal(t, pkg.OSLinux, pageViews[0].OS)
	assert.Empty(t, pageViews[0].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, pageViews[0].Browser)
	assert.Equal(t, "105.0", pageViews[0].BrowserVersion)
	assert.True(t, pageViews[0].Desktop)
	assert.False(t, pageViews[0].Mobile)
	assert.Equal(t, "Full HD", pageViews[0].ScreenClass)
	assert.Equal(t, "Source", pageViews[0].UTMSource)
	assert.Equal(t, "Medium", pageViews[0].UTMMedium)
	assert.Equal(t, "Campaign", pageViews[0].UTMCampaign)
	assert.Equal(t, "Content", pageViews[0].UTMContent)
	assert.Equal(t, "Term", pageViews[0].UTMTerm)
	assert.Contains(t, pageViews[0].TagKeys, "author")
	assert.Contains(t, pageViews[0].TagKeys, "type")
	assert.Contains(t, pageViews[0].TagValues, "John")
	assert.Contains(t, pageViews[0].TagValues, "blog_post")

	time.Sleep(time.Second)
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Add("User-Agent", userAgent)
	req.RemoteAddr = "81.2.69.142"
	tracker.PageView(req, 123, Options{
		Title: "Bar",
	})
	tracker.Flush()
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	assert.Len(t, sessions, 3)
	assert.Len(t, pageViews, 2)
	assert.Equal(t, int8(1), sessions[0].Sign)
	assert.Equal(t, int8(-1), sessions[1].Sign)
	assert.Equal(t, int8(1), sessions[2].Sign)

	assert.Equal(t, uint32(1), sessions[2].DurationSeconds)
	assert.Equal(t, "/foo/bar", sessions[2].EntryPath)
	assert.Equal(t, "Foo", sessions[2].EntryTitle)
	assert.Equal(t, "/test", sessions[2].ExitPath)
	assert.Equal(t, "Bar", sessions[2].ExitTitle)
	assert.Equal(t, uint16(2), sessions[2].PageViews)
	assert.False(t, sessions[2].IsBounce)
	assert.Equal(t, "fr", sessions[2].Language)
	assert.Equal(t, "gb", sessions[2].CountryCode)
	assert.Equal(t, "London", sessions[2].City)
	assert.Equal(t, "https://google.com", sessions[2].Referrer)
	assert.Equal(t, "Google", sessions[2].ReferrerName)
	assert.Equal(t, pkg.OSLinux, sessions[2].OS)
	assert.Empty(t, sessions[2].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, sessions[2].Browser)
	assert.Equal(t, "105.0", sessions[2].BrowserVersion)
	assert.True(t, sessions[2].Desktop)
	assert.False(t, sessions[2].Mobile)
	assert.Equal(t, "Full HD", sessions[2].ScreenClass)
	assert.Equal(t, "Source", sessions[2].UTMSource)
	assert.Equal(t, "Medium", sessions[2].UTMMedium)
	assert.Equal(t, "Campaign", sessions[2].UTMCampaign)
	assert.Equal(t, "Content", sessions[2].UTMContent)
	assert.Equal(t, "Term", sessions[2].UTMTerm)

	assert.Equal(t, uint64(123), pageViews[1].ClientID)
	assert.True(t, pageViews[1].Time.After(now))
	assert.Equal(t, uint32(1), pageViews[1].DurationSeconds)
	assert.Equal(t, "/test", pageViews[1].Path)
	assert.Equal(t, "Bar", pageViews[1].Title)
	assert.Equal(t, "fr", pageViews[1].Language)
	assert.Equal(t, "gb", pageViews[1].CountryCode)
	assert.Equal(t, "London", pageViews[1].City)
	assert.Equal(t, "https://google.com", pageViews[1].Referrer)
	assert.Equal(t, "Google", pageViews[1].ReferrerName)
	assert.Equal(t, pkg.OSLinux, pageViews[1].OS)
	assert.Empty(t, pageViews[1].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, pageViews[1].Browser)
	assert.Equal(t, "105.0", pageViews[1].BrowserVersion)
	assert.True(t, pageViews[1].Desktop)
	assert.False(t, pageViews[1].Mobile)
	assert.Equal(t, "Full HD", pageViews[1].ScreenClass)
	assert.Equal(t, "Source", pageViews[1].UTMSource)
	assert.Equal(t, "Medium", pageViews[1].UTMMedium)
	assert.Equal(t, "Campaign", pageViews[1].UTMCampaign)
	assert.Equal(t, "Content", pageViews[1].UTMContent)
	assert.Equal(t, "Term", pageViews[1].UTMTerm)

	userAgents := client.GetUserAgents()
	assert.Len(t, userAgents, 1)
	assert.Equal(t, userAgent, userAgents[0].UserAgent)
}

func TestTracker_PageViewBounce(t *testing.T) {
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	tracker.PageView(req, 123, Options{})
	tracker.PageView(req, 123, Options{})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 3)
	assert.Len(t, pageViews, 1)
	assert.Equal(t, uint16(1), sessions[0].PageViews)
	assert.Equal(t, uint16(1), sessions[1].PageViews)
	assert.Equal(t, uint16(1), sessions[2].PageViews)
	assert.Equal(t, uint32(0), sessions[2].DurationSeconds)
}

func TestTracker_PageViewReferrerIgnorePath(t *testing.T) {
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", "https://google.com")
	tracker.PageView(req, 0, Options{})
	req = httptest.NewRequest(http.MethodGet, "https://example.com/bar", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", "https://example.com/foo")
	tracker.PageView(req, 0, Options{})
	tracker.Stop()
	sessions := client.GetSessions()
	assert.Len(t, sessions, 3)
	assert.Equal(t, "https://google.com", sessions[0].Referrer)
	assert.Equal(t, "https://google.com", sessions[1].Referrer)
	assert.Equal(t, "https://google.com", sessions[2].Referrer)
}

func TestTracker_PageViewReferrerOverwriteIgnorePath(t *testing.T) {
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo", nil)
	req.Header.Add("User-Agent", userAgent)
	tracker.PageView(req, 0, Options{Referrer: "https://google.com"})
	req = httptest.NewRequest(http.MethodGet, "https://example.com/bar", nil)
	req.Header.Add("User-Agent", userAgent)
	tracker.PageView(req, 0, Options{Referrer: "https://example.com/foo"})
	tracker.Stop()
	sessions := client.GetSessions()
	assert.Len(t, sessions, 3)
	assert.Equal(t, "https://google.com", sessions[0].Referrer)
	assert.Equal(t, "https://google.com", sessions[1].Referrer)
	assert.Equal(t, "https://google.com", sessions[2].Referrer)
}

func TestTracker_PageViewTimeout(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store:         client,
		WorkerTimeout: time.Millisecond * 100,
	})
	tracker.PageView(req, 123, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
	})
	assert.Len(t, client.GetSessions(), 0)
	assert.Len(t, client.GetPageViews(), 0)
	time.Sleep(time.Millisecond * 110)
	assert.Len(t, client.GetSessions(), 1)
	assert.Len(t, client.GetPageViews(), 1)
}

func TestTracker_PageViewBuffer(t *testing.T) {
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store:            client,
		Worker:           1,
		WorkerBufferSize: 5,
	})

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/foo/bar/%d?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", i), nil)
		req.Header.Add("User-Agent", userAgent)
		req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
		req.Header.Set("Referer", "https://google.com")
		req.RemoteAddr = "81.2.69.142"
		tracker.PageView(req, 123, Options{})
	}

	time.Sleep(time.Millisecond * 20)
	assert.Len(t, client.GetSessions(), 7)
	assert.Len(t, client.GetPageViews(), 4)
	tracker.Stop()
	assert.Len(t, client.GetSessions(), 13)
	assert.Len(t, client.GetPageViews(), 7)
}

func TestTracker_PageViewOverwriteTime(t *testing.T) {
	now := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.PageView(req, 123, Options{
		Time: now.Add(-time.Second * 20),
	})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 1)
	assert.Len(t, pageViews, 1)
	assert.Equal(t, sessions[0].VisitorID, pageViews[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, pageViews[0].SessionID)
	assert.True(t, now.After(sessions[0].Time))
	assert.True(t, now.After(pageViews[0].Time))
}

func TestTracker_PageViewFindSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	cache := session.NewMemCache(client, 10)
	tracker := NewTracker(Config{
		Store:        client,
		SessionCache: cache,
	})
	tracker.PageView(req, 123, Options{})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 1)
	assert.Len(t, pageViews, 1)
	cachedSessions := cache.Sessions()
	var cachedSession model.Session
	var cachedKey string

	for key, value := range cachedSessions {
		cachedSession = value
		cachedKey = key
	}

	cachedSession.Time = time.Now().UTC().Add(time.Hour * -4)
	cachedSessions[cachedKey] = cachedSession
	tracker.PageView(req, 123, Options{})
	tracker.Flush()
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	assert.Len(t, sessions, 2)
	assert.Len(t, pageViews, 2)
}

func TestTracker_PageViewFindSessionRedis(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	cache := session.NewRedisCache(time.Minute*30, nil, &redis.Options{
		Addr: "localhost:6379",
	})
	tracker := NewTracker(Config{
		Store:        client,
		SessionCache: cache,
	})
	tracker.PageView(req, 123, Options{})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 1)
	assert.Len(t, pageViews, 1)
	cache.Put(123, tracker.fingerprint(tracker.config.Salt, userAgent, "81.2.69.142", time.Now().UTC()), &model.Session{
		Time: time.Now().UTC().Add(time.Hour * -4),
	})
	tracker.PageView(req, 123, Options{})
	tracker.Flush()
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	assert.Len(t, sessions, 2)
	assert.Len(t, pageViews, 2)
}

func TestTracker_Event(t *testing.T) {
	now := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../test/GeoIP2-City-Test.mmdb"))
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
	})
	tracker.Event(req, 123, EventOptions{
		Name:     "event",
		Duration: 42,
		Meta:     map[string]string{"key0": "value0", "key1": "value1"},
	}, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
	})
	tracker.Flush()
	sessions := client.GetSessions()
	events := client.GetEvents()
	assert.Len(t, sessions, 1)
	assert.Len(t, client.GetPageViews(), 1)
	assert.Len(t, events, 1)
	assert.Equal(t, sessions[0].VisitorID, events[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, events[0].SessionID)
	assert.Equal(t, uint16(1), sessions[0].PageViews)

	assert.Equal(t, uint64(123), events[0].ClientID)
	assert.True(t, events[0].Time.After(now))
	assert.Equal(t, "event", events[0].Name)
	assert.Len(t, events[0].MetaKeys, 2)
	assert.Len(t, events[0].MetaValues, 2)
	assert.Equal(t, uint32(42), events[0].DurationSeconds)
	assert.Equal(t, "/foo/bar", events[0].Path)
	assert.Equal(t, "Foo", events[0].Title)
	assert.Equal(t, "fr", events[0].Language)
	assert.Equal(t, "gb", events[0].CountryCode)
	assert.Equal(t, "London", events[0].City)
	assert.Equal(t, "https://google.com", events[0].Referrer)
	assert.Equal(t, "Google", events[0].ReferrerName)
	assert.Equal(t, pkg.OSLinux, events[0].OS)
	assert.Empty(t, events[0].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, events[0].Browser)
	assert.Equal(t, "105.0", events[0].BrowserVersion)
	assert.True(t, events[0].Desktop)
	assert.False(t, events[0].Mobile)
	assert.Equal(t, "Full HD", events[0].ScreenClass)
	assert.Equal(t, "Source", events[0].UTMSource)
	assert.Equal(t, "Medium", events[0].UTMMedium)
	assert.Equal(t, "Campaign", events[0].UTMCampaign)
	assert.Equal(t, "Content", events[0].UTMContent)
	assert.Equal(t, "Term", events[0].UTMTerm)

	time.Sleep(time.Second)
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Add("User-Agent", userAgent)
	req.RemoteAddr = "81.2.69.142"
	tracker.Event(req, 123, EventOptions{
		Name: "event2",
	}, Options{
		Title: "Bar",
	})
	tracker.Flush()
	sessions = client.GetSessions()
	events = client.GetEvents()
	assert.Len(t, sessions, 3)
	assert.Len(t, client.GetPageViews(), 2)
	assert.Len(t, events, 2)
	assert.Equal(t, int8(1), sessions[0].Sign)
	assert.Equal(t, int8(-1), sessions[1].Sign)
	assert.Equal(t, int8(1), sessions[2].Sign)
	assert.Equal(t, uint16(1), sessions[0].PageViews)

	assert.Equal(t, uint64(123), events[1].ClientID)
	assert.True(t, events[1].Time.After(now))
	assert.Equal(t, "event2", events[1].Name)
	assert.Len(t, events[1].MetaKeys, 0)
	assert.Len(t, events[1].MetaValues, 0)
	assert.Zero(t, events[1].DurationSeconds)
	assert.Equal(t, "/test", events[1].Path)
	assert.Equal(t, "Bar", events[1].Title)
	assert.Equal(t, "fr", events[1].Language)
	assert.Equal(t, "gb", events[1].CountryCode)
	assert.Equal(t, "London", events[1].City)
	assert.Equal(t, "https://google.com", events[1].Referrer)
	assert.Equal(t, "Google", events[1].ReferrerName)
	assert.Equal(t, pkg.OSLinux, events[1].OS)
	assert.Empty(t, events[1].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, events[1].Browser)
	assert.Equal(t, "105.0", events[1].BrowserVersion)
	assert.True(t, events[1].Desktop)
	assert.False(t, events[1].Mobile)
	assert.Equal(t, "Full HD", events[1].ScreenClass)
	assert.Equal(t, "Source", events[1].UTMSource)
	assert.Equal(t, "Medium", events[1].UTMMedium)
	assert.Equal(t, "Campaign", events[1].UTMCampaign)
	assert.Equal(t, "Content", events[1].UTMContent)
	assert.Equal(t, "Term", events[1].UTMTerm)

	userAgents := client.GetUserAgents()
	assert.Len(t, userAgents, 1)
	assert.Equal(t, userAgent, userAgents[0].UserAgent)
}

func TestTracker_EventDiscard(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.Event(req, 123, EventOptions{}, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
	})
	tracker.Flush()
	assert.Len(t, client.GetSessions(), 0)
	assert.Len(t, client.GetPageViews(), 0)
	assert.Len(t, client.GetEvents(), 0)
}

func TestTracker_EventOverwriteTime(t *testing.T) {
	now := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.Event(req, 123, EventOptions{
		Name: "event",
	}, Options{
		Time: now.Add(-time.Second * 20),
	})
	tracker.Flush()
	sessions := client.GetSessions()
	events := client.GetEvents()
	assert.Len(t, sessions, 1)
	assert.Len(t, events, 1)
	assert.Equal(t, sessions[0].VisitorID, events[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, events[0].SessionID)
	assert.True(t, now.After(sessions[0].Time))
	assert.True(t, now.After(events[0].Time))
}

func TestTracker_EventPageView(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/event/page", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.Event(req, 123, EventOptions{
		Name:     "event",
		Duration: 42,
		Meta:     map[string]string{"key0": "value0", "key1": "value1"},
	}, Options{})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	events := client.GetEvents()
	assert.Len(t, sessions, 1)
	assert.Len(t, pageViews, 1)
	assert.Len(t, events, 1)
	assert.Equal(t, uint16(1), sessions[0].PageViews)
	assert.True(t, sessions[0].IsBounce)
	assert.Equal(t, "/event/page", pageViews[0].Path)
	assert.Equal(t, "event", events[0].Name)
	assert.Equal(t, "/event/page", events[0].Path)

	// do not track a new page view if another event is triggered on the same page
	// no longer count as bounced
	tracker.Event(req, 123, EventOptions{
		Name: "event",
	}, Options{})
	tracker.Flush()
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	events = client.GetEvents()
	assert.Len(t, sessions, 3) // first + cancel + new
	assert.Len(t, pageViews, 1)
	assert.Len(t, events, 2)
	assert.Equal(t, uint16(1), sessions[0].PageViews)
	assert.Equal(t, uint16(1), sessions[1].PageViews)
	assert.Equal(t, uint16(1), sessions[2].PageViews)
	assert.True(t, sessions[0].IsBounce)
	assert.True(t, sessions[1].IsBounce)
	assert.False(t, sessions[2].IsBounce)
	assert.Equal(t, "/event/page", pageViews[0].Path)
	assert.Equal(t, "event", events[0].Name)
	assert.Equal(t, "event", events[1].Name)
	assert.Equal(t, "/event/page", events[0].Path)
	assert.Equal(t, "/event/page", events[1].Path)

	// track new page view if event is triggered on a different page
	req = httptest.NewRequest(http.MethodGet, "/new/event/page", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	tracker.Event(req, 123, EventOptions{
		Name: "event",
	}, Options{})
	tracker.Flush()
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	events = client.GetEvents()
	assert.Len(t, sessions, 5) // cancel + new
	assert.Len(t, pageViews, 2)
	assert.Len(t, events, 3)
	assert.Equal(t, uint16(1), sessions[0].PageViews)
	assert.Equal(t, uint16(1), sessions[1].PageViews)
	assert.Equal(t, uint16(1), sessions[2].PageViews)
	assert.Equal(t, uint16(1), sessions[3].PageViews)
	assert.Equal(t, uint16(2), sessions[4].PageViews)
	assert.True(t, sessions[0].IsBounce)
	assert.True(t, sessions[1].IsBounce)
	assert.False(t, sessions[2].IsBounce)
	assert.False(t, sessions[3].IsBounce)
	assert.False(t, sessions[4].IsBounce)
	assert.Equal(t, "/event/page", pageViews[0].Path)
	assert.Equal(t, "/new/event/page", pageViews[1].Path)
	assert.Equal(t, "event", events[0].Name)
	assert.Equal(t, "event", events[1].Name)
	assert.Equal(t, "event", events[2].Name)
	assert.Equal(t, "/event/page", events[0].Path)
	assert.Equal(t, "/event/page", events[1].Path)
	assert.Equal(t, "/new/event/page", events[2].Path)
}

func TestTracker_ExtendSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.PageView(req, 123, Options{})
	time.Sleep(time.Second * 2)
	tracker.ExtendSession(req, 123, Options{})
	tracker.Flush()
	sessions := client.GetSessions()
	assert.Len(t, sessions, 3)
	assert.Equal(t, uint32(2), sessions[2].DurationSeconds)
	assert.Equal(t, uint16(1), sessions[2].Extended)
	tracker.ExtendSession(req, 123, Options{})
	tracker.Flush()
	sessions = client.GetSessions()
	assert.Len(t, sessions, 5)
	assert.Equal(t, uint16(2), sessions[4].Extended)
}

func TestTracker_ExtendSessionNoSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.ExtendSession(req, 123, Options{}) // do not create sessions using this method
	tracker.ExtendSession(req, 123, Options{})
	tracker.Flush()
	assert.Len(t, client.GetSessions(), 0)
}

func TestTracker_ExtendSessionOverwriteTime(t *testing.T) {
	now := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.PageView(req, 123, Options{
		Time: now.Add(-time.Second * 20),
	})
	time.Sleep(time.Second * 2)
	tracker.ExtendSession(req, 123, Options{
		Time: now.Add(-time.Second * 10),
	})
	tracker.Flush()
	sessions := client.GetSessions()
	assert.Len(t, sessions, 3)
	assert.Equal(t, uint32(10), sessions[2].DurationSeconds)
	assert.Equal(t, uint16(1), sessions[2].Extended)
	assert.True(t, now.After(sessions[0].Time))
	assert.True(t, now.After(sessions[1].Time))
	assert.True(t, now.After(sessions[2].Time))
}

func TestTracker_Flush(t *testing.T) {
	db.CleanupDB(t, dbClient)
	tracker := NewTracker(Config{
		Store:        dbClient,
		SessionCache: session.NewMemCache(dbClient, 100),
	})

	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		req.RemoteAddr = "187.65.23.54"
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:79.0) Gecko/20220101 Firefox/100.0")
		req.Header.Set("Accept-Language", "en")
		go tracker.Event(req, 0, EventOptions{Name: "event"}, Options{})
	}

	time.Sleep(time.Second)
	tracker.Flush()
	count, err := dbClient.Count(context.Background(), `SELECT count(*) FROM event`)
	assert.NoError(t, err)
	assert.Equal(t, 10, count)
}

func TestTrackerBots(t *testing.T) {
	store := db.NewClientMock()
	tracker := NewTracker(Config{
		Store:        store,
		SessionCache: session.NewMemCache(store, 100),
	})

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/path", nil)
		req.RemoteAddr = "187.65.23.54"
		req.Header.Set("User-Agent", "Bot")
		req.Header.Set("Accept-Language", "en")
		go tracker.PageView(req, 42, Options{})
		time.Sleep(time.Millisecond * 5)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/event/path", nil)
	req.RemoteAddr = "187.65.23.54"
	req.Header.Set("User-Agent", "Event Bot")
	req.Header.Set("Accept-Language", "en")
	go tracker.Event(req, 42, EventOptions{Name: "event"}, Options{})

	time.Sleep(time.Second)
	tracker.Flush()
	pageViews := store.GetPageViews()
	assert.Len(t, pageViews, 0)
	bots := store.GetBots()
	assert.Len(t, bots, 4)
	assert.Equal(t, uint64(42), bots[0].ClientID)
	assert.Equal(t, uint64(42), bots[1].ClientID)
	assert.Equal(t, uint64(42), bots[2].ClientID)
	assert.Equal(t, uint64(42), bots[3].ClientID)
	assert.NotZero(t, bots[0].VisitorID)
	assert.NotZero(t, bots[1].VisitorID)
	assert.NotZero(t, bots[2].VisitorID)
	assert.NotZero(t, bots[3].VisitorID)
	assert.Equal(t, "Bot", bots[0].UserAgent)
	assert.Equal(t, "Bot", bots[1].UserAgent)
	assert.Equal(t, "Bot", bots[2].UserAgent)
	assert.Equal(t, "Event Bot", bots[3].UserAgent)
	assert.Equal(t, "/path", bots[0].Path)
	assert.Equal(t, "/path", bots[1].Path)
	assert.Equal(t, "/path", bots[2].Path)
	assert.Equal(t, "/event/path", bots[3].Path)
	assert.Empty(t, bots[0].Event)
	assert.Empty(t, bots[1].Event)
	assert.Empty(t, bots[2].Event)
	assert.Equal(t, "event", bots[3].Event)
}

func TestTrackerPageViewAndEvent(t *testing.T) {
	pageViewsAndEvents := []struct {
		path              string
		timeOnPageSeconds int
		event             string
	}{
		{"/", 2, "a"},
		{"/foo", 3, "b"},
		{"/bar", 1, "c"},
	}

	client := db.Connect()
	defer db.Disconnect(client)
	db.CleanupDB(t, client)
	tracker := NewTracker(Config{
		Store: client,
	})

	for _, row := range pageViewsAndEvents {
		req := httptest.NewRequest(http.MethodGet, row.path, nil)
		req.Header.Add("User-Agent", userAgent)
		tracker.PageView(req, 0, Options{})
		time.Sleep(time.Duration(row.timeOnPageSeconds) * time.Second)
		tracker.Event(req, 0, EventOptions{Name: row.event, Duration: 99}, Options{})
		time.Sleep(time.Second)
	}

	tracker.Stop()
	time.Sleep(time.Second)
	a := analyzer.NewAnalyzer(client)
	sessionDuration, err := a.Time.AvgSessionDuration(&analyzer.Filter{
		From: util.Today(),
		To:   util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, sessionDuration, 1)
	assert.Equal(t, 8, sessionDuration[0].AverageTimeSpentSeconds)
	timeOnPage, err := a.Pages.ByPath(&analyzer.Filter{
		From:              util.Today(),
		To:                util.Today(),
		IncludeTimeOnPage: true,
	})
	assert.NoError(t, err)
	assert.Len(t, timeOnPage, 3)
	slices.SortFunc(timeOnPage, func(a, b model.PageStats) int {
		if a.Path > b.Path {
			return 1
		} else if a.Path < b.Path {
			return -1
		}

		return 0
	})
	assert.Equal(t, "/", timeOnPage[0].Path)
	assert.Equal(t, "/bar", timeOnPage[1].Path)
	assert.Equal(t, "/foo", timeOnPage[2].Path)
}

func TestTracker_ignorePrefetch(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("X-Moz", "prefetch")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with X-Moz header must be ignored")
	}

	req.Header.Del("X-Moz")
	req.Header.Set("X-Purpose", "prefetch")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with X-Purpose header must be ignored")
	}

	req.Header.Set("X-Purpose", "preview")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with X-Purpose header must be ignored")
	}

	req.Header.Del("X-Purpose")
	req.Header.Set("Purpose", "prefetch")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with Purpose header must be ignored")
	}

	req.Header.Set("Purpose", "preview")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with Purpose header must be ignored")
	}

	req.Header.Del("Purpose")

	if _, _, ignore := tracker.ignore(req); ignore {
		t.Fatal("Session must not be ignored")
	}
}

func TestTracker_ignoreUserAgent(t *testing.T) {
	userAgents := []struct {
		userAgent string
		ignore    bool
	}{
		{"This is a bot request", true},
		{"This is a crawler request", true},
		{"This is a spider request", true},
		{"Visit http://spam.com!", true},
		{"", true},
		{"172.22.0.11:30004", true},
		{"172.22.0.11", true},
		{"2345:0425:2CA1:0000:0000:0567:5673:23b5", true},
		{"2345:425:2CA1:0000:0000:567:5673:23b5", true},
		{"2345:0425:2CA1:0:0:0567:5673:23b5", true},
		{"[2345:0425:2CA1:0:0:0567:5673:23b5]:8080", true},
		{userAgent, false},
	}

	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for _, userAgent := range userAgents {
		req.Header.Set("User-Agent", userAgent.userAgent)

		if _, _, ignore := tracker.ignore(req); ignore != userAgent.ignore {
			if userAgent.ignore {
				t.Fatalf("Request with User-Agent '%s' must be ignored", userAgent.userAgent)
			} else {
				t.Fatalf("Request with User-Agent '%s' must not be ignored", userAgent.userAgent)
			}
		}
	}
}

func TestTracker_ignoreBotUserAgent(t *testing.T) {
	tracker := NewTracker(Config{})

	for _, botUserAgent := range ua.Blacklist {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", botUserAgent)

		if _, _, ignore := tracker.ignore(req); !ignore {
			t.Fatalf("Request with user agent '%v' must have been ignored", botUserAgent)
		}
	}

	botUserAgent := []string{
		"-1' OR 2+990-990-1=0+0+0+1 or 'J2HdM1AB'='",
		"0'XOR(if(now()=sysdate(),sleep(15),0))XOR'Z",
		"1 waitfor delay '0:0:15' --",
		"14wpthYh' OR 294=(SELECT 294 FROM PG_SLEEP(15))--",
		"{{2959082-1}}",
		"{{ 2959082-1 }}",
		"7144de67-08ee-4fce-9997-49ef5af582d8",
		"a9c4b36c-71e8-4cdf-81a8-e178edcc7f30",
		"5b49398f-26bd-4bc3-ab16-3ca223d4d218",
	}

	for _, userAgent := range botUserAgent {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", userAgent)

		if _, _, ignore := tracker.ignore(req); !ignore {
			t.Fatalf("Request with user agent '%v' must have been ignored", botUserAgent)
		}
	}
}

func TestTracker_ignoreReferrer(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "ua")
	req.Header.Set("Referer", "2your.site")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}

	req.Header.Set("Referer", "subdomain.2your.site")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request for subdomain must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/?ref=2your.site", nil)

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}
}

func TestTracker_ignoreBrowserVersion(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.4147.135 Safari/537.36")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", userAgent)

	if _, _, ignore := tracker.ignore(req); ignore {
		t.Fatal("Request must not have been ignored")
	}
}

func TestTracker_ignoreDoNotTrack(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", userAgent)

	if _, _, ignore := tracker.ignore(req); ignore {
		t.Fatal("Request must not have been ignored")
	}

	req.Header.Set("DNT", "1")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}
}

func TestTracker_ignoreIP(t *testing.T) {
	filter := ip.NewUdger("", "")
	filter.Update([]string{"90.154.29.38"}, []string{}, []ip.Range{}, []ip.Range{})
	tracker := NewTracker(Config{
		IPFilter: filter,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", userAgent)

	if _, _, ignore := tracker.ignore(req); ignore {
		t.Fatal("Request must not have been ignored")
	}

	req.RemoteAddr = "90.154.29.38"

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}
}

func TestTracker_ignorePageViews(t *testing.T) {
	client := db.NewClientMock()
	cache := session.NewMemCache(client, 10)
	tracker := NewTracker(Config{
		Store:        client,
		SessionCache: cache,
		MaxPageViews: 5,
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%d", i), nil)
		req.Header.Set("User-Agent", userAgent)
		tracker.PageView(req, 0, Options{})
		time.Sleep(time.Millisecond * 5)
	}

	tracker.Stop()
	assert.Len(t, client.GetPageViews(), 5)
}

func TestTracker_getLanguage(t *testing.T) {
	input := []string{
		"",
		"  \t ",
		"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5",
		"en-us, en",
		"en-gb, en",
		"invalid",
	}
	expected := []string{
		"",
		"",
		"fr",
		"en",
		"en",
		"",
	}
	tracker := NewTracker(Config{})

	for i, in := range input {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", in)

		if lang := tracker.getLanguage(req); lang != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], lang)
		}
	}
}

func TestTracker_getScreenClass(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Sec-CH-Width", "1920")
	req.Header.Set("Sec-CH-Viewport-Width", "1919")
	tracker := NewTracker(Config{})
	assert.Equal(t, "XS", tracker.getScreenClass(req, 42))
	assert.Equal(t, "XL", tracker.getScreenClass(req, 1024))
	assert.Equal(t, "XL", tracker.getScreenClass(req, 1025))
	assert.Equal(t, "HD", tracker.getScreenClass(req, 1919))
	assert.Equal(t, "Full HD", tracker.getScreenClass(req, 2559))
	assert.Equal(t, "WQHD", tracker.getScreenClass(req, 3839))
	assert.Equal(t, "UHD 4K", tracker.getScreenClass(req, 5119))
	assert.Equal(t, "UHD 5K", tracker.getScreenClass(req, 5120))
	assert.Equal(t, "Full HD", tracker.getScreenClass(req, 0))
	req.Header.Del("Sec-CH-Width")
	assert.Equal(t, "HD", tracker.getScreenClass(req, 0))
	req.Header.Del("Sec-CH-Viewport-Width")
	assert.Equal(t, "", tracker.getScreenClass(req, 0))
}

func TestTracker_referrerOrCampaignChanged(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Referer", "https://referrer.com")
	s := &model.Session{
		Referrer:     "https://referrer.com",
		ReferrerName: "referrer.com",
	}
	assert.False(t, tracker.referrerOrCampaignChanged(req, s, "", ""))
	s.Referrer = ""
	assert.True(t, tracker.referrerOrCampaignChanged(req, s, "", ""))
	s.Referrer = "https://referrer.com"
	req = httptest.NewRequest(http.MethodGet, "/test?ref=https://different.com", nil)
	assert.True(t, tracker.referrerOrCampaignChanged(req, s, "", ""))
	req = httptest.NewRequest(http.MethodGet, "/test?utm_source=Referrer", nil)
	assert.True(t, tracker.referrerOrCampaignChanged(req, s, "", ""))
	s.ReferrerName = "Referrer"
	s.UTMSource = "Referrer"
	assert.False(t, tracker.referrerOrCampaignChanged(req, s, "", ""))
	s = &model.Session{Referrer: "https://referrer.com"}
	req = httptest.NewRequest(http.MethodGet, "/test?ref=Referrer", nil)
	assert.True(t, tracker.referrerOrCampaignChanged(req, s, "", ""))
}
