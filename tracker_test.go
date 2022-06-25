package pirsch

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestTrackerConfig_Validate(t *testing.T) {
	cfg := &TrackerConfig{}
	cfg.validate()
	assert.Equal(t, runtime.NumCPU(), cfg.Worker)
	assert.Equal(t, defaultWorkerBufferSize, cfg.WorkerBufferSize)
	assert.Equal(t, defaultWorkerTimeout, cfg.WorkerTimeout)
	assert.Len(t, cfg.ReferrerDomainBlacklist, 0)
	assert.False(t, cfg.ReferrerDomainBlacklistIncludesSubdomains)
	cfg = &TrackerConfig{
		Worker:                  123,
		WorkerBufferSize:        42,
		WorkerTimeout:           time.Second * 57,
		ReferrerDomainBlacklist: []string{"localhost"},
		ReferrerDomainBlacklistIncludesSubdomains: true,
	}
	cfg.validate()
	assert.Equal(t, 123, cfg.Worker)
	assert.Equal(t, 42, cfg.WorkerBufferSize)
	assert.Equal(t, time.Second*57, cfg.WorkerTimeout)
	assert.Len(t, cfg.ReferrerDomainBlacklist, 1)
	assert.True(t, cfg.ReferrerDomainBlacklistIncludesSubdomains)
	cfg = &TrackerConfig{WorkerTimeout: time.Second * 142}
	cfg.validate()
	assert.Equal(t, maxWorkerTimeout, cfg.WorkerTimeout)
}

func TestNewTracker(t *testing.T) {
	tracker := NewTracker(nil, "", nil)
	assert.Len(t, tracker.headerParser, len(DefaultHeaderParser))
	tracker = NewTracker(nil, "", &TrackerConfig{
		HeaderParser: []HeaderParser{},
	})
	assert.Len(t, tracker.headerParser, 0)
}

func TestTracker_HitTimeout(t *testing.T) {
	uaString1 := "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0"
	uaString2 := "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/88.0"
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", uaString1)
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", uaString2)
	client := NewMockClient()
	sessionCache := NewSessionCacheMem(client, 100)
	tracker := NewTracker(client, "salt", &TrackerConfig{WorkerTimeout: time.Millisecond * 200, SessionCache: sessionCache})
	tracker.Hit(req1, nil)
	tracker.Hit(req2, nil)
	time.Sleep(time.Millisecond * 210)
	assert.Len(t, client.Sessions, 2)
	assert.Len(t, client.UserAgents, 2)

	// ignore order...
	if client.Sessions[0].ExitPath != "/" && client.Sessions[0].ExitPath != "/hello-world" ||
		client.Sessions[1].ExitPath != "/" && client.Sessions[1].ExitPath != "/hello-world" {
		t.Fatalf("Sessions not as expected: %v %v", client.Sessions[0], client.Sessions[1])
	}

	if client.UserAgents[0].UserAgent != uaString1 && client.UserAgents[0].UserAgent != uaString2 ||
		client.UserAgents[1].UserAgent != uaString1 && client.UserAgents[1].UserAgent != uaString2 {
		t.Fatalf("UserAgents not as expected: %v %v", client.UserAgents[0], client.UserAgents[1])
	}

	tracker.ClearSessionCache()
	assert.Len(t, sessionCache.sessions, 0)
}

func TestTracker_HitLimit(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker:              1,
		WorkerBufferSize:    10,
		DisableFlaggingBots: true,
	})

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		tracker.Hit(req, nil)
	}

	tracker.Stop()
	assert.Len(t, client.PageViews, 7)
	assert.Len(t, client.Sessions, 13)
	assert.Len(t, client.UserAgents, 1)
}

func TestTracker_HitDiscard(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 5,
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		tracker.Hit(req, nil)

		if i > 3 {
			tracker.Stop()
		}
	}

	assert.Len(t, client.PageViews, 5)
	assert.Len(t, client.Sessions, 9)
	assert.Len(t, client.UserAgents, 1)
}

