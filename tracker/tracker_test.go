package tracker

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ua"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

const (
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:105.0) Gecko/20100101 Firefox/105.0"
)

func TestTracker_PageView(t *testing.T) {
	now := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/foo/bar?utm_source=Source&utm_campaign=Campaign&utm_medium=Medium&utm_content=Content&utm_term=Term", nil)
	req.Header.Add("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	req.Header.Set("Referer", "https://google.com")
	req.RemoteAddr = "81.2.69.142"
	geoDB, err := geodb.NewGeoDB(geodb.Config{
		File: filepath.Join("geodb/GeoIP2-City-Test.mmdb"),
	})
	assert.NoError(t, err)
	client := db.NewMockClient()
	tracker := NewTracker(Config{
		Store: client,
		GeoDB: geoDB,
	})
	tracker.PageView(req, 123, Options{
		Title:        "Foo",
		ScreenWidth:  1920,
		ScreenHeight: 1080,
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
	assert.Equal(t, pirsch.OSLinux, sessions[0].OS)
	assert.Empty(t, sessions[0].OSVersion)
	assert.Equal(t, pirsch.BrowserFirefox, sessions[0].Browser)
	assert.Equal(t, "105.0", sessions[0].BrowserVersion)
	assert.True(t, sessions[0].Desktop)
	assert.False(t, sessions[0].Mobile)
	assert.Equal(t, uint16(1920), sessions[0].ScreenWidth)
	assert.Equal(t, uint16(1080), sessions[0].ScreenHeight)
	assert.Equal(t, "Full HD", sessions[0].ScreenClass)
	assert.Equal(t, "Source", sessions[0].UTMSource)
	assert.Equal(t, "Medium", sessions[0].UTMMedium)
	assert.Equal(t, "Campaign", sessions[0].UTMCampaign)
	assert.Equal(t, "Content", sessions[0].UTMContent)
	assert.Equal(t, "Term", sessions[0].UTMTerm)
	assert.Equal(t, uint8(0), sessions[0].IsBot)

	assert.Equal(t, uint64(123), pageViews[0].ClientID)
	assert.True(t, pageViews[0].Time.After(now))
	assert.Equal(t, uint32(0), pageViews[0].DurationSeconds)
	assert.Equal(t, "/foo/bar", pageViews[0].Path)
	assert.Equal(t, "fr", pageViews[0].Language)
	assert.Equal(t, "gb", pageViews[0].CountryCode)
	assert.Equal(t, "London", pageViews[0].City)
	assert.Equal(t, "https://google.com", pageViews[0].Referrer)
	assert.Equal(t, "Google", pageViews[0].ReferrerName)
	assert.Equal(t, pirsch.OSLinux, pageViews[0].OS)
	assert.Empty(t, pageViews[0].OSVersion)
	assert.Equal(t, pirsch.BrowserFirefox, pageViews[0].Browser)
	assert.Equal(t, "105.0", pageViews[0].BrowserVersion)
	assert.True(t, pageViews[0].Desktop)
	assert.False(t, pageViews[0].Mobile)
	assert.Equal(t, uint16(1920), pageViews[0].ScreenWidth)
	assert.Equal(t, uint16(1080), pageViews[0].ScreenHeight)
	assert.Equal(t, "Full HD", pageViews[0].ScreenClass)
	assert.Equal(t, "Source", pageViews[0].UTMSource)
	assert.Equal(t, "Medium", pageViews[0].UTMMedium)
	assert.Equal(t, "Campaign", pageViews[0].UTMCampaign)
	assert.Equal(t, "Content", pageViews[0].UTMContent)
	assert.Equal(t, "Term", pageViews[0].UTMTerm)

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
	assert.Equal(t, pirsch.OSLinux, sessions[2].OS)
	assert.Empty(t, sessions[2].OSVersion)
	assert.Equal(t, pirsch.BrowserFirefox, sessions[2].Browser)
	assert.Equal(t, "105.0", sessions[2].BrowserVersion)
	assert.True(t, sessions[2].Desktop)
	assert.False(t, sessions[2].Mobile)
	assert.Equal(t, uint16(1920), sessions[2].ScreenWidth)
	assert.Equal(t, uint16(1080), sessions[2].ScreenHeight)
	assert.Equal(t, "Full HD", sessions[2].ScreenClass)
	assert.Equal(t, "Source", sessions[2].UTMSource)
	assert.Equal(t, "Medium", sessions[2].UTMMedium)
	assert.Equal(t, "Campaign", sessions[2].UTMCampaign)
	assert.Equal(t, "Content", sessions[2].UTMContent)
	assert.Equal(t, "Term", sessions[2].UTMTerm)
	assert.Equal(t, uint8(0), sessions[2].IsBot)

	assert.Equal(t, uint64(123), pageViews[1].ClientID)
	assert.True(t, pageViews[1].Time.After(now))
	assert.Equal(t, uint32(1), pageViews[1].DurationSeconds)
	assert.Equal(t, "/test", pageViews[1].Path)
	assert.Equal(t, "fr", pageViews[1].Language)
	assert.Equal(t, "gb", pageViews[1].CountryCode)
	assert.Equal(t, "London", pageViews[1].City)
	assert.Equal(t, "https://google.com", pageViews[1].Referrer)
	assert.Equal(t, "Google", pageViews[1].ReferrerName)
	assert.Equal(t, pirsch.OSLinux, pageViews[1].OS)
	assert.Empty(t, pageViews[1].OSVersion)
	assert.Equal(t, pirsch.BrowserFirefox, pageViews[1].Browser)
	assert.Equal(t, "105.0", pageViews[1].BrowserVersion)
	assert.True(t, pageViews[1].Desktop)
	assert.False(t, pageViews[1].Mobile)
	assert.Equal(t, uint16(1920), pageViews[1].ScreenWidth)
	assert.Equal(t, uint16(1080), pageViews[1].ScreenHeight)
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

/*func TestTracker_HitIgnoreSubdomain(t *testing.T) {
	client := db.NewMockClient()
	tracker := NewTracker(client, "salt", &Config{
		WorkerTimeout: time.Second,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req.RemoteAddr = "81.2.69.142"
	tracker.Hit(req, &HitOptions{
		ReferrerDomainBlacklist: []string{"pirsch.io"},
		Referrer:                "https://pirsch.io/",
	})
	tracker.Hit(req, &HitOptions{
		ReferrerDomainBlacklist:                   []string{"pirsch.io"},
		ReferrerDomainBlacklistIncludesSubdomains: true,
		Referrer: "https://www.pirsch.io/",
	})
	tracker.Hit(req, &HitOptions{
		ReferrerDomainBlacklist: []string{"pirsch.io", "www.pirsch.io"},
		Referrer:                "https://www.pirsch.io/",
	})
	tracker.Hit(req, &HitOptions{
		ReferrerDomainBlacklist: []string{"pirsch.io"},
		Referrer:                "pirsch.io",
	})
	tracker.Stop()
	sessions := client.GetSessions()
	assert.Len(t, client.GetPageViews(), 4)
	assert.Len(t, sessions, 7)
	assert.Len(t, client.GetUserAgents(), 1)

	for _, hit := range sessions {
		assert.Empty(t, hit.Referrer)
	}
}*/

/*func TestTracker_HitIsBot(t *testing.T) {
	db.CleanupDB(t, dbClient)
	cache := session2.NewRedisCache(time.Second*60, nil, &redis.Options{
		Addr: "localhost:6379",
	})
	cache.Clear()
	tracker := NewTracker(dbClient, "salt", &Config{
		Worker:           4,
		WorkerBufferSize: 5,
		WorkerTimeout:    time.Second * 2,
		SessionCache:     cache,
	})

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		req.URL.Path = fmt.Sprintf("/page/%d", i)
		go tracker.Hit(req, nil)
		time.Sleep(time.Millisecond * 5)
	}

	tracker.Stop()
	var session model.Session
	assert.NoError(t, dbClient.QueryRow(`SELECT entry_path, exit_path, max(page_views) page_views, max(is_bot) is_bot
		FROM session
		GROUP BY entry_path, exit_path
		HAVING sum(sign) > 0`).Scan(&session.EntryPath, &session.ExitPath, &session.PageViews, &session.IsBot))
	assert.Equal(t, uint8(5), session.IsBot)
	assert.Equal(t, 6, int(session.PageViews))
	assert.Equal(t, "/page/0", session.EntryPath)
	assert.Equal(t, "/page/5", session.ExitPath)
}*/

func TestTracker_Event(t *testing.T) {
	// TODO
}

func TestTracker_ExtendSession(t *testing.T) {
	// TODO
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
		{userAgent, false},
	}

	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for _, userAgent := range userAgents {
		req.Header.Set("User-Agent", userAgent.userAgent)

		if _, _, ignore := tracker.ignore(req); ignore != userAgent.ignore {
			t.Fatalf("Request with User-Agent '%s' must be ignored", userAgent.userAgent)
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
	tracker := NewTracker(Config{})
	assert.Equal(t, "", tracker.getScreenClass(0))
	assert.Equal(t, "XS", tracker.getScreenClass(42))
	assert.Equal(t, "XL", tracker.getScreenClass(1024))
	assert.Equal(t, "XL", tracker.getScreenClass(1025))
	assert.Equal(t, "HD", tracker.getScreenClass(1919))
	assert.Equal(t, "Full HD", tracker.getScreenClass(2559))
	assert.Equal(t, "WQHD", tracker.getScreenClass(3839))
	assert.Equal(t, "UHD 4K", tracker.getScreenClass(5119))
	assert.Equal(t, "UHD 5K", tracker.getScreenClass(5120))
}

func TestTracker_referrerOrCampaignChanged(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Referer", "https://referrer.com")
	session := &model.Session{Referrer: "https://referrer.com"}
	assert.False(t, tracker.referrerOrCampaignChanged(req, session, ""))
	session.Referrer = ""
	assert.True(t, tracker.referrerOrCampaignChanged(req, session, ""))
	session.Referrer = "https://referrer.com"
	req = httptest.NewRequest(http.MethodGet, "/test?ref=https://different.com", nil)
	assert.True(t, tracker.referrerOrCampaignChanged(req, session, ""))
	req = httptest.NewRequest(http.MethodGet, "/test?utm_source=Referrer", nil)
	assert.True(t, tracker.referrerOrCampaignChanged(req, session, ""))
	session.UTMSource = "Referrer"
	assert.False(t, tracker.referrerOrCampaignChanged(req, session, ""))
}
