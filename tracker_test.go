package pirsch

import (
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

	if cfg.Worker != runtime.NumCPU() ||
		cfg.WorkerBufferSize != defaultWorkerBufferSize ||
		cfg.WorkerTimeout != defaultWorkerTimeout ||
		len(cfg.ReferrerDomainBlacklist) != 0 ||
		cfg.ReferrerDomainBlacklistIncludesSubdomains {
		t.Fatal("TrackerConfig must have default values")
	}

	cfg = &TrackerConfig{
		Worker:                  123,
		WorkerBufferSize:        42,
		WorkerTimeout:           time.Second * 57,
		ReferrerDomainBlacklist: []string{"localhost"},
		ReferrerDomainBlacklistIncludesSubdomains: true,
	}
	cfg.validate()

	if cfg.Worker != 123 ||
		cfg.WorkerBufferSize != 42 ||
		cfg.WorkerTimeout != time.Second*57 ||
		len(cfg.ReferrerDomainBlacklist) != 1 ||
		!cfg.ReferrerDomainBlacklistIncludesSubdomains {
		t.Fatal("TrackerConfig must have set values")
	}

	cfg = &TrackerConfig{WorkerTimeout: time.Second * 142}
	cfg.validate()

	if cfg.WorkerTimeout != maxWorkerTimeout {
		t.Fatalf("WorkerTimout must have been limited, but was: %v", cfg.WorkerTimeout)
	}
}

func TestTrackerHitTimeout(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "valid")
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "valid")
	store := newTestStore()
	tracker := NewTracker(store, "salt", &TrackerConfig{WorkerTimeout: time.Millisecond * 200})
	tracker.Hit(req1, nil)
	tracker.Hit(req2, nil)
	time.Sleep(time.Millisecond * 210)

	if len(store.hits) != 2 {
		t.Fatalf("Two requests must have been tracked, but was: %v", len(store.hits))
	}

	// ignore order...
	if store.hits[0].Path != "/" && store.hits[0].Path != "/hello-world" ||
		store.hits[1].Path != "/" && store.hits[1].Path != "/hello-world" {
		t.Fatalf("Hits not as expected: %v %v", store.hits[0], store.hits[1])
	}
}

func TestTrackerHitLimit(t *testing.T) {
	store := newTestStore()
	tracker := NewTracker(store, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 10,
	})

	for i := 0; i < 7; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "valid")
		tracker.Hit(req, nil)
	}

	tracker.Stop()

	if len(store.hits) != 7 {
		t.Fatalf("All requests must have been tracked, but was: %v", len(store.hits))
	}
}

func TestTrackerHitDiscard(t *testing.T) {
	store := newTestStore()
	tracker := NewTracker(store, "salt", &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 5,
	})

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("User-Agent", "valid")
		tracker.Hit(req, nil)

		if i > 3 {
			tracker.Stop()
		}
	}

	if len(store.hits) != 5 {
		t.Fatalf("All requests must have been tracked, but was: %v", len(store.hits))
	}
}

func TestTrackerCountryCode(t *testing.T) {
	geoDB, err := NewGeoDB(GeoDBConfig{
		File: filepath.Join("geodb/GeoIP2-Country-Test.mmdb"),
	})

	if err != nil {
		t.Fatalf("Geo DB must have been loaded, but was: %v", err)
	}

	defer geoDB.Close()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "valid")
	req1.RemoteAddr = "81.2.69.142"
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "valid")
	req2.RemoteAddr = "127.0.0.1"
	store := newTestStore()
	tracker := NewTracker(store, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	tracker.SetGeoDB(geoDB)
	tracker.Hit(req1, nil)
	tracker.Hit(req2, nil)
	tracker.Stop()

	if len(store.hits) != 2 {
		t.Fatalf("Two requests must have been tracked, but was: %v", len(store.hits))
	}

	foundGB := false
	foundEmpty := false

	for _, hit := range store.hits {
		if hit.CountryCode.String == "gb" {
			foundGB = true
		} else if hit.CountryCode.String == "" {
			foundEmpty = true
		}
	}

	if !foundGB || !foundEmpty {
		t.Fatalf("Hits not as expected: %v", store.hits)
	}
}

func TestTrackerHitSession(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Add("User-Agent", "valid")
	req2 := httptest.NewRequest(http.MethodGet, "/hello-world", nil)
	req2.Header.Add("User-Agent", "valid")
	store := newTestStore()
	tracker := NewTracker(store, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
		Sessions:      true,
	})
	tracker.Hit(req1, nil)
	tracker.Hit(req2, nil)
	tracker.Stop()

	if len(store.hits) != 2 {
		t.Fatalf("Two requests must have been tracked, but was: %v", len(store.hits))
	}

	// ignore order...
	if !store.hits[0].Session.Valid || !store.hits[1].Session.Valid ||
		store.hits[0].Session.Time.IsZero() || store.hits[1].Session.Time.IsZero() {
		t.Fatalf("Hits not as expected: %v %v", store.hits[0], store.hits[1])
	}
}

func TestTrackerIgnoreSubdomain(t *testing.T) {
	store := newTestStore()
	tracker := NewTracker(store, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "valid")
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
	tracker.Stop()

	if len(store.hits) != 3 {
		t.Fatalf("Three hits must have been tracked, but was: %v", len(store.hits))
	}

	for _, hit := range store.hits {
		if hit.Referrer.Valid {
			t.Fatalf("No referrers must have been kept, but was: %v", hit.Referrer.String)
		}
	}
}

func BenchmarkTracker(b *testing.B) {
	geoDB, err := NewGeoDB(GeoDBConfig{
		File: filepath.Join("geodb/GeoIP2-Country-Test.mmdb"),
	})

	if err != nil {
		b.Fatalf("Geo DB must have been loaded, but was: %v", err)
	}

	defer geoDB.Close()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "valid")
	req.RemoteAddr = "81.2.69.142"
	store := NewPostgresStore(postgresDB, nil)
	tracker := NewTracker(store, "salt", &TrackerConfig{
		WorkerTimeout: time.Second,
		Sessions:      true,
		GeoDB:         geoDB,
	})

	for i := 0; i < 10000; i++ {
		tracker.Hit(req, nil)
	}

	tracker.Stop()
}
