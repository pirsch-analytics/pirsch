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
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"
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
	assert.Equal(t, uint16(1), sessions[0].Version)
	assert.Equal(t, uint16(1), sessions[1].Version)
	assert.Equal(t, uint16(2), sessions[2].Version)
	assert.Equal(t, uint16(2), sessions[3].Version)
	assert.Equal(t, uint16(3), sessions[4].Version)
	assert.Equal(t, uint16(3), sessions[5].Version)
	assert.Equal(t, uint16(4), sessions[6].Version)
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
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
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
	assert.Equal(t, "example.com", sessions[0].Hostname)
	assert.Equal(t, "/foo/bar", sessions[0].EntryPath)
	assert.Equal(t, "Foo", sessions[0].EntryTitle)
	assert.Equal(t, "/foo/bar", sessions[0].ExitPath)
	assert.Equal(t, "Foo", sessions[0].ExitTitle)
	assert.Equal(t, uint16(1), sessions[0].PageViews)
	assert.True(t, sessions[0].IsBounce)
	assert.Equal(t, "fr", sessions[0].Language)
	assert.Equal(t, "gb", sessions[0].CountryCode)
	assert.Equal(t, "England", sessions[0].Region)
	assert.Equal(t, "London", sessions[0].City)
	assert.Equal(t, "https://google.com", sessions[0].Referrer)
	assert.Equal(t, "Google", sessions[0].ReferrerName)
	assert.Equal(t, pkg.OSLinux, sessions[0].OS)
	assert.Empty(t, sessions[0].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, sessions[0].Browser)
	assert.Equal(t, "128.0", sessions[0].BrowserVersion)
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
	assert.Equal(t, "example.com", pageViews[0].Hostname)
	assert.Equal(t, "/foo/bar", pageViews[0].Path)
	assert.Equal(t, "Foo", pageViews[0].Title)
	assert.Equal(t, "fr", pageViews[0].Language)
	assert.Equal(t, "gb", pageViews[0].CountryCode)
	assert.Equal(t, "England", pageViews[0].Region)
	assert.Equal(t, "London", pageViews[0].City)
	assert.Equal(t, "https://google.com", pageViews[0].Referrer)
	assert.Equal(t, "Google", pageViews[0].ReferrerName)
	assert.Equal(t, pkg.OSLinux, pageViews[0].OS)
	assert.Empty(t, pageViews[0].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, pageViews[0].Browser)
	assert.Equal(t, "128.0", pageViews[0].BrowserVersion)
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
	assert.Equal(t, "example.com", sessions[2].Hostname)
	assert.Equal(t, "/foo/bar", sessions[2].EntryPath)
	assert.Equal(t, "Foo", sessions[2].EntryTitle)
	assert.Equal(t, "/test", sessions[2].ExitPath)
	assert.Equal(t, "Bar", sessions[2].ExitTitle)
	assert.Equal(t, uint16(2), sessions[2].PageViews)
	assert.False(t, sessions[2].IsBounce)
	assert.Equal(t, "fr", sessions[2].Language)
	assert.Equal(t, "gb", sessions[2].CountryCode)
	assert.Equal(t, "England", sessions[2].Region)
	assert.Equal(t, "London", sessions[2].City)
	assert.Equal(t, "https://google.com", sessions[2].Referrer)
	assert.Equal(t, "Google", sessions[2].ReferrerName)
	assert.Equal(t, pkg.OSLinux, sessions[2].OS)
	assert.Empty(t, sessions[2].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, sessions[2].Browser)
	assert.Equal(t, "128.0", sessions[2].BrowserVersion)
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
	assert.Equal(t, "example.com", pageViews[1].Hostname)
	assert.Equal(t, "/test", pageViews[1].Path)
	assert.Equal(t, "Bar", pageViews[1].Title)
	assert.Equal(t, "fr", pageViews[1].Language)
	assert.Equal(t, "gb", pageViews[1].CountryCode)
	assert.Equal(t, "England", pageViews[1].Region)
	assert.Equal(t, "London", pageViews[1].City)
	assert.Equal(t, "https://google.com", pageViews[1].Referrer)
	assert.Equal(t, "Google", pageViews[1].ReferrerName)
	assert.Equal(t, pkg.OSLinux, pageViews[1].OS)
	assert.Empty(t, pageViews[1].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, pageViews[1].Browser)
	assert.Equal(t, "128.0", pageViews[1].BrowserVersion)
	assert.True(t, pageViews[1].Desktop)
	assert.False(t, pageViews[1].Mobile)
	assert.Equal(t, "Full HD", pageViews[1].ScreenClass)
	assert.Equal(t, "Source", pageViews[1].UTMSource)
	assert.Equal(t, "Medium", pageViews[1].UTMMedium)
	assert.Equal(t, "Campaign", pageViews[1].UTMCampaign)
	assert.Equal(t, "Content", pageViews[1].UTMContent)
	assert.Equal(t, "Term", pageViews[1].UTMTerm)

	requests := client.GetRequests()
	assert.Len(t, requests, 1)
	assert.Empty(t, requests[0].IP)
	assert.Equal(t, userAgent, requests[0].UserAgent)
	assert.Equal(t, "https://google.com", requests[0].Referrer)
	assert.Equal(t, "Source", requests[0].UTMSource)
	assert.Equal(t, "Medium", requests[0].UTMMedium)
	assert.Equal(t, "Campaign", requests[0].UTMCampaign)
	assert.False(t, requests[0].Bot)
}

