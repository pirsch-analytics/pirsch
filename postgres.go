package pirsch

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
	"time"
)

// number of arguments to store
const hitParamCount = 8

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
	args := make([]interface{}, 0, len(hits)*hitParamCount)
	var query strings.Builder
	query.WriteString(`INSERT INTO "hit" (tenant_id, fingerprint, path, url, language, user_agent, ref, time) VALUES `)

	for i, hit := range hits {
		args = append(args, hit.TenantID)
		args = append(args, shortenString(hit.Fingerprint, 2000))
		args = append(args, shortenString(hit.Path, 2000))
		args = append(args, shortenString(hit.URL, 2000))
		args = append(args, shortenString(hit.Language, 10))
		args = append(args, shortenString(hit.UserAgent, 200))
		args = append(args, shortenString(hit.Ref, 200))
		args = append(args, hit.Time)
		index := i * hitParamCount
		query.WriteString(fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d),`,
			index+1, index+2, index+3, index+4, index+5, index+6, index+7, index+8))
	}

	queryStr := query.String()
	_, err := store.DB.Exec(queryStr[:len(queryStr)-1], args...)

	if err != nil {
		return err
	}

	return nil
}

// DeleteHitsByDay implements the Store interface.
func (store *PostgresStore) DeleteHitsByDay(tenantID sql.NullInt64, day time.Time) error {
	_, err := store.DB.Exec(`DELETE FROM "hit" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND time >= $2 AND time < $2 + INTERVAL '1 day'`, tenantID, day)

	if err != nil {
		return err
	}

	return nil
}

// SaveVisitorsPerDay implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerDay(visitors *VisitorsPerDay) error {
	day := new(VisitorsPerDay)
	err := store.DB.Get(day, `SELECT * FROM "visitors_per_day" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND date(day) = date($2)`, visitors.TenantID, visitors.Day)

	if err != nil {
		rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_day" (tenant_id, day, visitors) VALUES (:tenant_id, :day, :visitors)`, visitors)

		if err != nil {
			return err
		}

		closeRows(rows)
	} else {
		day.Visitors += visitors.Visitors

		if _, err := store.DB.NamedExec(`UPDATE "visitors_per_day" SET visitors = :visitors WHERE id = :id`, day); err != nil {
			return err
		}
	}

	return nil
}

// SaveVisitorsPerHour implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerHour(visitors *VisitorsPerHour) error {
	day := new(VisitorsPerHour)
	err := store.DB.Get(day, `SELECT * FROM "visitors_per_hour" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND date(day_and_hour) = date($2) AND extract(hour from day_and_hour) = extract(hour from $2)`, visitors.TenantID, visitors.DayAndHour)

	if err != nil {
		rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_hour" (tenant_id, day_and_hour, visitors) VALUES (:tenant_id, :day_and_hour, :visitors)`, visitors)

		if err != nil {
			return err
		}

		closeRows(rows)
	} else {
		day.Visitors += visitors.Visitors

		if _, err := store.DB.NamedExec(`UPDATE "visitors_per_hour" SET visitors = :visitors WHERE id = :id`, day); err != nil {
			return err
		}
	}

	return nil
}

// SaveVisitorsPerLanguage implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerLanguage(visitors *VisitorsPerLanguage) error {
	day := new(VisitorsPerLanguage)
	err := store.DB.Get(day, `SELECT * FROM "visitors_per_language" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND date(day) = date($2) AND language = $3`, visitors.TenantID, visitors.Day, visitors.Language)

	if err != nil {
		rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_language" (tenant_id, day, language, visitors) VALUES (:tenant_id, :day, :language, :visitors)`, visitors)

		if err != nil {
			return err
		}

		closeRows(rows)
	} else {
		day.Visitors += visitors.Visitors

		if _, err := store.DB.NamedExec(`UPDATE "visitors_per_language" SET visitors = :visitors WHERE id = :id`, day); err != nil {
			return err
		}
	}

	return nil
}

