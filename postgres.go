package pirsch

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
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
		args = append(args, shortenString(hit.Fingerprint, 2000))
		args = append(args, shortenString(hit.Path, 2000))
		args = append(args, shortenString(hit.URL, 2000))
		args = append(args, shortenString(hit.Language, 10))
		args = append(args, shortenString(hit.UserAgent, 200))
		args = append(args, shortenString(hit.Ref, 200))
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

// DeleteHitsByDay implements the Store interface.
func (store *PostgresStore) DeleteHitsByDay(day time.Time) error {
	_, err := store.DB.Exec(`DELETE FROM "hit" WHERE time >= $1 AND time < $1 + INTERVAL '1 day'`, day)

	if err != nil {
		return err
	}

	return nil
}

// SaveVisitorsPerDay implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerDay(visitors *VisitorsPerDay) error {
	rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_day" (day, visitors) VALUES (:day, :visitors)`, visitors)

	if err != nil {
		return err
	}

	closeRows(rows)
	return nil
}

// SaveVisitorsPerHour implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerHour(visitors *VisitorsPerHour) error {
	rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_hour" (day_and_hour, visitors) VALUES (:day_and_hour, :visitors)`, visitors)

	if err != nil {
		return err
	}

	closeRows(rows)
	return nil
}

// SaveVisitorsPerLanguage implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerLanguage(visitors *VisitorsPerLanguage) error {
	rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_language" (day, language, visitors) VALUES (:day, :language, :visitors)`, visitors)

	if err != nil {
		return err
	}

	closeRows(rows)
	return nil
}

// SaveVisitorsPerPage implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerPage(visitors *VisitorsPerPage) error {
	rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_page" (day, path, visitors) VALUES (:day, :path, :visitors)`, visitors)

	if err != nil {
		return err
	}

	closeRows(rows)
	return nil
}

// Days implements the Store interface.
func (store *PostgresStore) Days() ([]time.Time, error) {
	var days []time.Time

	if err := store.DB.Select(&days, `SELECT DISTINCT date(time) FROM "hit" WHERE time < current_date`); err != nil {
		return nil, err
	}

	return days, nil
}

// VisitorsPerDay implements the Store interface.
func (store *PostgresStore) VisitorsPerDay(day time.Time) (int, error) {
	query := `SELECT count(DISTINCT fingerprint) FROM "hit" WHERE date("time") = $1`
	var visitors int

	if err := store.DB.Get(&visitors, query, day); err != nil {
		return 0, err
	}

	return visitors, nil
}

