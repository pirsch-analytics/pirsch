package pirsch

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

const (
	postgresSaveQuery = `INSERT INTO "hit" (fingerprint, path, url, language, user_agent, ref, time) VALUES `
)

// PostgresStore implements the Store interface.
type PostgresStore struct {
	DB *sqlx.DB
}

// NewPostgresStore creates a new postgres storage for given database connection.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{sqlx.NewDb(db, "postgres")}
}

// Save implements the Store interface.
func (store *PostgresStore) Save(hits []Hit) error {
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
		return err
	}

	return nil
}

func (store *PostgresStore) DeleteHitsByDay(day time.Time) error {
	_, err := store.DB.Exec(`DELETE FROM "hit" WHERE time >= $1 AND time < $1 + INTERVAL '1 day'`, day)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) SaveVisitorsPerDay(visitors *VisitorsPerDay) error {
	_, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_day" (day, visitors) VALUES (:day, :visitors)`, visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) SaveVisitorsPerHour(visitors *VisitorsPerHour) error {
	_, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_hour" (day_and_hour, visitors) VALUES (:day_and_hour, :visitors)`, visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) SaveVisitorsPerLanguage(visitors *VisitorsPerLanguage) error {
	_, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_language" (day, language, visitors) VALUES (:day, :language, :visitors)`, visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) SaveVisitorsPerPage(visitors *VisitorsPerPage) error {
	_, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_page" (day, path, visitors) VALUES (:day, :path, :visitors)`, visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) Days() ([]time.Time, error) {
	var days []time.Time

	if err := store.DB.Select(&days, `SELECT DISTINCT date(time) FROM "hit"`); err != nil {
		return nil, err
	}

	return days, nil
}

func (store *PostgresStore) VisitorsPerDay(day time.Time) (int, error) {
	var visitors int

	if err := store.DB.Get(&visitors, `SELECT count(DISTINCT fingerprint) FROM "hit" WHERE time > $1::timestamp AND time < $1::timestamp + INTERVAL '1 day'`, day); err != nil {
		return 0, err
	}

	return visitors, nil
}

func (store *PostgresStore) VisitorsPerDayAndHour(day time.Time) ([]VisitorsPerHour, error) {
	query := `SELECT "day_and_hour", (
			SELECT count(DISTINCT fingerprint) FROM "hit"
			WHERE time >= "day_and_hour"
			AND time < "day_and_hour" + INTERVAL '1 hour'
		) "visitors"
		FROM (
			SELECT * FROM generate_series(
				$1::timestamp,
				$1::timestamp + INTERVAL '23 hours',
				interval '1 hour'
			) "day_and_hour"
		) AS hours`
	var visitors []VisitorsPerHour

	if err := store.DB.Select(&visitors, query, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

func (store *PostgresStore) VisitorsPerLanguage(day time.Time) ([]VisitorsPerLanguage, error) {
	query := `SELECT $1::timestamp "day", "language", count(DISTINCT fingerprint) "visitors"
		FROM hit
		WHERE time >= $1::timestamp
		AND time < $1::timestamp + INTERVAL '1 day'
		GROUP BY "language"`
	var visitors []VisitorsPerLanguage

	if err := store.DB.Select(&visitors, query, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

func (store *PostgresStore) VisitorsPerPage(day time.Time) ([]VisitorsPerPage, error) {
	query := `SELECT $1::timestamp "day", "path", count(DISTINCT fingerprint) "visitors"
		FROM hit
		WHERE time >= $1::timestamp
		AND time < $1::timestamp + INTERVAL '1 day'
		GROUP BY "path"`
	var visitors []VisitorsPerPage

	if err := store.DB.Select(&visitors, query, day); err != nil {
		return nil, err
	}

	return visitors, nil
}