func TestTracker_PageViewHostnameEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.RemoteAddr = "81.2.69.142"
	req.Host = ""
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.PageView(req, 123, Options{
		Hostname: "example.com",
	})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 1)
	assert.Len(t, pageViews, 1)
	assert.Equal(t, sessions[0].VisitorID, pageViews[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, pageViews[0].SessionID)
	assert.Equal(t, "example.com", sessions[0].Hostname)
	assert.Equal(t, "example.com", pageViews[0].Hostname)
}

func TestTracker_PageViewHostnameOverwrite(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://foo.com/foo/bar", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.RemoteAddr = "81.2.69.142"
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})
	tracker.PageView(req, 123, Options{
		Hostname: "example.com",
	})
	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 1)
	assert.Len(t, pageViews, 1)
	assert.Equal(t, sessions[0].VisitorID, pageViews[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, pageViews[0].SessionID)
	assert.Equal(t, "example.com", sessions[0].Hostname)
	assert.Equal(t, "example.com", pageViews[0].Hostname)
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
	time.Sleep(time.Millisecond * 200)
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

	time.Sleep(time.Millisecond * 100)
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

func TestTracker_PageViewClientHints(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Sec-Ch-Ua", "\"Not A(Brand\";v=\"99\", \"Google Chrome\";v=\"121\", \"Chromium\";v=\"121\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Linux\"")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../test/GeoIP2-City-Test.mmdb"))
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
	})
	tracker.PageView(req, 123, Options{})
	tracker.Flush()
	sessions := client.GetSessions()
	assert.Len(t, sessions, 1)
	assert.Equal(t, "Chrome", sessions[0].Browser)
	assert.Equal(t, "121", sessions[0].BrowserVersion)
	assert.False(t, sessions[0].Mobile)
	assert.Equal(t, "Linux", sessions[0].OS)
}

func TestTracker_PageViewGclid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?gclid=xy123", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
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
	assert.Equal(t, "https://google.com", sessions[0].Referrer)
	assert.Equal(t, "Google", sessions[0].ReferrerName)
	assert.Empty(t, sessions[0].UTMSource)
	assert.Equal(t, "(gclid)", sessions[0].UTMMedium)
	assert.Empty(t, sessions[0].UTMCampaign)
	assert.Empty(t, sessions[0].UTMContent)
	assert.Empty(t, sessions[0].UTMTerm)
	assert.Equal(t, "https://google.com", pageViews[0].Referrer)
	assert.Equal(t, "Google", pageViews[0].ReferrerName)
	assert.Empty(t, pageViews[0].UTMSource)
	assert.Equal(t, "(gclid)", pageViews[0].UTMMedium)
	assert.Empty(t, pageViews[0].UTMCampaign)
	assert.Empty(t, pageViews[0].UTMContent)
	assert.Empty(t, pageViews[0].UTMTerm)

	// do not override utm_medium
	req = httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?utm_medium=Medium&gclid=xy123", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
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
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	assert.Len(t, sessions, 2)
	assert.Len(t, pageViews, 2)
	assert.Equal(t, sessions[1].VisitorID, pageViews[1].VisitorID)
	assert.Equal(t, sessions[1].SessionID, pageViews[1].SessionID)
	assert.Equal(t, "https://google.com", sessions[1].Referrer)
	assert.Equal(t, "Google", sessions[1].ReferrerName)
	assert.Equal(t, "Medium", sessions[1].UTMMedium)
	assert.Equal(t, "https://google.com", pageViews[1].Referrer)
	assert.Equal(t, "Google", pageViews[1].ReferrerName)
	assert.Equal(t, "Medium", pageViews[1].UTMMedium)
}