func TestTracker_HitCountryCode(t *testing.T) {
	geoDB, err := NewGeoDB(GeoDBConfig{
		File: filepath.Join("geodb/GeoIP2-City-Test.mmdb"),
	})
	assert.NoError(t, err)
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req1.RemoteAddr = "81.2.69.142"
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req2.RemoteAddr = "127.0.0.1"
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	tracker.SetGeoDB(geoDB)
	tracker.Hit(req1, nil)
	tracker.Hit(req2, nil)
	tracker.Stop()
	assert.Len(t, client.PageViews, 2)
	assert.Len(t, client.Sessions, 2)
	assert.Len(t, client.UserAgents, 2)
	foundGB := false
	foundEmpty := false

	for _, hit := range client.Sessions {
		if hit.CountryCode == "gb" {
			foundGB = true
		} else if hit.CountryCode == "" {
			foundEmpty = true
		}
	}

	assert.True(t, foundGB)
	assert.True(t, foundEmpty)
}

func TestTracker_HitSession(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	tracker.Hit(req1, nil)
	tracker.Hit(req2, nil)
	tracker.Stop()
	assert.Len(t, client.PageViews, 2)
	assert.Len(t, client.Sessions, 3)
	assert.Len(t, client.UserAgents, 1)
	assert.Equal(t, client.Sessions[0].SessionID, client.Sessions[1].SessionID)
}

func TestTracker_HitTitle(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	tracker.Hit(req, &HitOptions{
		Title: "title",
	})
	tracker.Stop()
	assert.Len(t, client.PageViews, 1)
	assert.Len(t, client.Sessions, 1)
	assert.Len(t, client.UserAgents, 1)
	assert.Equal(t, "title", client.Sessions[0].EntryTitle)
	assert.Equal(t, "title", client.Sessions[0].ExitTitle)
	assert.Equal(t, "title", client.PageViews[0].Title)
}

func TestTracker_HitIgnoreSubdomain(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
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
	assert.Len(t, client.PageViews, 4)
	assert.Len(t, client.Sessions, 7)
	assert.Len(t, client.UserAgents, 1)

	for _, hit := range client.Sessions {
		assert.Empty(t, hit.Referrer)
	}
}

