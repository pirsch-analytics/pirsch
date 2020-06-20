package pirsch

import (
	"database/sql"
	"log"
)

// PostgresStore implements the Store interface.
type PostgresStore struct {
	DB *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db}
}

// TODO batch insert
func (store *PostgresStore) Save(hits []Hit) {
	query := `INSERT INTO "hit" (fingerprint, path, query, fragment, url, language, browser, ref, time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, hit := range hits {
		_, err := store.DB.Exec(query,
			hit.Fingerprint,
			hit.Path,
			hit.Query,
			hit.Fragment,
			hit.URL,
			hit.Language,
			hit.Browser,
			hit.Ref,
			hit.Time)

		if err != nil {
			log.Printf("error saving hit: %s", err)
			break
		}
	}
}
