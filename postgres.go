package pirsch

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

const (
	postgresSaveQuery = `INSERT INTO "hit" (fingerprint, path, url, language, user_agent, ref, time) VALUES `
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
	args := make([]interface{}, 0, len(hits)*7)
	var query strings.Builder
	query.WriteString(postgresSaveQuery)

	for i, hit := range hits {
		args = append(args, hit.Fingerprint)
		args = append(args, hit.Path)
		args = append(args, hit.URL)
		args = append(args, hit.Language)
		args = append(args, hit.UserAgent)
		args = append(args, hit.Ref)
		args = append(args, hit.Time)
		index := i * 7
		query.WriteString(fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d),`,
			index+1, index+2, index+3, index+4, index+5, index+6, index+7))
	}

	queryStr := query.String()
	_, err := store.DB.Exec(queryStr[:len(queryStr)-1], args...)

	if err != nil {
		log.Printf("error saving hits: %s", err)
	}
}