func TestTracker_HitConcurrency(t *testing.T) {
	cleanupDB()
	cache := NewSessionCacheRedis(time.Second*60, nil, &redis.Options{
		Addr: "localhost:6379",
	})
	cache.Clear()
	config := &TrackerConfig{
		Worker:              4,
		WorkerBufferSize:    5,
		WorkerTimeout:       time.Second * 2,
		SessionCache:        cache,
		DisableFlaggingBots: true,
	}
	tracker := make([]*Tracker, 10)

	for i := 0; i < 10; i++ {
		tracker[i] = NewTracker(dbClient, "salt", config)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")

	for i := 0; i < 100; i++ {
		req.URL.Path = fmt.Sprintf("/page/%d", i+1)
		go tracker[i%10].Hit(req, nil)
		time.Sleep(time.Millisecond)
	}

	time.Sleep(time.Second * 3)

	for i := 0; i < 10; i++ {
		tracker[i].Stop()
	}

	var session Session
	assert.NoError(t, dbClient.Get(&session, `SELECT entry_path, exit_path, max(page_views) page_views
		FROM session
		GROUP BY entry_path, exit_path
		HAVING sum(sign) > 0`))
	assert.Equal(t, 100, int(session.PageViews))
	assert.Equal(t, "/page/1", session.EntryPath)
	assert.Equal(t, "/page/100", session.ExitPath)
}

func TestTracker_HitIsBot(t *testing.T) {
	cleanupDB()
	cache := NewSessionCacheRedis(time.Second*60, nil, &redis.Options{
		Addr: "localhost:6379",
	})
	cache.Clear()
	tracker := NewTracker(dbClient, "salt", &TrackerConfig{
		Worker:           4,
		WorkerBufferSize: 5,
		WorkerTimeout:    time.Second * 2,
		SessionCache:     cache,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")

	for i := 0; i < 7; i++ {
		req.URL.Path = fmt.Sprintf("/page/%d", i)
		go tracker.Hit(req, nil)
		time.Sleep(time.Millisecond * 5)
	}

	tracker.Stop()
	var session Session
	assert.NoError(t, dbClient.Get(&session, `SELECT entry_path, exit_path, max(page_views) page_views, max(is_bot) is_bot
		FROM session
		GROUP BY entry_path, exit_path
		HAVING sum(sign) > 0`))
	assert.Equal(t, uint8(5), session.IsBot)
	assert.Equal(t, 6, int(session.PageViews))
	assert.Equal(t, "/page/0", session.EntryPath)
	assert.Equal(t, "/page/5", session.ExitPath)
}

func TestTracker_Event(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", nil)
	tracker.Event(req, EventOptions{Name: "  "}, nil)                                                                                // ignore (invalid name)
	tracker.Event(req, EventOptions{Name: ""}, nil)                                                                                  // ignore (invalid name)
	tracker.Event(req, EventOptions{Name: " event  ", Duration: 42, Meta: map[string]string{"hello": "world", "meta": "data"}}, nil) // store duration and meta data
	tracker.Stop()
	assert.Len(t, client.Events, 1)
	assert.Equal(t, "event", client.Events[0].Name)
	assert.Equal(t, uint32(42), client.Events[0].DurationSeconds)
	assert.Len(t, client.Events[0].MetaKeys, 2)
	assert.Len(t, client.Events[0].MetaValues, 2)
	assert.Contains(t, client.Events[0].MetaKeys, "hello")
	assert.Contains(t, client.Events[0].MetaKeys, "meta")
	assert.Contains(t, client.Events[0].MetaValues, "world")
	assert.Contains(t, client.Events[0].MetaValues, "data")
}

func TestTracker_EventTimeout(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{WorkerTimeout: time.Millisecond * 200})
	tracker.Event(req1, EventOptions{Name: "event"}, nil)
	tracker.Event(req2, EventOptions{Name: "event"}, nil)
	time.Sleep(time.Millisecond * 210)
	assert.Len(t, client.Events, 2)

	// ignore order...
	if client.Events[0].Path != "/" && client.Events[0].Path != "/hello-world" ||
		client.Events[1].Path != "/" && client.Events[1].Path != "/hello-world" {
		t.Fatalf("Sessions not as expected: %v %v", client.Events[0], client.Events[1])
	}
}

func TestTracker_EventLimit(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 10,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		tracker.Event(req, EventOptions{Name: "event"}, nil)
	}

	tracker.Stop()
	assert.Len(t, client.Events, 7)
}

func TestTracker_EventDiscard(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 5,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		tracker.Event(req, EventOptions{Name: "event"}, nil)

		if i > 3 {
			tracker.Stop()
		}
	}

	assert.Len(t, client.Events, 5)
}

func TestTracker_EventCountryCode(t *testing.T) {
	geoDB, err := NewGeoDB(GeoDBConfig{
		File: filepath.Join("geodb/GeoIP2-City-Test.mmdb"),
	})
	assert.NoError(t, err)
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req1.RemoteAddr = "81.2.69.142"
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req2.RemoteAddr = "127.0.0.1"
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	tracker.SetGeoDB(geoDB)
	tracker.Event(req1, EventOptions{Name: "event"}, nil)
	tracker.Event(req2, EventOptions{Name: "event"}, nil)
	tracker.Stop()
	assert.Len(t, client.Events, 2)
	foundGB := false
	foundEmpty := false

	for _, event := range client.Events {
		if event.CountryCode == "gb" {
			foundGB = true
		} else if event.CountryCode == "" {
			foundEmpty = true
		}
	}

	assert.True(t, foundGB)
	assert.True(t, foundEmpty)
}