// SaveVisitorsPerPage implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerPage(visitors *VisitorsPerPage) error {
	day := new(VisitorsPerPage)
	err := store.DB.Get(day, `SELECT * FROM "visitors_per_page" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND date(day) = date($2) AND path = $3`, visitors.TenantID, visitors.Day, visitors.Path)

	if err != nil {
		rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_page" (tenant_id, day, path, visitors) VALUES (:tenant_id, :day, :path, :visitors)`, visitors)

		if err != nil {
			return err
		}

		closeRows(rows)
	} else {
		day.Visitors += visitors.Visitors

		if _, err := store.DB.NamedExec(`UPDATE "visitors_per_page" SET visitors = :visitors WHERE id = :id`, day); err != nil {
			return err
		}
	}

	return nil
}

// SaveVisitorsPerReferer implements the Store interface.
func (store *PostgresStore) SaveVisitorsPerReferer(visitors *VisitorsPerReferer) error {
	day := new(VisitorsPerReferer)
	err := store.DB.Get(day, `SELECT * FROM "visitors_per_referer" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND date(day) = date($2) AND ref = $3`, visitors.TenantID, visitors.Day, visitors.Ref)

	if err != nil {
		rows, err := store.DB.NamedQuery(`INSERT INTO "visitors_per_referer" (tenant_id, day, ref, visitors) VALUES (:tenant_id, :day, :ref, :visitors)`, visitors)

		if err != nil {
			return err
		}

		closeRows(rows)
	} else {
		day.Visitors += visitors.Visitors

		if _, err := store.DB.NamedExec(`UPDATE "visitors_per_referer" SET visitors = :visitors WHERE id = :id`, day); err != nil {
			return err
		}
	}

	return nil
}

// Days implements the Store interface.
func (store *PostgresStore) Days(tenantID sql.NullInt64) ([]time.Time, error) {
	var days []time.Time

	if err := store.DB.Select(&days, `SELECT DISTINCT date(time) FROM "hit" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND time < current_date`, tenantID); err != nil {
		return nil, err
	}

	return days, nil
}

// VisitorsPerDay implements the Store interface.
func (store *PostgresStore) CountVisitorsPerDay(tenantID sql.NullInt64, day time.Time) (int, error) {
	query := `SELECT count(DISTINCT fingerprint) FROM "hit" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND date("time") = $2`
	var visitors int

	if err := store.DB.Get(&visitors, query, tenantID, day); err != nil {
		return 0, err
	}

	return visitors, nil
}