func TestTracker_PageViewMsclkid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?msclkid=xy123", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://bing.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
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
	assert.Equal(t, "https://bing.com", sessions[0].Referrer)
	assert.Equal(t, "Bing", sessions[0].ReferrerName)
	assert.Empty(t, sessions[0].UTMSource)
	assert.Equal(t, "(msclkid)", sessions[0].UTMMedium)
	assert.Empty(t, sessions[0].UTMCampaign)
	assert.Empty(t, sessions[0].UTMContent)
	assert.Empty(t, sessions[0].UTMTerm)
	assert.Equal(t, "https://bing.com", pageViews[0].Referrer)
	assert.Equal(t, "Bing", pageViews[0].ReferrerName)
	assert.Empty(t, pageViews[0].UTMSource)
	assert.Equal(t, "(msclkid)", pageViews[0].UTMMedium)
	assert.Empty(t, pageViews[0].UTMCampaign)
	assert.Empty(t, pageViews[0].UTMContent)
	assert.Empty(t, pageViews[0].UTMTerm)

	// do not override utm_medium
	req = httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?utm_medium=Medium&msclkid=xy123", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://bing.com")
	req.RemoteAddr = "81.2.69.142"
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
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	assert.Len(t, sessions, 2)
	assert.Len(t, pageViews, 2)
	assert.Equal(t, sessions[1].VisitorID, pageViews[1].VisitorID)
	assert.Equal(t, sessions[1].SessionID, pageViews[1].SessionID)
	assert.Equal(t, "https://bing.com", sessions[1].Referrer)
	assert.Equal(t, "Bing", sessions[1].ReferrerName)
	assert.Equal(t, "Medium", sessions[1].UTMMedium)
	assert.Equal(t, "https://bing.com", pageViews[1].Referrer)
	assert.Equal(t, "Bing", pageViews[1].ReferrerName)
	assert.Equal(t, "Medium", pageViews[1].UTMMedium)
}

func TestTracker_PageViewChannel(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com?utm_medium=paid", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://suche.aol.de")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
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
	assert.Equal(t, "https://suche.aol.de", sessions[0].Referrer)
	assert.Equal(t, "AOL", sessions[0].ReferrerName)
	assert.Equal(t, "paid", sessions[0].UTMMedium)
	assert.Equal(t, "Paid Search", sessions[0].Channel)
	assert.Equal(t, "https://suche.aol.de", pageViews[0].Referrer)
	assert.Equal(t, "AOL", pageViews[0].ReferrerName)
	assert.Equal(t, "paid", pageViews[0].UTMMedium)
	assert.Equal(t, "Paid Search", pageViews[0].Channel)

	req = httptest.NewRequest(http.MethodGet, "https://example.com?gclid=xyz123", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
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
	sessions = client.GetSessions()
	pageViews = client.GetPageViews()
	assert.Len(t, sessions, 2)
	assert.Len(t, pageViews, 2)
	assert.Equal(t, sessions[1].VisitorID, pageViews[1].VisitorID)
	assert.Equal(t, sessions[1].SessionID, pageViews[1].SessionID)
	assert.Equal(t, "https://google.com", sessions[1].Referrer)
	assert.Equal(t, "Google", sessions[1].ReferrerName)
	assert.Equal(t, "(gclid)", sessions[1].UTMMedium)
	assert.Equal(t, "Paid Search", sessions[1].Channel)
	assert.Equal(t, "https://google.com", pageViews[1].Referrer)
	assert.Equal(t, "Google", pageViews[1].ReferrerName)
	assert.Equal(t, "(gclid)", pageViews[1].UTMMedium)
	assert.Equal(t, "Paid Search", pageViews[1].Channel)
}

