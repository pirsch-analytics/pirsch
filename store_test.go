package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

// This is a list of all storage backends to be used in tests.
// We test against real databases. To test all storage solutions, they must be installed an configured.
func testStorageBackends() []Store {
	return []Store{
		NewPostgresStore(postgresDB, nil),
	}
}

type storeMock struct {
	hits []Hit
}

func newTestStore() *storeMock {
	return &storeMock{make([]Hit, 0)}
}

func (store *storeMock) NewTx() *sqlx.Tx {
	return nil
}

func (store *storeMock) Commit(tx *sqlx.Tx) {
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func (store *storeMock) Rollback(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil {
		panic(err)
	}
}

func (store *storeMock) SaveHits(hits []Hit) error {
	log.Printf("Saved %d hits", len(hits))
	store.hits = append(store.hits, hits...)
	return nil
}

func (store *storeMock) DeleteHitsByDay(tx *sqlx.Tx, tenantID sql.NullInt64, t time.Time) error {
	panic("implement me")
}

func (store *storeMock) Days(tenantID sql.NullInt64) ([]time.Time, error) {
	panic("implement me")
}

func (store *storeMock) Paths(tenantID sql.NullInt64, day time.Time) ([]string, error) {
	panic("implement me")
}

func (store *storeMock) SaveVisitorStats(tx *sqlx.Tx, entity *VisitorStats) error {
	panic("implement me")
}

func (store *storeMock) SaveVisitorTimeStats(tx *sqlx.Tx, entity *VisitorTimeStats) error {
	panic("implement me")
}

func (store *storeMock) SaveLanguageStats(tx *sqlx.Tx, entity *LanguageStats) error {
	panic("implement me")
}

func (store *storeMock) SaveReferrerStats(tx *sqlx.Tx, entity *ReferrerStats) error {
	panic("implement me")
}

func (store *storeMock) SaveOSStats(tx *sqlx.Tx, entity *OSStats) error {
	panic("implement me")
}

func (store *storeMock) SaveBrowserStats(tx *sqlx.Tx, entity *BrowserStats) error {
	panic("implement me")
}

func (store *storeMock) CountVisitorsByPath(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]VisitorStats, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsByPathAndHour(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]VisitorTimeStats, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsByPathAndLanguage(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]LanguageStats, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsByPathAndReferrer(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]ReferrerStats, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsByPathAndOS(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]OSStats, error) {
	panic("implement me")
}

func (store *storeMock) CountVisitorsByPathAndBrowser(*sqlx.Tx, sql.NullInt64, time.Time, string) ([]BrowserStats, error) {
	panic("implement me")
}
