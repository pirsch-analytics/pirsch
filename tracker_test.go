package pirsch

import (
	"database/sql"
	"log"
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
		len(cfg.RefererDomainBlacklist) != 0 ||
		cfg.RefererDomainBlacklistIncludesSubdomains {
		t.Fatal("TrackerConfig must have default values")
	}

	cfg = &TrackerConfig{
		Worker:                                   123,
		WorkerBufferSize:                         42,
		WorkerTimeout:                            time.Second * 57,
		RefererDomainBlacklist:                   []string{"localhost"},
		RefererDomainBlacklistIncludesSubdomains: true,
	}
	cfg.validate()

	if cfg.Worker != 123 ||
		cfg.WorkerBufferSize != 42 ||
		cfg.WorkerTimeout != time.Second*57 ||
		len(cfg.RefererDomainBlacklist) != 1 ||
		!cfg.RefererDomainBlacklistIncludesSubdomains {
		t.Fatal("TrackerConfig must have set values")
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

	time.Sleep(time.Second) // allow all hits to be tracked
	tracker.Stop()

	if len(store.hits) != 7 {
		t.Fatalf("All requests must have been tracked, but was: %v", len(store.hits))
	}
}

type testStore struct {
	hits []Hit
}

func newTestStore() *testStore {
	return &testStore{make([]Hit, 0)}
}

func (store *testStore) Save(hits []Hit) error {
	log.Printf("Saved %d hits", len(hits))
	store.hits = append(store.hits, hits...)
	return nil
}

func (store *testStore) DeleteHitsByDay(tenantID sql.NullInt64, t time.Time) error {
	panic("implement me")
}

func (store *testStore) SaveVisitorsPerDay(day *VisitorsPerDay) error {
	panic("implement me")
}

func (store *testStore) SaveVisitorsPerHour(hour *VisitorsPerHour) error {
	panic("implement me")
}

func (store *testStore) SaveVisitorsPerLanguage(language *VisitorsPerLanguage) error {
	panic("implement me")
}

func (store *testStore) SaveVisitorsPerPage(page *VisitorsPerPage) error {
	panic("implement me")
}

func (store *testStore) SaveVisitorsPerReferer(page *VisitorsPerReferer) error {
	panic("implement me")
}

func (store *testStore) Days(tenantID sql.NullInt64) ([]time.Time, error) {
	panic("implement me")
}

func (store *testStore) CountVisitorsPerDay(tenantID sql.NullInt64, t time.Time) (int, error) {
	panic("implement me")
}

func (store *testStore) CountVisitorsPerDayAndHour(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerHour, error) {
	panic("implement me")
}

func (store *testStore) CountVisitorsPerLanguage(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerLanguage, error) {
	panic("implement me")
}

func (store *testStore) CountVisitorsPerPage(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerPage, error) {
	panic("implement me")
}

func (store *testStore) CountVisitorsPerReferer(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerReferer, error) {
	panic("implement me")
}

func (store *testStore) Paths(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]string, error) {
	panic("implement me")
}

func (store *testStore) Referer(nullInt64 sql.NullInt64, t time.Time, t2 time.Time) ([]string, error) {
	panic("implement me")
}

func (store *testStore) Visitors(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]VisitorsPerDay, error) {
	panic("implement me")
}

func (store *testStore) PageVisits(tenantID sql.NullInt64, s string, t time.Time, t2 time.Time) ([]VisitorsPerDay, error) {
	panic("implement me")
}

func (store *testStore) RefererVisits(tenantID sql.NullInt64, s string, t time.Time, t2 time.Time) ([]VisitorsPerReferer, error) {
	panic("implement me")
}

func (store *testStore) VisitorLanguages(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]VisitorLanguage, error) {
	panic("implement me")
}

func (store *testStore) HourlyVisitors(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]HourlyVisitors, error) {
	panic("implement me")
}

func (store *testStore) ActiveVisitors(tenantID sql.NullInt64, t time.Time) (int, error) {
	panic("implement me")
}

func (store *testStore) ActiveVisitorsPerPage(tenantID sql.NullInt64, t time.Time) ([]PageVisitors, error) {
	panic("implement me")
}

func (store *testStore) CountHits(tenantID sql.NullInt64) int {
	panic("implement me")
	return 0
}

func (store *testStore) VisitorsPerDay(tenantID sql.NullInt64) []VisitorsPerDay {
	panic("implement me")
	return nil
}

func (store *testStore) VisitorsPerHour(tenantID sql.NullInt64) []VisitorsPerHour {
	panic("implement me")
	return nil
}

func (store *testStore) VisitorsPerLanguage(tenantID sql.NullInt64) []VisitorsPerLanguage {
	panic("implement me")
	return nil
}

func (store *testStore) VisitorsPerPage(tenantID sql.NullInt64) []VisitorsPerPage {
	panic("implement me")
	return nil
}

func (store *testStore) VisitorsPerReferer(tenantID sql.NullInt64) []VisitorsPerReferer {
	panic("implement me")
	return nil
}