func TestTracker_PageViewSessionDurationAndTimeOnPage(t *testing.T) {
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
	})

	pages := []struct {
		path  string
		delay time.Duration
	}{
		{path: "/", delay: 9},
		{path: "/pricing", delay: 5},
		{path: "/pricing", delay: 3},
		{path: "/about", delay: 10},
		{path: "/", delay: 7},
		{path: "/order", delay: 0},
	}

	for _, pv := range pages {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("https://example.com/%s", pv.path), nil)
		req.Header.Add("User-Agent", userAgent)
		req.Header.Set("Accept-Language", "en;q=0.8, de;q=0.7, *;q=0.5")
		req.Header.Set("Referer", "https://google.com")
		req.RemoteAddr = "81.2.69.142"
		tracker.PageView(req, 123, Options{})
		time.Sleep(pv.delay * time.Second)
	}

	tracker.Flush()
	sessions := client.GetSessions()
	pageViews := client.GetPageViews()
	assert.Len(t, sessions, 11)
	assert.Len(t, pageViews, 6)
	assert.Equal(t, sessions[0].VisitorID, pageViews[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, pageViews[0].SessionID)
	assert.Equal(t, uint32(34), sessions[10].DurationSeconds)
	assert.Equal(t, uint32(0), pageViews[0].DurationSeconds)
	assert.Equal(t, uint32(9), pageViews[1].DurationSeconds)
	assert.Equal(t, uint32(5), pageViews[2].DurationSeconds)
	assert.Equal(t, uint32(3), pageViews[3].DurationSeconds)
	assert.Equal(t, uint32(10), pageViews[4].DurationSeconds)
	assert.Equal(t, uint32(7), pageViews[5].DurationSeconds)
}

func TestTracker_Event(t *testing.T) {
	now := time.Now()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../test/GeoIP2-City-Test.mmdb"))
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
		LogIP: true,
	})
	tracker.Event(req, 123, EventOptions{
		Name:     "event",
		Duration: 42,
		Meta:     map[string]string{"key0": "value0", "key1": "value1"},
	}, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		Tags:         map[string]string{"key0": "override", "key2": "value2"},
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
	assert.Len(t, events[0].MetaKeys, 3)
	assert.Len(t, events[0].MetaValues, 3)
	assert.Contains(t, events[0].MetaKeys, "key0")
	assert.Contains(t, events[0].MetaKeys, "key1")
	assert.Contains(t, events[0].MetaKeys, "key2")
	assert.Contains(t, events[0].MetaValues, "value0")
	assert.Contains(t, events[0].MetaValues, "value1")
	assert.Contains(t, events[0].MetaValues, "value2")
	assert.Equal(t, uint32(42), events[0].DurationSeconds)
	assert.Equal(t, "example.com", events[0].Hostname)
	assert.Equal(t, "/foo/bar", events[0].Path)
	assert.Equal(t, "Foo", events[0].Title)
	assert.Equal(t, "fr", events[0].Language)
	assert.Equal(t, "gb", events[0].CountryCode)
	assert.Equal(t, "England", events[0].Region)
	assert.Equal(t, "London", events[0].City)
	assert.Equal(t, "https://google.com", events[0].Referrer)
	assert.Equal(t, "Google", events[0].ReferrerName)
	assert.Equal(t, pkg.OSLinux, events[0].OS)
	assert.Empty(t, events[0].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, events[0].Browser)
	assert.Equal(t, "128.0", events[0].BrowserVersion)
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
	assert.Equal(t, "example.com", events[1].Hostname)
	assert.Equal(t, "/test", events[1].Path)
	assert.Equal(t, "Bar", events[1].Title)
	assert.Equal(t, "fr", events[1].Language)
	assert.Equal(t, "gb", events[1].CountryCode)
	assert.Equal(t, "England", events[1].Region)
	assert.Equal(t, "London", events[1].City)
	assert.Equal(t, "https://google.com", events[1].Referrer)
	assert.Equal(t, "Google", events[1].ReferrerName)
	assert.Equal(t, pkg.OSLinux, events[1].OS)
	assert.Empty(t, events[1].OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, events[1].Browser)
	assert.Equal(t, "128.0", events[1].BrowserVersion)
	assert.True(t, events[1].Desktop)
	assert.False(t, events[1].Mobile)
	assert.Equal(t, "Full HD", events[1].ScreenClass)
	assert.Equal(t, "Source", events[1].UTMSource)
	assert.Equal(t, "Medium", events[1].UTMMedium)
	assert.Equal(t, "Campaign", events[1].UTMCampaign)
	assert.Equal(t, "Content", events[1].UTMContent)
	assert.Equal(t, "Term", events[1].UTMTerm)

	requests := client.GetRequests()
	assert.Len(t, requests, 1)
	assert.Equal(t, "81.2.69.142", requests[0].IP)
	assert.Equal(t, userAgent, requests[0].UserAgent)
	assert.Equal(t, "https://google.com", requests[0].Referrer)
	assert.Equal(t, "Source", requests[0].UTMSource)
	assert.Equal(t, "Medium", requests[0].UTMMedium)
	assert.Equal(t, "Campaign", requests[0].UTMCampaign)
	assert.False(t, requests[0].Bot)
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

func TestTracker_EventClientHints(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Sec-Ch-Ua", "\"Not A(Brand\";v=\"99\", \"Google Chrome\";v=\"121\", \"Chromium\";v=\"121\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Linux\"")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../test/GeoIP2-City-Test.mmdb"))
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
	})
	tracker.PageView(req, 123, Options{})
	tracker.Event(req, 123, EventOptions{
		Name: "event",
	}, Options{})
	tracker.Flush()
	events := client.GetEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "Chrome", events[0].Browser)
	assert.Equal(t, "121", events[0].BrowserVersion)
	assert.False(t, events[0].Mobile)
	assert.Equal(t, "Linux", events[0].OS)
}

