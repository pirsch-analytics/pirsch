package pirsch

import (
	"database/sql"
	"log"
)

const (
	postgresSaveQuery = `INSERT INTO "hit" (fingerprint, path, query, fragment, url, language, browser, ref, time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
)

// PostgresStore implements the Store interface.
type PostgresStore struct {
	DB *sql.DB
}

// NewPostgresStore creates a new postgres storage for given database connection.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db}
}

// Save implements the Store interface.
func (store *PostgresStore) Save(hits []Hit) {
	// TODO batch insert
	for _, hit := range hits {
		_, err := store.DB.Exec(postgresSaveQuery,
			hit.Fingerprint,
			hit.Path,
			hit.Query,
			hit.Fragment,
			hit.URL,
			hit.Language,
			hit.UserAgent,
			hit.Ref,
			hit.Time)

		if err != nil {
			log.Printf("error saving hit: %s", err)
			break
		}
	}
}
