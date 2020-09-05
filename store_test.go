package pirsch

import (
	"github.com/jmoiron/sqlx"
	"log"
)

// This is a list of all storage backends to be used in tests.
// We test against real databases. To test all storage solutions, they must be installed an configured.
func testStorageBackends() []Store {
	return []Store{
		NewPostgresStore(postgresDB, nil),
	}
}

// storeMock is a Store implementation for testing purposes.
// It overwrites all functions required for testing using PostgresStore as a base implementation.
type storeMock struct {
	PostgresStore

	hits []Hit
}

func newTestStore() *storeMock {
	return &storeMock{hits: make([]Hit, 0)}
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