func TestTracker_EventChannel(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/?utm_medium=paid", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://suche.aol.de")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../test/GeoIP2-City-Test.mmdb"))
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
		LogIP: true,
	})
	tracker.Event(req, 123, EventOptions{
		Name:     "event",
		Duration: 42,
		Meta:     map[string]string{"key0": "value0", "key1": "value1"},
	}, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		Tags:         map[string]string{"key0": "override", "key2": "value2"},
	})
	tracker.Flush()
	sessions := client.GetSessions()
	events := client.GetEvents()
	assert.Len(t, sessions, 1)
	assert.Len(t, client.GetPageViews(), 1)
	assert.Len(t, events, 1)
	assert.Equal(t, sessions[0].VisitorID, events[0].VisitorID)
	assert.Equal(t, sessions[0].SessionID, events[0].SessionID)
	assert.Equal(t, "https://suche.aol.de", sessions[0].Referrer)
	assert.Equal(t, "AOL", sessions[0].ReferrerName)
	assert.Equal(t, "paid", sessions[0].UTMMedium)
	assert.Equal(t, "Paid Search", sessions[0].Channel)
	assert.Equal(t, "https://suche.aol.de", events[0].Referrer)
	assert.Equal(t, "AOL", events[0].ReferrerName)
	assert.Equal(t, "paid", events[0].UTMMedium)
	assert.Equal(t, "Paid Search", events[0].Channel)

	req = httptest.NewRequest(http.MethodGet, "https://example.com?gclid=xyz123", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	tracker.Event(req, 123, EventOptions{
		Name:     "event",
		Duration: 42,
		Meta:     map[string]string{"key0": "value0", "key1": "value1"},
	}, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
		Tags:         map[string]string{"key0": "override", "key2": "value2"},
	})
	tracker.Flush()
	sessions = client.GetSessions()
	events = client.GetEvents()
	assert.Len(t, sessions, 2)
	assert.Len(t, events, 2)
	assert.Equal(t, sessions[1].VisitorID, events[1].VisitorID)
	assert.Equal(t, sessions[1].SessionID, events[1].SessionID)
	assert.Equal(t, "https://google.com", sessions[1].Referrer)
	assert.Equal(t, "Google", sessions[1].ReferrerName)
	assert.Equal(t, "(gclid)", sessions[1].UTMMedium)
	assert.Equal(t, "Paid Search", sessions[1].Channel)
	assert.Equal(t, "https://google.com", events[1].Referrer)
	assert.Equal(t, "Google", events[1].ReferrerName)
	assert.Equal(t, "(gclid)", events[1].UTMMedium)
	assert.Equal(t, "Paid Search", events[1].Channel)
}

