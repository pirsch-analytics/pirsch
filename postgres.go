package pirsch

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
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
	_, err := store.DB.Exec(`INSERT INTO "visitors_per_day" (day, visitors) VALUES ($1, $2)`, visitors.Day, visitors.Visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) SaveVisitorsPerHour(visitors *VisitorsPerHour) error {
	_, err := store.DB.Exec(`INSERT INTO "visitors_per_hour" (day_and_hour, visitors) VALUES ($1, $2)`, visitors.DayAndHour, visitors.Visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) SaveVisitorsPerLanguage(visitors *VisitorsPerLanguage) error {
	_, err := store.DB.Exec(`INSERT INTO "visitor_per_language" (day, language, visitors) VALUES ($1, $2, $3)`, visitors.Day, visitors.Language, visitors.Visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) SaveVisitorsPerPage(visitors *VisitorsPerPage) error {
	_, err := store.DB.Exec(`INSERT INTO "visitor_per_page" (day, path, visitors) VALUES ($1, $2, $3)`, visitors.Day, visitors.Path, visitors.Visitors)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostgresStore) Days() ([]time.Time, error) {
	rows, err := store.DB.Query(`SELECT DISTINCT date(time) FROM "hit"`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	days := make([]time.Time, 0)

	for rows.Next() {
		var day time.Time

		if err := rows.Scan(&day); err != nil {
			return nil, err
		}

		days = append(days, day)
	}

	return days, nil
}

func (store *PostgresStore) VisitorsPerDay(day time.Time) (int, error) {
	row := store.DB.QueryRow(`SELECT count(DISTINCT fingerprint) FROM "hit" WHERE time > $1 AND time < $1 + INTERVAL '1 day'`, day)
	var visitors int

	if err := row.Scan(&visitors); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return visitors, nil
}

func (store *PostgresStore) VisitorsPerDayAndHour(day time.Time) ([]VisitorsPerHour, error) {
	rows, err := store.DB.Query(`SELECT "day_and_hour", (
			SELECT count(DISTINCT fingerprint) FROM "hit"
			WHERE time >= "day_and_hour"
			AND time < "day_and_hour" + INTERVAL '1 hour'
		) "visitors"
		FROM (
			SELECT * FROM generate_series(
				$1,
				$1 + INTERVAL '23 hours',
				interval '1 hour'
			) "day_and_hour"
		) AS hours`, day)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	visitors := make([]VisitorsPerHour, 0)

	for rows.Next() {
		var visitor VisitorsPerHour

		if err := rows.Scan(&visitor.DayAndHour, &visitor.Visitors); err != nil {
			return nil, err
		}

		visitors = append(visitors, visitor)
	}

	return visitors, nil
}

func (store *PostgresStore) VisitorsPerLanguage(day time.Time) ([]VisitorsPerLanguage, error) {
	rows, err := store.DB.Query(`SELECT $1 "day", "language", count(DISTINCT fingerprint) "visitors"
		FROM hit
		WHERE time >= $1
		AND time < $1 + INTERVAL '1 day'
		GROUP BY "language"`, day)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	visitors := make([]VisitorsPerLanguage, 0)

	for rows.Next() {
		var visitor VisitorsPerLanguage

		if err := rows.Scan(&visitor.Day, &visitor.Language, &visitor.Visitors); err != nil {
			return nil, err
		}

		visitors = append(visitors, visitor)
	}

	return visitors, nil
}

func (store *PostgresStore) VisitorsPerPage(day time.Time) ([]VisitorsPerPage, error) {
	rows, err := store.DB.Query(`SELECT $1 "day", "path", count(DISTINCT fingerprint) "visitors"
		FROM hit
		WHERE time >= $1
		AND time < $1 + INTERVAL '1 day'
		GROUP BY "path"`, day)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	visitors := make([]VisitorsPerPage, 0)

	for rows.Next() {
		var visitor VisitorsPerPage

		if err := rows.Scan(&visitor.Day, &visitor.Path, &visitor.Visitors); err != nil {
			return nil, err
		}

		visitors = append(visitors, visitor)
	}

	return visitors, nil
}
