package pirsch

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testStore struct {
	hits []Hit
}

func newTestStore() *testStore {
	return &testStore{make([]Hit, 0)}
}

func (store *testStore) Save(hits []Hit) {
	log.Printf("Saved %d hits", len(hits))
	store.hits = append(store.hits, hits...)
}

func TestTrackerHitTimeout(t *testing.T) {
	store := newTestStore()
	tracker := NewTracker(store, &TrackerConfig{WorkerTimeout: time.Second * 2})
	tracker.Hit(httptest.NewRequest(http.MethodGet, "/", nil))
	tracker.Hit(httptest.NewRequest(http.MethodGet, "/hello-world", nil))
	time.Sleep(time.Second * 3)

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
	tracker := NewTracker(store, &TrackerConfig{
		Worker:           1,
		WorkerBufferSize: 10,
	})

	for i := 0; i < 7; i++ {
		tracker.Hit(httptest.NewRequest(http.MethodGet, "/", nil))
	}

	time.Sleep(time.Second) // allow all hits to be tracked
	tracker.Flush()

	if len(store.hits) != 7 {
		t.Fatalf("All requests must have been tracked, but was: %v", len(store.hits))
	}
}

func TestTrackerIgnoreHitPrefetch(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Moz", "prefetch")
	tracker := NewTracker(newTestStore(), nil)

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with X-Moz header must be ignored")
	}

	req.Header.Del("X-Moz")
	req.Header.Set("X-Purpose", "prefetch")

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with X-Purpose header must be ignored")
	}

	req.Header.Set("X-Purpose", "preview")

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with X-Purpose header must be ignored")
	}

	req.Header.Del("X-Purpose")
	req.Header.Set("Purpose", "prefetch")

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with Purpose header must be ignored")
	}

	req.Header.Set("Purpose", "preview")

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with Purpose header must be ignored")
	}

	req.Header.Del("Purpose")

	if tracker.ignoreHit(req) {
		t.Fatal("Hit must not be ignored")
	}
}

func TestTrackerIgnoreHitUserAgent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "This is a bot request")
	tracker := NewTracker(newTestStore(), nil)

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a crawler request")

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "This is a spider request")

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with keyword in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Visit http://spam.com!")

	if !tracker.ignoreHit(req) {
		t.Fatal("Hit with URL in User-Agent must be ignored")
	}

	req.Header.Set("User-Agent", "Mozilla/123.0")

	if tracker.ignoreHit(req) {
		t.Fatal("Hit with regular User-Agent must not be ignored")
	}
}