func TestTracker_ExtendSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
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
	assert.Equal(t, "example.com", sessions[2].Hostname)
	assert.Equal(t, uint32(2), sessions[2].DurationSeconds)
	assert.Equal(t, uint16(1), sessions[2].Extended)
	tracker.ExtendSession(req, 123, Options{})
	tracker.Flush()
	sessions = client.GetSessions()
	assert.Len(t, sessions, 5)
	assert.Equal(t, "example.com", sessions[4].Hostname)
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

func TestTracker_Accept(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, _ := geodb.NewGeoDB("", "", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../test/GeoIP2-City-Test.mmdb"))
	client := db.NewClientMock()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
	})
	s := tracker.Accept(req, 123, Options{})
	assert.NotNil(t, s)
	assert.Equal(t, "example.com", s.Hostname)
	assert.Equal(t, "fr", s.Language)
	assert.Equal(t, "Linux", s.OS)
	assert.Equal(t, "gb", s.CountryCode)
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

func TestTrackerRequests(t *testing.T) {
	store := db.NewClientMock()
	tracker := NewTracker(Config{
		Store:        store,
		SessionCache: session.NewMemCache(store, 100),
	})

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, "https://www.example.com/path", nil)
		req.RemoteAddr = "187.65.23.54"
		req.Header.Set("User-Agent", "Bot")
		req.Header.Set("Accept-Language", "en")
		go tracker.PageView(req, 42, Options{})
		time.Sleep(time.Millisecond * 5)
	}

	req, _ := http.NewRequest(http.MethodGet, "https://www.example.com/event/path", nil)
	req.RemoteAddr = "187.65.23.54"
	req.Header.Set("User-Agent", "Event Bot")
	req.Header.Set("Accept-Language", "en")
	go tracker.Event(req, 42, EventOptions{Name: "event"}, Options{})

	time.Sleep(time.Second)
	tracker.Flush()
	pageViews := store.GetPageViews()
	assert.Len(t, pageViews, 0)
	requests := store.GetRequests()
	assert.Len(t, requests, 4)
	assert.Equal(t, uint64(42), requests[0].ClientID)
	assert.Equal(t, uint64(42), requests[1].ClientID)
	assert.Equal(t, uint64(42), requests[2].ClientID)
	assert.Equal(t, uint64(42), requests[3].ClientID)
	assert.NotZero(t, requests[0].VisitorID)
	assert.NotZero(t, requests[1].VisitorID)
	assert.NotZero(t, requests[2].VisitorID)
	assert.NotZero(t, requests[3].VisitorID)
	assert.Equal(t, "Bot", requests[0].UserAgent)
	assert.Equal(t, "Bot", requests[1].UserAgent)
	assert.Equal(t, "Bot", requests[2].UserAgent)
	assert.Equal(t, "Event Bot", requests[3].UserAgent)
	assert.Equal(t, "example.com", requests[0].Hostname)
	assert.Equal(t, "example.com", requests[1].Hostname)
	assert.Equal(t, "example.com", requests[2].Hostname)
	assert.Equal(t, "example.com", requests[3].Hostname)
	assert.Equal(t, "/path", requests[0].Path)
	assert.Equal(t, "/path", requests[1].Path)
	assert.Equal(t, "/path", requests[2].Path)
	assert.Equal(t, "/event/path", requests[3].Path)
	assert.Empty(t, requests[0].Event)
	assert.Empty(t, requests[1].Event)
	assert.Empty(t, requests[2].Event)
	assert.Equal(t, "event", requests[3].Event)
	assert.True(t, requests[0].Bot)
	assert.True(t, requests[1].Bot)
	assert.True(t, requests[2].Bot)
	assert.True(t, requests[3].Bot)
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
	assert.InDelta(t, 10, sessionDuration[0].AverageTimeSpentSeconds, 2)
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

