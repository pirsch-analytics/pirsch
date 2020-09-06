package pirsch

import (
	"net/http"
	"net/http/httptest"
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
	tracker := NewTracker(store, "salt", &TrackerConfig{WorkerTimeout: time.Second * 2})
	tracker.Hit(req1, nil)
	tracker.Hit(req2, nil)
	time.Sleep(time.Second * 4)

	if len(store.hits) != 2 {
		t.Fatalf("Two requests must have been tracked, but was: %v", len(store.hits))
	}

	// ignore order...
	if store.hits[0].Path.String != "/" && store.hits[0].Path.String != "/hello-world" ||
		store.hits[1].Path.String != "/" && store.hits[1].Path.String != "/hello-world" {
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

	time.Sleep(time.Second) // allow all hits to be tracked
	tracker.Stop()

	if len(store.hits) != 7 {
		t.Fatalf("All requests must have been tracked, but was: %v", len(store.hits))
	}
}