// VisitorsPerDayAndHour implements the Store interface.
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
		) AS hours
		ORDER BY "day_and_hour" ASC`
	var visitors []VisitorsPerHour

	if err := store.DB.Select(&visitors, query, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorsPerLanguage implements the Store interface.
func (store *PostgresStore) VisitorsPerLanguage(day time.Time) ([]VisitorsPerLanguage, error) {
	query := `SELECT * FROM (
			SELECT $1::timestamp "day", "language", count(DISTINCT fingerprint) "visitors"
			FROM hit
			WHERE time >= $1::timestamp
			AND time < $1::timestamp + INTERVAL '1 day'
			GROUP BY "language"
		) AS results ORDER BY "day" ASC`
	var visitors []VisitorsPerLanguage

	if err := store.DB.Select(&visitors, query, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorsPerPage implements the Store interface.
func (store *PostgresStore) VisitorsPerPage(day time.Time) ([]VisitorsPerPage, error) {
	query := `SELECT * FROM (
			SELECT $1::timestamp "day", "path", count(DISTINCT fingerprint) "visitors"
			FROM hit
			WHERE time >= $1::timestamp
			AND time < $1::timestamp + INTERVAL '1 day'
			GROUP BY "path"
		) AS results ORDER BY "day", "path" ASC`
	var visitors []VisitorsPerPage

	if err := store.DB.Select(&visitors, query, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// Paths implements the Store interface.
func (store *PostgresStore) Paths(from, to time.Time) ([]string, error) {
	query := `SELECT * FROM (
			SELECT DISTINCT "path" FROM "visitors_per_page" WHERE "day" >= $1 AND "day" <= $2
		) AS paths ORDER BY length("path"), "path" ASC`
	var paths []string

	if err := store.DB.Select(&paths, query, from, to); err != nil {
		return nil, err
	}

	return paths, nil
}

// Visitors implements the Store interface.
func (store *PostgresStore) Visitors(from, to time.Time) ([]VisitorsPerDay, error) {
	query := `SELECT "date" "day",
		CASE WHEN "visitors_per_day".visitors IS NULL THEN 0 ELSE "visitors_per_day".visitors END
		FROM (
			SELECT * FROM generate_series(
				$1::timestamp,
				$2::timestamp,
				INTERVAL '1 day'
			) "date"
		) AS date_series
		LEFT JOIN "visitors_per_day" ON date("visitors_per_day"."day") = date("date")
		ORDER BY "date" ASC`
	var visitors []VisitorsPerDay

	if err := store.DB.Select(&visitors, query, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// PageVisits implements the Store interface.
func (store *PostgresStore) PageVisits(path string, from, to time.Time) ([]VisitorsPerDay, error) {
	query := `SELECT "date" "day",
		CASE WHEN "visitors_per_page".visitors IS NULL THEN 0 ELSE "visitors_per_page".visitors END
		FROM (
			SELECT * FROM generate_series(
				$1::timestamp,
				$2::timestamp,
				INTERVAL '1 day'
			) "date"
		) AS date_series
		LEFT JOIN "visitors_per_page" ON date("visitors_per_page"."day") = date("date") AND "visitors_per_page"."path" = $3
		ORDER BY "date" ASC`
	var visitors []VisitorsPerDay

	if err := store.DB.Select(&visitors, query, from, to, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorLanguages implements the Store interface.
func (store *PostgresStore) VisitorLanguages(from, to time.Time) ([]VisitorLanguage, error) {
	query := `SELECT * FROM (
			SELECT "language", sum("visitors") "visitors" FROM (
				SELECT lower("language") "language", sum("visitors") "visitors" FROM "visitors_per_language"
				WHERE "day" >= date($1::timestamp)
				AND "day" <= date($2::timestamp)
				GROUP BY "language"
				UNION
				SELECT lower("language") "language", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE date("time") >= date($1::timestamp)
				AND date("time") <= date($2::timestamp)
				GROUP BY "language"
			) AS results
			GROUP BY "language"
		) AS langs
		ORDER BY "visitors" DESC`
	var languages []VisitorLanguage

	if err := store.DB.Select(&languages, query, from, to); err != nil {
		return nil, err
	}

	return languages, nil
}

// VisitorLanguages implements the Store interface.
func (store *PostgresStore) HourlyVisitors(from, to time.Time) ([]HourlyVisitors, error) {
	query := `SELECT * FROM (
			SELECT "hour", sum("visitors") "visitors" FROM (
				SELECT EXTRACT(HOUR FROM "day_and_hour") "hour", sum("visitors") "visitors" FROM "visitors_per_hour"
				WHERE date("day_and_hour") >= date($1::timestamp)
				AND date("day_and_hour") <= date($2::timestamp)
				GROUP BY "hour"
				UNION
				SELECT EXTRACT(HOUR FROM "time") "hour", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE date("time") >= date($1::timestamp)
				AND date("time") <= date($2::timestamp)
				GROUP BY "hour"
			) AS results
			GROUP BY "hour"
		) AS hours
		ORDER BY "hour" ASC`
	var visitors []HourlyVisitors

	if err := store.DB.Select(&visitors, query, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// ActiveVisitors implements the Store interface.
func (store *PostgresStore) ActiveVisitors(from time.Time) (int, error) {
	query := `SELECT count(DISTINCT fingerprint) FROM "hit" WHERE "time" > $1`
	var visitors int

	if err := store.DB.Get(&visitors, query, from); err != nil {
		return 0, err
	}

	return visitors, nil
}

func closeRows(rows *sqlx.Rows) {
	if err := rows.Close(); err != nil {
		log.Printf("error closing rows: %s", err)
	}
}