func TestTrackerIgnorePrefetch(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("X-Moz", "prefetch")
	_, _, ignore := tracker.ignore(req)
	assert.Equal(t, "prefetch", ignore)
	req.Header.Del("X-Moz")
	req.Header.Set("X-Purpose", "prefetch")
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "prefetch", ignore)
	req.Header.Set("X-Purpose", "preview")
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "prefetch", ignore)
	req.Header.Del("X-Purpose")
	req.Header.Set("Purpose", "prefetch")
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "prefetch", ignore)
	req.Header.Set("Purpose", "preview")
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "prefetch", ignore)
	req.Header.Del("Purpose")
	_, _, ignore = tracker.ignore(req)
	assert.Empty(t, ignore)
}

func TestTrackerIgnoreUserAgent(t *testing.T) {
	userAgents := []struct {
		userAgent string
		ignore    string
	}{
		{"This is a bot request", "ua-keyword"},
		{"This is a crawler request", "ua-keyword"},
		{"This is a spider request", "ua-keyword"},
		{"Visit http://spam.com!", "ua-keyword"},
		{"", "ua-chars"},
		{"172.22.0.11:30004", "ua-ip"},
		{"172.22.0.11", "ua-chars"},
		{"2345:0425:2CA1:0000:0000:0567:5673:23b5", "ua-ip"},
		{"2345:425:2CA1:0000:0000:567:5673:23b5", "ua-ip"},
		{"2345:0425:2CA1:0:0:0567:5673:23b5", "ua-ip"},
		{"[2345:0425:2CA1:0:0:0567:5673:23b5]:8080", "ua-ip"},
		{userAgent, ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/21B101 Instagram 312.0.1.19.124 (iPhone14,2; iOS 17_1_2; de_FR; de; scale=3.00; 1170x2532; 548339486)", ""},
		{"Mozilla/5.0 (Linux; Android 9; SM-G950F Build/PPR1.180610.011; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/116.0.0.0 Mobile Safari/537.36 trill_2023102050 JsSdk/1.0 NetType/4G Channel/googleplay AppName/musical_ly app_version/31.2.5 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:31.0) Gecko/20130401 Firefox/31.0", "browser"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 musical_ly_32.7.0 JsSdk/2.0 NetType/WIFI Channel/App Store ByteLocale/de Region/DE isDarkMode/0 WKWebView/1 RevealType/Dialog BytedanceWebview/d8a21c6 FalconTag/523CAEFB-209D-4BCF-A7A7-FEE8BD659140", ""},
		{"Mozilla/5.0 (Linux; Android 12; SM-G973F Build/SP1A.210812.016; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 11; SM-T500 Build/RP1A.200720.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.193 Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S918B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; 22081212UG Build/TKQ1.220829.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 12; M2102J20SG Build/SKQ1.211006.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/ru-RU ByteFullLocale/ru-RU Region/RU AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; RMX3511 Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S901B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-A725F Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 trill_2021707000 JsSdk/1.0 NetType/WIFI Channel/tt_eu_samsung2020_yz1 AppName/musical_ly app_version/17.7.0 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S916B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/MOBILE Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-A225F Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 Instagram 312.1.0.34.111 Android (33/13; 300dpi; 720x1452; samsung; SM-A225F; a22; mt6769t; de_DE; 548323754)", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-A546B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-A336B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 musical_ly_32.7.0 JsSdk/2.0 NetType/WIFI Channel/App Store ByteLocale/de Region/DE isDarkMode/0 WKWebView/1 RevealType/Dialog BytedanceWebview/d8a21c6 FalconTag/D99C3025-1798-4643-9FD5-00CFABA0DA30", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-A546B Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-G780G Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 trill_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0", ""},
		{"Mozilla/5.0 (Linux; Android 13; 23053RN02Y Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/111.0.5563.116 Mobile Safari/537.36 Instagram 312.1.0.34.111 Android (33/13; 440dpi; 1080x2226; Xiaomi/Redmi; 23053RN02Y; heat; mt6768; es_US; 548323755)", ""},
		{"Mozilla/5.0 (Linux; Android 10; M2003J15SC Build/QP1A.190711.020; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/es ByteFullLocale/es Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-A145R Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 trill_2023000000 JsSdk/1.0 NetType/WIFI Channel/samsung_preload AppName/musical_ly app_version/30.0.0 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 10; PPA-LX2 Build/HUAWEIPPA-LX2; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/92.0.4515.105 Mobile Safari/537.36 musical_ly_2023206050 JsSdk/1.0 NetType/WIFI Channel/huaweiadsglobal_int AppName/musical_ly app_version/32.6.5 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.7.4-bugfix AppVersion/32.6.5 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S911B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 9; ZTE Blade A7 2020 Build/PPR1.180610.011; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 musical_ly_32.7.0 JsSdk/2.0 NetType/4G Channel/App Store ByteLocale/de Region/DE isDarkMode/0 WKWebView/1 RevealType/Dialog BytedanceWebview/d8a21c6 FalconTag/6C26B20B-D898-4AA5-9455-688897104628", ""},
		{"Mozilla/5.0 (Linux; Android 13; 21051182G Build/TKQ1.221013.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; Pixel 6a Build/UP1A.231128.003; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/21E236 [FBAN/FBIOS;FBAV/465.0.1.41.103;FBBV/602060281;FBDV/iPhone13,4;FBMD/iPhone;FBSN/iOS;FBSV/17.4.1;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/603032588]", ""},
		{"-8235OR 5208=5208", "ua-regex"},
		{"-4368OR 2918=6019 AND ('Veeg'='Veeg", "ua-regex"},
		{"-2985OR 6255=1124 AND ('lNxX' LIKE 'lNxX", "ua-regex"},
		{"(CASE WHEN 3116=9361 THEN 3116 ELSE NULL END)", "ua-regex"},
		{"IiF(3856=6771,3856,1/0)", "ua-regex"},
	}

	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for _, userAgent := range userAgents {
		req.Header.Set("User-Agent", userAgent.userAgent)
		_, _, ignore := tracker.ignore(req)
		assert.Equal(t, userAgent.ignore, ignore, userAgent.userAgent)
	}
}

func TestTrackerIgnoreBotUserAgent(t *testing.T) {
	tracker := NewTracker(Config{})

	for _, botUserAgent := range ua.Blacklist {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", botUserAgent)
		_, _, ignore := tracker.ignore(req)
		assert.NotEmpty(t, ignore, botUserAgent)
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
		_, _, ignore := tracker.ignore(req)
		assert.NotEmpty(t, ignore, botUserAgent)
	}
}

func TestTrackerIgnoreReferrer(t *testing.T) {
	hostname := "2your.site"
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", hostname)
	_, _, ignore := tracker.ignore(req)
	assert.Equal(t, "referrer", ignore)
	req.Header.Set("Referer", fmt.Sprintf("subdomain.%s", hostname))
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "referrer", ignore)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?ref=%s", hostname), nil)
	req.Header.Set("User-Agent", userAgent)
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "referrer", ignore)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?ref=%s", hostname), nil)
	req.Header.Set("User-Agent", userAgent)
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "referrer", ignore)
}