// VisitorsPerDayAndHour implements the Store interface.
func (store *PostgresStore) CountVisitorsPerDayAndHour(tenantID sql.NullInt64, day time.Time) ([]VisitorsPerHour, error) {
	query := `SELECT "day_and_hour", (
			SELECT count(DISTINCT fingerprint) FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND time >= "day_and_hour"
			AND time < "day_and_hour" + INTERVAL '1 hour'
		) "visitors",
		$1 AS "tenant_id"
		FROM (
			SELECT * FROM generate_series(
				$2::timestamp,
				$2::timestamp + INTERVAL '23 hours',
				interval '1 hour'
			) "day_and_hour"
		) AS hours
		ORDER BY "day_and_hour" ASC`
	var visitors []VisitorsPerHour

	if err := store.DB.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorsPerLanguage implements the Store interface.
func (store *PostgresStore) CountVisitorsPerLanguage(tenantID sql.NullInt64, day time.Time) ([]VisitorsPerLanguage, error) {
	query := `SELECT * FROM (
			SELECT tenant_id, $2::timestamp "day", "language", count(DISTINCT fingerprint) "visitors"
			FROM hit
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND time >= $2::timestamp
			AND time < $2::timestamp + INTERVAL '1 day'
			GROUP BY tenant_id, "language"
		) AS results ORDER BY "day" ASC`
	var visitors []VisitorsPerLanguage

	if err := store.DB.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorsPerPage implements the Store interface.
func (store *PostgresStore) CountVisitorsPerPage(tenantID sql.NullInt64, day time.Time) ([]VisitorsPerPage, error) {
	query := `SELECT * FROM (
			SELECT tenant_id, $2::timestamp "day", "path", count(DISTINCT fingerprint) "visitors"
			FROM hit
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND time >= $2::timestamp
			AND time < $2::timestamp + INTERVAL '1 day'
			GROUP BY tenant_id, "path"
		) AS results ORDER BY "day" ASC, "visitors" DESC`
	var visitors []VisitorsPerPage

	if err := store.DB.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsPerReferer implements the Store interface.
func (store *PostgresStore) CountVisitorsPerReferer(tenantID sql.NullInt64, day time.Time) ([]VisitorsPerReferer, error) {
	query := `SELECT * FROM (
			SELECT tenant_id, $2::timestamp "day", "ref", count(DISTINCT fingerprint) "visitors"
			FROM hit
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND time >= $2::timestamp
			AND time < $2::timestamp + INTERVAL '1 day'
			GROUP BY tenant_id, ref
		) AS results ORDER BY "day" ASC, "visitors" DESC`
	var visitors []VisitorsPerReferer

	if err := store.DB.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// Paths implements the Store interface.
func (store *PostgresStore) Paths(tenantID sql.NullInt64, from, to time.Time) ([]string, error) {
	query := `SELECT * FROM (
			SELECT DISTINCT "path" FROM "visitors_per_page" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND "day" >= $2 AND "day" <= $3
		) AS paths ORDER BY length("path"), "path" ASC`
	var paths []string

	if err := store.DB.Select(&paths, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return paths, nil
}

// Referer implements the Store interface.
func (store *PostgresStore) Referer(tenantID sql.NullInt64, from, to time.Time) ([]string, error) {
	query := `SELECT * FROM (
			SELECT DISTINCT "ref" FROM "visitors_per_referer" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND "day" >= $2 AND "day" <= $3
		) AS referer ORDER BY length("ref"), "ref" ASC`
	var referer []string

	if err := store.DB.Select(&referer, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return referer, nil
}

// Visitors implements the Store interface.
func (store *PostgresStore) Visitors(tenantID sql.NullInt64, from, to time.Time) ([]VisitorsPerDay, error) {
	query := `SELECT tenant_id, "date" "day",
		CASE WHEN "visitors_per_day".visitors IS NULL THEN 0 ELSE "visitors_per_day".visitors END
		FROM (
			SELECT * FROM generate_series(
				$2::timestamp,
				$3::timestamp,
				INTERVAL '1 day'
			) "date"
		) AS date_series
		LEFT JOIN "visitors_per_day" ON ($1::bigint IS NULL OR tenant_id = $1) AND date("visitors_per_day"."day") = date("date")
		ORDER BY "date" ASC`
	var visitors []VisitorsPerDay

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// PageVisits implements the Store interface.
func (store *PostgresStore) PageVisits(tenantID sql.NullInt64, path string, from, to time.Time) ([]VisitorsPerDay, error) {
	query := `SELECT tenant_id, "date" "day",
		CASE WHEN "visitors_per_page".visitors IS NULL THEN 0 ELSE "visitors_per_page".visitors END
		FROM (
			SELECT * FROM generate_series(
				$2::timestamp,
				$3::timestamp,
				INTERVAL '1 day'
			) "date"
		) AS date_series
		LEFT JOIN "visitors_per_page" ON ($1::bigint IS NULL OR tenant_id = $1) AND date("visitors_per_page"."day") = date("date") AND "visitors_per_page"."path" = $4
		ORDER BY "date" ASC`
	var visitors []VisitorsPerDay

	if err := store.DB.Select(&visitors, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// RefererVisits implements the Store interface.
func (store *PostgresStore) RefererVisits(tenantID sql.NullInt64, referer string, from, to time.Time) ([]VisitorsPerReferer, error) {
	query := `SELECT tenant_id, "date" "day",
		CASE WHEN "visitors_per_referer".visitors IS NULL THEN 0 ELSE "visitors_per_referer".visitors END
		FROM (
			SELECT * FROM generate_series(
				$2::timestamp,
				$3::timestamp,
				INTERVAL '1 day'
			) "date"
		) AS date_series
		LEFT JOIN "visitors_per_referer" ON ($1::bigint IS NULL OR tenant_id = $1) AND date("visitors_per_referer"."day") = date("date") AND "visitors_per_referer"."ref" = $4
		ORDER BY "date" ASC`
	var visitors []VisitorsPerReferer

	if err := store.DB.Select(&visitors, query, tenantID, from, to, referer); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorLanguages implements the Store interface.
func (store *PostgresStore) VisitorLanguages(tenantID sql.NullInt64, from, to time.Time) ([]VisitorLanguage, error) {
	query := `SELECT * FROM (
			SELECT "language", sum("visitors") "visitors" FROM (
				SELECT lower("language") "language", sum("visitors") "visitors" FROM "visitors_per_language"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= date($2::timestamp)
				AND "day" <= date($3::timestamp)
				GROUP BY "language"
				UNION
				SELECT lower("language") "language", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				GROUP BY "language"
			) AS results
			GROUP BY "language"
		) AS langs
		ORDER BY "visitors" DESC`
	var languages []VisitorLanguage

	if err := store.DB.Select(&languages, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return languages, nil
}

// VisitorLanguages implements the Store interface.
func (store *PostgresStore) HourlyVisitors(tenantID sql.NullInt64, from, to time.Time) ([]HourlyVisitors, error) {
	query := `SELECT * FROM (
			SELECT "hour", sum("visitors") "visitors" FROM (
				SELECT EXTRACT(HOUR FROM "day_and_hour") "hour", sum("visitors") "visitors" FROM "visitors_per_hour"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("day_and_hour") >= date($2::timestamp)
				AND date("day_and_hour") <= date($3::timestamp)
				GROUP BY "hour"
				UNION
				SELECT EXTRACT(HOUR FROM "time") "hour", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				GROUP BY "hour"
			) AS results
			GROUP BY "hour"
		) AS hours
		ORDER BY "hour" ASC`
	var visitors []HourlyVisitors

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// ActiveVisitors implements the Store interface.
func (store *PostgresStore) ActiveVisitors(tenantID sql.NullInt64, from time.Time) (int, error) {
	query := `SELECT count(DISTINCT fingerprint) FROM "hit" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND "time" > $2`
	var visitors int

	if err := store.DB.Get(&visitors, query, tenantID, from); err != nil {
		return 0, err
	}

	return visitors, nil
}

// ActiveVisitorsPerPage implements the Store interface.
func (store *PostgresStore) ActiveVisitorsPerPage(tenantID sql.NullInt64, from time.Time) ([]PageVisitors, error) {
	query := `SELECT "path", count(DISTINCT fingerprint) AS "visitors"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "time" > $2
		GROUP BY "path"
		ORDER BY "visitors" DESC`
	var visitors []PageVisitors

	if err := store.DB.Select(&visitors, query, tenantID, from); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountHits implements the Store interface.
func (store *PostgresStore) CountHits(tenantID sql.NullInt64) int {
	var count int

	if err := store.DB.Get(&count, `SELECT COUNT(1) FROM "hit" WHERE ($1::bigint IS NULL OR tenant_id = $1)`, tenantID); err != nil {
		return 0
	}

	return count
}

// VisitorsPerDay implements the Store interface.
func (store *PostgresStore) VisitorsPerDay(tenantID sql.NullInt64) []VisitorsPerDay {
	var entities []VisitorsPerDay

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_day" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day"`, tenantID); err != nil {
		return nil
	}

	return entities
}

// VisitorsPerHour implements the Store interface.
func (store *PostgresStore) VisitorsPerHour(tenantID sql.NullInt64) []VisitorsPerHour {
	var entities []VisitorsPerHour

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_hour" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day_and_hour"`, tenantID); err != nil {
		return nil
	}

	return entities
}

// VisitorsPerLanguage implements the Store interface.
func (store *PostgresStore) VisitorsPerLanguage(tenantID sql.NullInt64) []VisitorsPerLanguage {
	var entities []VisitorsPerLanguage

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_language" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day", "language"`, tenantID); err != nil {
		return nil
	}

	return entities
}

// VisitorsPerPage implements the Store interface.
func (store *PostgresStore) VisitorsPerPage(tenantID sql.NullInt64) []VisitorsPerPage {
	var entities []VisitorsPerPage

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_page" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day" ASC, "visitors" DESC`, tenantID); err != nil {
		return nil
	}

	return entities
}

// VisitorsPerReferer implements the Store interface.
func (store *PostgresStore) VisitorsPerReferer(tenantID sql.NullInt64) []VisitorsPerReferer {
	var entities []VisitorsPerReferer

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_referer" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day" ASC, "visitors" DESC`, tenantID); err != nil {
		return nil
	}

	return entities
}

func closeRows(rows *sqlx.Rows) {
	if err := rows.Close(); err != nil {
		log.Printf("error closing rows: %s", err)
	}
}
