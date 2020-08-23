package pirsch

import (
	"database/sql"
	"log"
	"time"
)

// This is a list of all storage backends to be used in tests.
// We test against real databases. To test all storage solutions, they must be installed an configured.
func testStorageBackends() []Store {
	return []Store{
		NewPostgresStore(postgresDB),
	}
}

type storeMock struct {
	hits []Hit
}

func newTestStore() *storeMock {
	return &storeMock{make([]Hit, 0)}
}

func (store *storeMock) Save(hits []Hit) error {
	log.Printf("Saved %d hits", len(hits))
	store.hits = append(store.hits, hits...)
	return nil
}

func (store *storeMock) DeleteHitsByDay(tenantID sql.NullInt64, t time.Time) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorsPerDay(day *VisitorsPerDay) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorsPerHour(hour *VisitorsPerHour) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorsPerLanguage(language *VisitorsPerLanguage) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorsPerPage(page *VisitorsPerPage) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorsPerReferrer(page *VisitorsPerReferrer) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorsPerOS(visitors *VisitorsPerOS) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorsPerBrowser(visitors *VisitorsPerBrowser) error {
	panic("implement me")
}

func (store *storeMock) Days(tenantID sql.NullInt64) ([]time.Time, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsPerDay(tenantID sql.NullInt64, t time.Time) (int, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsPerDayAndHour(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerHour, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsPerLanguage(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerLanguage, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsPerPage(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerPage, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsPerReferrer(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerReferrer, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsPerOSAndVersion(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerOS, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsPerBrowserAndVersion(tenantID sql.NullInt64, t time.Time) ([]VisitorsPerBrowser, error) {
	panic("implement me")
}

func (store *storeMock) Paths(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]string, error) {
	panic("implement me")
}

func (store *storeMock) Referrer(nullInt64 sql.NullInt64, t time.Time, t2 time.Time) ([]string, error) {
	panic("implement me")
}

func (store *storeMock) Visitors(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]VisitorsPerDay, error) {
	panic("implement me")
}

func (store *storeMock) PageVisits(tenantID sql.NullInt64, s string, t time.Time, t2 time.Time) ([]VisitorsPerDay, error) {
	panic("implement me")
}

func (store *storeMock) ReferrerVisits(tenantID sql.NullInt64, s string, t time.Time, t2 time.Time) ([]VisitorsPerReferrer, error) {
	panic("implement me")
}

func (store *storeMock) VisitorPages(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]Stats, error) {
	panic("implement me")
}

func (store *storeMock) VisitorLanguages(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]Stats, error) {
	panic("implement me")
}

func (store *storeMock) VisitorReferrer(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]Stats, error) {
	panic("implement me")
}

func (store *storeMock) VisitorOS(tenantID sql.NullInt64, from time.Time, to time.Time) ([]Stats, error) {
	panic("implement me")
}

func (store *storeMock) VisitorBrowser(tenantID sql.NullInt64, from time.Time, to time.Time) ([]Stats, error) {
	panic("implement me")
}

func (store *storeMock) HourlyVisitors(tenantID sql.NullInt64, t time.Time, t2 time.Time) ([]Stats, error) {
	panic("implement me")
}

func (store *storeMock) ActiveVisitors(tenantID sql.NullInt64, t time.Time) (int, error) {
	panic("implement me")
}

func (store *storeMock) ActiveVisitorsPerPage(tenantID sql.NullInt64, t time.Time) ([]Stats, error) {
	panic("implement me")
}

func (store *storeMock) CountHits(tenantID sql.NullInt64) int {
	panic("implement me")
}

func (store *storeMock) VisitorsPerDay(tenantID sql.NullInt64) []VisitorsPerDay {
	panic("implement me")
}

func (store *storeMock) VisitorsPerHour(tenantID sql.NullInt64) []VisitorsPerHour {
	panic("implement me")
}

func (store *storeMock) VisitorsPerLanguage(tenantID sql.NullInt64) []VisitorsPerLanguage {
	panic("implement me")
}

func (store *storeMock) VisitorsPerPage(tenantID sql.NullInt64) []VisitorsPerPage {
	panic("implement me")
}

func (store *storeMock) VisitorsPerReferrer(tenantID sql.NullInt64) []VisitorsPerReferrer {
	panic("implement me")
}

func (store *storeMock) VisitorsPerOS(tenantID sql.NullInt64) []VisitorsPerOS {
	panic("implement me")
}

func (store *storeMock) VisitorsPerBrowser(tenantID sql.NullInt64) []VisitorsPerBrowser {
	panic("implement me")
}