func TestTrackerIgnoreBrowserVersion(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.4147.135 Safari/537.36")
	_, _, ignore := tracker.ignore(req)
	assert.Equal(t, "browser", ignore)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", userAgent)
	_, _, ignore = tracker.ignore(req)
	assert.Empty(t, ignore)
}

func TestTrackerIgnoreIP(t *testing.T) {
	filter := ip.NewUdger("", "", "")
	filter.Update([]string{"90.154.29.38"}, []string{}, []string{}, []string{}, []ip.Range{}, []ip.Range{})
	tracker := NewTracker(Config{
		IPFilter: filter,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", userAgent)
	_, _, ignore := tracker.ignore(req)
	assert.Empty(t, ignore)
	req.RemoteAddr = "90.154.29.38"
	_, _, ignore = tracker.ignore(req)
	assert.Equal(t, "ip", ignore)
}

func TestTrackerPageViewsLimit(t *testing.T) {
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

func TestTrackerPageViewsLimitOverride(t *testing.T) {
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
		tracker.PageView(req, 0, Options{
			MaxPageViews: 8,
		})
		time.Sleep(time.Millisecond * 5)
	}

	tracker.Stop()
	assert.Len(t, client.GetPageViews(), 8)
}

func TestTrackerGetLanguage(t *testing.T) {
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

func TestTrackerGetScreenClass(t *testing.T) {
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

func TestTrackerReferrerOrCampaignChanged(t *testing.T) {
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