func TestTracker_EventSession(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	tracker.Event(req1, EventOptions{Name: "event"}, nil)
	tracker.Event(req2, EventOptions{Name: "event"}, nil)
	tracker.Stop()
	assert.Len(t, client.Events, 2)
	assert.Equal(t, client.Events[0].SessionID, client.Events[1].SessionID)
}

func TestTracker_EventTitle(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	tracker.Event(req, EventOptions{Name: "event"}, &HitOptions{
		Title: "title",
	})
	tracker.Stop()
	assert.Len(t, client.PageViews, 0)
	assert.Len(t, client.Sessions, 1)
	assert.Len(t, client.UserAgents, 0)
	assert.Len(t, client.Events, 1)
	assert.Equal(t, "title", client.Events[0].Title)
}

func TestTracker_EventIgnoreSubdomain(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req.RemoteAddr = "81.2.69.142"
	tracker.Event(req, EventOptions{Name: "event"}, &HitOptions{
		ReferrerDomainBlacklist: []string{"pirsch.io"},
		Referrer:                "https://pirsch.io/",
	})
	tracker.Event(req, EventOptions{Name: "event"}, &HitOptions{
		ReferrerDomainBlacklist:                   []string{"pirsch.io"},
		ReferrerDomainBlacklistIncludesSubdomains: true,
		Referrer: "https://www.pirsch.io/",
	})
	tracker.Event(req, EventOptions{Name: "event"}, &HitOptions{
		ReferrerDomainBlacklist: []string{"pirsch.io", "www.pirsch.io"},
		Referrer:                "https://www.pirsch.io/",
	})
	tracker.Event(req, EventOptions{Name: "event"}, &HitOptions{
		ReferrerDomainBlacklist: []string{"pirsch.io"},
		Referrer:                "pirsch.io",
	})
	tracker.Stop()
	assert.Len(t, client.Events, 4)

	for _, hit := range client.Events {
		assert.Empty(t, hit.Referrer)
	}
}

func TestTracker_EventUpdateSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker: 1,
	})
	tracker.Hit(req, nil)
	time.Sleep(time.Millisecond * 10)
	tracker.Event(req, EventOptions{Name: "event"}, nil)
	tracker.Stop()
	assert.Len(t, client.PageViews, 1)
	assert.Len(t, client.Sessions, 3)
	assert.Len(t, client.UserAgents, 1)
	assert.Len(t, client.Events, 1)
	assert.Equal(t, int8(1), client.Sessions[0].Sign)
	assert.Equal(t, int8(-1), client.Sessions[1].Sign)
	assert.Equal(t, int8(1), client.Sessions[2].Sign)
	assert.True(t, client.Sessions[0].IsBounce)
	assert.True(t, client.Sessions[1].IsBounce)
	assert.False(t, client.Sessions[2].IsBounce)
	assert.Equal(t, uint16(1), client.Sessions[0].PageViews)
	assert.Equal(t, uint16(1), client.Sessions[1].PageViews)
	assert.Equal(t, uint16(1), client.Sessions[2].PageViews)
}

func TestTracker_ExtendSession(t *testing.T) {
	cleanupDB()
	uaString := "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"
	sessionCache := NewSessionCacheMem(dbClient, 100)
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.Header.Set("User-Agent", uaString)
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		SessionCache:  sessionCache,
		SessionMaxAge: time.Second * 10,
	})
	tracker.Hit(req, nil)
	tracker.Flush()
	assert.Len(t, client.Sessions, 1)
	at := client.Sessions[0].Time
	fingerprint := client.Sessions[0].VisitorID
	time.Sleep(time.Millisecond * 20)
	tracker.ExtendSession(req, nil)
	tracker.Flush()
	assert.Len(t, client.Sessions, 1)
	hit := sessionCache.Get(0, fingerprint, time.Now().UTC().Add(-time.Second))
	assert.NotEqual(t, at, hit.Time)
	assert.True(t, hit.Time.After(at))
}
