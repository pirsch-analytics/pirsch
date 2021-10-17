package pirsch

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestTrackerConfigValidate(t *testing.T) {
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

func TestTrackerHitTimeout(t *testing.T) {
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
	assert.Len(t, client.Hits, 2)
	assert.Len(t, client.UserAgents, 2)

	// ignore order...
	if client.Hits[0].Path != "/" && client.Hits[0].Path != "/hello-world" ||
		client.Hits[1].Path != "/" && client.Hits[1].Path != "/hello-world" {
		t.Fatalf("Hits not as expected: %v %v", client.Hits[0], client.Hits[1])
	}

	if client.UserAgents[0].UserAgent != uaString1 && client.UserAgents[0].UserAgent != uaString2 ||
		client.UserAgents[1].UserAgent != uaString1 && client.UserAgents[1].UserAgent != uaString2 {
		t.Fatalf("UserAgents not as expected: %v %v", client.UserAgents[0], client.UserAgents[1])
	}

	tracker.ClearSessionCache()
	assert.Len(t, sessionCache.sessions, 0)
}

func TestTrackerHitLimit(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 10,
	})

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		tracker.Hit(req, nil)
	}

	tracker.Stop()
	assert.Len(t, client.Hits, 7)
	assert.Len(t, client.UserAgents, 1)
}

func TestTrackerHitDiscard(t *testing.T) {
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

	assert.Len(t, client.Hits, 5)
	assert.Len(t, client.UserAgents, 1)
}

func TestTrackerHitCountryCode(t *testing.T) {
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
	assert.Len(t, client.Hits, 2)
	assert.Len(t, client.UserAgents, 2)
	foundGB := false
	foundEmpty := false

	for _, hit := range client.Hits {
		if hit.CountryCode == "gb" {
			foundGB = true
		} else if hit.CountryCode == "" {
			foundEmpty = true
		}
	}

	assert.True(t, foundGB)
	assert.True(t, foundEmpty)
}

func TestTrackerHitSession(t *testing.T) {
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
	assert.Len(t, client.Hits, 2)
	assert.Len(t, client.UserAgents, 1)
	assert.Equal(t, client.Hits[0].SessionID, client.Hits[1].SessionID)
}

func TestTrackerHitIgnoreSubdomain(t *testing.T) {
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
	assert.Len(t, client.Hits, 4)
	assert.Len(t, client.UserAgents, 1)

	for _, hit := range client.Hits {
		assert.Empty(t, hit.Referrer)
	}
}

func TestTrackerEvent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	client := NewMockClient()
	tracker := NewTracker(client, "salt", nil)
	tracker.Event(req, EventOptions{Name: "  "}, nil)                                                                                // ignore (invalid name)
	tracker.Event(req, EventOptions{Name: ""}, nil)                                                                                  // ignore (invalid name)
	tracker.Event(req, EventOptions{Name: " event  ", Duration: 42, Meta: map[string]string{"hello": "world", "meta": "data"}}, nil) // store duration and meta data
	tracker.Stop()
	assert.Len(t, client.Events, 1)
	assert.Len(t, client.UserAgents, 1)
	assert.Equal(t, "event", client.Events[0].Name)
	assert.Equal(t, uint32(42), client.Events[0].DurationSeconds)
	assert.Len(t, client.Events[0].MetaKeys, 2)
	assert.Len(t, client.Events[0].MetaValues, 2)
	assert.Contains(t, client.Events[0].MetaKeys, "hello")
	assert.Contains(t, client.Events[0].MetaKeys, "meta")
	assert.Contains(t, client.Events[0].MetaValues, "world")
	assert.Contains(t, client.Events[0].MetaValues, "data")
}

func TestTrackerEventTimeout(t *testing.T) {
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
	assert.Len(t, client.UserAgents, 1)

	// ignore order...
	if client.Events[0].Path != "/" && client.Events[0].Path != "/hello-world" ||
		client.Events[1].Path != "/" && client.Events[1].Path != "/hello-world" {
		t.Fatalf("Hits not as expected: %v %v", client.Events[0], client.Events[1])
	}
}

func TestTrackerEventLimit(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 10,
	})

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		tracker.Event(req, EventOptions{Name: "event"}, nil)
	}

	tracker.Stop()
	assert.Len(t, client.Events, 7)
	assert.Len(t, client.UserAgents, 1)
}

func TestTrackerEventDiscard(t *testing.T) {
	client := NewMockClient()
	tracker := NewTracker(client, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 5,
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
		tracker.Event(req, EventOptions{Name: "event"}, nil)

		if i > 3 {
			tracker.Stop()
		}
	}

	assert.Len(t, client.Events, 5)
	assert.Len(t, client.UserAgents, 1)
}

func TestTrackerEventCountryCode(t *testing.T) {
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
	assert.Len(t, client.UserAgents, 2)
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

func TestTrackerEventSession(t *testing.T) {
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
	assert.Len(t, client.UserAgents, 1)
	assert.Equal(t, client.Events[0].SessionID, client.Events[1].SessionID)
}

func TestTrackerExtendSession(t *testing.T) {
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
	assert.Len(t, client.Hits, 1)
	at := client.Hits[0].Time
	fingerprint := client.Hits[0].VisitorID
	time.Sleep(time.Millisecond * 20)
	tracker.ExtendSession(req, 0)
	tracker.Flush()
	assert.Len(t, client.Hits, 1)
	hit := sessionCache.Get(0, fingerprint, time.Now().UTC().Add(-time.Second))
	assert.NotEqual(t, at, hit.Time)
	assert.True(t, hit.Time.After(at))
}

func TestTrackerEventIgnoreSubdomain(t *testing.T) {
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
	assert.Len(t, client.UserAgents, 1)

	for _, hit := range client.Events {
		assert.Empty(t, hit.Referrer)
	}
}
