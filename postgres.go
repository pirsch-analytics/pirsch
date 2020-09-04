package pirsch

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"strings"
	"time"
)

const (
	logPrefix = "[pirsch] "
)

// PostgresConfig is the optional configuration for the PostgresStore.
type PostgresConfig struct {
	// Logger is the log.Logger used for logging.
	// The default log will be used printing to os.Stdout with "pirsch" in its prefix in case it is not set.
	Logger *log.Logger
}

// PostgresStore implements the Store interface.
type PostgresStore struct {
	DB     *sqlx.DB
	logger *log.Logger
}

// NewPostgresStore creates a new postgres storage for given database connection and logger.
func NewPostgresStore(db *sql.DB, config *PostgresConfig) *PostgresStore {
	if config == nil {
		config = &PostgresConfig{
			Logger: log.New(os.Stdout, logPrefix, log.LstdFlags),
		}
	}

	return &PostgresStore{
		DB:     sqlx.NewDb(db, "postgres"),
		logger: config.Logger,
	}
}

// NewTx implements the Store interface.
func (store *PostgresStore) NewTx() *sqlx.Tx {
	tx, err := store.DB.Beginx()

	if err != nil {
		store.logger.Fatalf("error creating new transaction: %s", err)
	}

	return tx
}

// Commit implements the Store interface.
func (store *PostgresStore) Commit(tx *sqlx.Tx) {
	if err := tx.Commit(); err != nil {
		store.logger.Printf("error committing transaction: %s", err)
	}
}

// Rollback implements the Store interface.
func (store *PostgresStore) Rollback(tx *sqlx.Tx) {
	if err := tx.Rollback(); err != nil {
		store.logger.Printf("error rolling back transaction: %s", err)
	}
}

// Save implements the Store interface.
func (store *PostgresStore) SaveHits(hits []Hit) error {
	args := make([]interface{}, 0, len(hits)*14)
	var query strings.Builder
	query.WriteString(`INSERT INTO "hit" (tenant_id, fingerprint, path, url, language, user_agent, referrer, os, os_version, browser, browser_version, desktop, mobile, time) VALUES `)

	for i, hit := range hits {
		args = append(args, hit.TenantID)
		args = append(args, hit.Fingerprint)
		args = append(args, hit.Path)
		args = append(args, hit.URL)
		args = append(args, hit.Language)
		args = append(args, hit.UserAgent)
		args = append(args, hit.Referrer)
		args = append(args, hit.OS)
		args = append(args, hit.OSVersion)
		args = append(args, hit.Browser)
		args = append(args, hit.BrowserVersion)
		args = append(args, hit.Desktop)
		args = append(args, hit.Mobile)
		args = append(args, hit.Time)
		index := i * 14
		query.WriteString(fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d),`,
			index+1, index+2, index+3, index+4, index+5, index+6, index+7, index+8, index+9, index+10, index+11, index+12, index+13, index+14))
	}

	queryStr := query.String()
	_, err := store.DB.Exec(queryStr[:len(queryStr)-1], args...)

	if err != nil {
		return err
	}

	return nil
}

// DeleteHitsByDay implements the Store interface.
func (store *PostgresStore) DeleteHitsByDay(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `DELETE FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND time >= $2
		AND time < $2 + INTERVAL '1 day'`

	_, err := tx.Exec(query, tenantID, day)

	if err != nil {
		return err
	}

	return nil
}

// SaveVisitorStats implements the Store interface.
func (store *PostgresStore) SaveVisitorStats(tx *sqlx.Tx, entity *VisitorStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(VisitorStats)
	err := tx.Get(existing, `SELECT id, visitors, platform_desktop, platform_mobile, platform_unknown FROM "visitor_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND LOWER("path") = LOWER($3)`, entity.TenantID, entity.Day, entity.Path)

	if err == nil {
		existing.Visitors += entity.Visitors
		existing.PlatformDesktop += entity.PlatformDesktop
		existing.PlatformMobile += entity.PlatformMobile
		existing.PlatformUnknown += entity.PlatformUnknown

		if _, err := tx.Exec(`UPDATE "visitor_stats" SET visitors = $1, platform_desktop = $2, platform_mobile = $3, platform_unknown = $4 WHERE id = $5`,
			existing.Visitors,
			existing.PlatformDesktop,
			existing.PlatformMobile,
			existing.PlatformUnknown,
			existing.ID); err != nil {
			return err
		}
	} else {
		rows, err := tx.NamedQuery(`INSERT INTO "visitor_stats" ("tenant_id", "day", "path", "visitors", "platform_desktop", "platform_mobile", "platform_unknown") VALUES (:tenant_id, :day, :path, :visitors, :platform_desktop, :platform_mobile, :platform_unknown)`, entity)

		if err != nil {
			return err
		}

		store.closeRows(rows)
	}

	return nil
}

// SaveVisitorTimeStats implements the Store interface.
func (store *PostgresStore) SaveVisitorTimeStats(tx *sqlx.Tx, entity *VisitorTimeStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(VisitorTimeStats)
	err := tx.Get(existing, `SELECT id, visitors FROM "visitor_time_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND LOWER("path") = LOWER($3)
		AND "hour" = $4`, entity.TenantID, entity.Day, entity.Path, entity.Hour)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "visitor_time_stats" ("tenant_id", "day", "path", "hour", "visitors") VALUES (:tenant_id, :day, :path, :hour, :visitors)`,
		`UPDATE "visitor_time_stats" SET visitors = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// SaveLanguageStats implements the Store interface.
func (store *PostgresStore) SaveLanguageStats(tx *sqlx.Tx, entity *LanguageStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(LanguageStats)
	err := tx.Get(existing, `SELECT id, visitors FROM "language_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND LOWER("path") = LOWER($3)
		AND LOWER("language") = LOWER($4)`, entity.TenantID, entity.Day, entity.Path, entity.Language)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "language_stats" ("tenant_id", "day", "path", "language", "visitors") VALUES (:tenant_id, :day, :path, :language, :visitors)`,
		`UPDATE "language_stats" SET visitors = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// SaveReferrerStats implements the Store interface.
func (store *PostgresStore) SaveReferrerStats(tx *sqlx.Tx, entity *ReferrerStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(ReferrerStats)
	err := tx.Get(existing, `SELECT id, visitors FROM "referrer_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND LOWER("path") = LOWER($3)
		AND LOWER("referrer") = LOWER($4)`, entity.TenantID, entity.Day, entity.Path, entity.Referrer)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "referrer_stats" ("tenant_id", "day", "path", "referrer", "visitors") VALUES (:tenant_id, :day, :path, :referrer, :visitors)`,
		`UPDATE "referrer_stats" SET visitors = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// SaveOSStats implements the Store interface.
func (store *PostgresStore) SaveOSStats(tx *sqlx.Tx, entity *OSStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(OSStats)
	err := tx.Get(existing, `SELECT id, visitors FROM "os_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND LOWER("path") = LOWER($3)
		AND "os" = $4
		AND "os_version" = $5`, entity.TenantID, entity.Day, entity.Path, entity.OS, entity.OSVersion)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "os_stats" ("tenant_id", "day", "path", "os", "os_version", "visitors") VALUES (:tenant_id, :day, :path, :os, :os_version, :visitors)`,
		`UPDATE "os_stats" SET visitors = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// SaveBrowserStats implements the Store interface.
func (store *PostgresStore) SaveBrowserStats(tx *sqlx.Tx, entity *BrowserStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(BrowserStats)
	err := tx.Get(existing, `SELECT id, visitors FROM "browser_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND LOWER("path") = LOWER($3)
		AND "browser" = $4
		AND "browser_version" = $5`, entity.TenantID, entity.Day, entity.Path, entity.Browser, entity.BrowserVersion)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "browser_stats" ("tenant_id", "day", "path", "browser", "browser_version", "visitors") VALUES (:tenant_id, :day, :path, :browser, :browser_version, :visitors)`,
		`UPDATE "browser_stats" SET visitors = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// HitDays implements the Store interface.
func (store *PostgresStore) HitDays(tenantID sql.NullInt64) ([]time.Time, error) {
	query := `SELECT DISTINCT date("time") AS "day"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND date("time") < current_date
		ORDER BY "day" ASC`
	var days []time.Time

	if err := store.DB.Select(&days, query, tenantID); err != nil {
		return nil, err
	}

	return days, nil
}

// HitPaths implements the Store interface.
func (store *PostgresStore) HitPaths(tenantID sql.NullInt64, day time.Time) ([]string, error) {
	query := `SELECT DISTINCT "path" FROM "hit" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND date("time") = $2 ORDER BY "path" ASC`
	var paths []string

	if err := store.DB.Select(&paths, query, tenantID, day); err != nil {
		return nil, err
	}

	return paths, nil
}

// Paths implements the Store interface.
func (store *PostgresStore) Paths(tenantID sql.NullInt64, from, to time.Time) ([]string, error) {
	query := `SELECT DISTINCT "path" FROM (
			SELECT "path"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND date("time") >= $2::date
			AND date("time") <= $3::date
			UNION
			SELECT "path"
			FROM "visitor_stats"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "day" >= $2::date
			AND "day" <= $3::date
		) AS results`
	var paths []string

	if err := store.DB.Select(&paths, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return paths, nil
}

func (store *PostgresStore) CountVisitors(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) *Stats {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT date("time") "day", count(DISTINCT fingerprint) "visitors"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND date("time") = $2::date
		GROUP BY "day"`
	visitors := new(Stats)

	if err := tx.Get(visitors, query, tenantID, day); err != nil {
		return nil
	}

	return visitors
}

// CountVisitorsByPath implements the Store interface.
func (store *PostgresStore) CountVisitorsByPath(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string, includePlatform bool) ([]VisitorStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT * FROM (SELECT "tenant_id", $2::date "day", $3::varchar "path", count(DISTINCT fingerprint) "visitors" `

	if includePlatform {
		query += `, (
				SELECT COUNT(1) FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "time" >= $2::date
				AND "time" < $2::date + INTERVAL '1 day'
				AND LOWER("path") = LOWER($3)
				AND desktop IS TRUE
				AND mobile IS FALSE
			) AS "platform_desktop",
			(
				SELECT COUNT(1) FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "time" >= $2::date
				AND "time" < $2::date + INTERVAL '1 day'
				AND LOWER("path") = LOWER($3)
				AND desktop IS FALSE
				AND mobile IS TRUE
			) AS "platform_mobile",
			(
				SELECT COUNT(1) FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "time" >= $2::date
				AND "time" < $2::date + INTERVAL '1 day'
				AND LOWER("path") = LOWER($3)
				AND desktop IS FALSE
				AND mobile IS FALSE
			) AS "platform_unknown" `
	}

	query += `FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= $2::date
			AND "time" < $2::date + INTERVAL '1 day'
			AND LOWER("path") = LOWER($3)
			GROUP BY tenant_id, "path"
		) AS results ORDER BY "day" ASC`
	var visitors []VisitorStats

	if err := tx.Select(&visitors, query, tenantID, day, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByPathAndHour implements the Store interface.
func (store *PostgresStore) CountVisitorsByPathAndHour(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) ([]VisitorTimeStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT $1::bigint AS "tenant_id",
		$2::date AS "day",
		$3::varchar AS "path",
		EXTRACT(HOUR FROM "day_and_hour") "hour",
		(
			SELECT count(DISTINCT fingerprint) FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= "day_and_hour"
			AND "time" < "day_and_hour" + INTERVAL '1 hour'
			AND LOWER("path") = LOWER($3)
		) "visitors"
		FROM (
			SELECT * FROM generate_series(
				$2::timestamp,
				$2::timestamp + INTERVAL '23 hours',
				interval '1 hour'
			) "day_and_hour"
		) AS hours`
	var visitors []VisitorTimeStats

	if err := tx.Select(&visitors, query, tenantID, day, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByPathAndLanguage implements the Store interface.
func (store *PostgresStore) CountVisitorsByPathAndLanguage(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) ([]LanguageStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT * FROM (
			SELECT "tenant_id", $2::date "day", $3::varchar "path", "language", count(DISTINCT fingerprint) "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= $2::date
			AND "time" < $2::date + INTERVAL '1 day'
			AND LOWER("path") = LOWER($3)
			GROUP BY tenant_id, "language"
		) AS results ORDER BY "day" ASC`
	var visitors []LanguageStats

	if err := tx.Select(&visitors, query, tenantID, day, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByPathAndReferrer implements the Store interface.
func (store *PostgresStore) CountVisitorsByPathAndReferrer(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) ([]ReferrerStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT * FROM (
			SELECT "tenant_id", $2::date "day", $3::varchar "path", "referrer", count(DISTINCT fingerprint) "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= $2::date
			AND "time" < $2::date + INTERVAL '1 day'
			AND LOWER("path") = LOWER($3)
			GROUP BY tenant_id, "referrer"
		) AS results ORDER BY "day" ASC`
	var visitors []ReferrerStats

	if err := tx.Select(&visitors, query, tenantID, day, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByPathAndOS implements the Store interface.
func (store *PostgresStore) CountVisitorsByPathAndOS(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) ([]OSStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT * FROM (
			SELECT "tenant_id", $2::date "day", $3::varchar "path", "os", "os_version", count(DISTINCT fingerprint) "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= $2::date
			AND "time" < $2::date + INTERVAL '1 day'
			AND LOWER("path") = LOWER($3)
			GROUP BY tenant_id, "os", "os_version"
		) AS results ORDER BY "day" ASC`
	var visitors []OSStats

	if err := tx.Select(&visitors, query, tenantID, day, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByPathAndBrowser implements the Store interface.
func (store *PostgresStore) CountVisitorsByPathAndBrowser(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) ([]BrowserStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT * FROM (
			SELECT "tenant_id", $2::date "day", $3::varchar "path", "browser", "browser_version", count(DISTINCT fingerprint) "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= $2::date
			AND "time" < $2::date + INTERVAL '1 day'
			AND LOWER("path") = LOWER($3)
			GROUP BY tenant_id, "browser", "browser_version"
		) AS results ORDER BY "day" ASC`
	var visitors []BrowserStats

	if err := tx.Select(&visitors, query, tenantID, day, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// ActiveVisitors implements the Store interface.
func (store *PostgresStore) ActiveVisitors(tenantID sql.NullInt64, path string, from time.Time) ([]Stats, error) {
	args := make([]interface{}, 0, 3)
	args = append(args, tenantID)
	args = append(args, from)
	query := `SELECT * FROM (
			SELECT "tenant_id", "path", count(DISTINCT fingerprint) "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" > $2 `

	if path != "" {
		args = append(args, path)
		query += `AND LOWER("path") = LOWER($3) `
	}

	query += `GROUP BY tenant_id, "path") AS results ORDER BY "visitors" DESC, "path" ASC`
	var visitors []Stats

	if err := store.DB.Select(&visitors, query, args...); err != nil {
		return nil, err
	}

	return visitors, nil
}

// Visitors implements the Store interface.
func (store *PostgresStore) Visitors(tenantID sql.NullInt64, from, to time.Time) ([]Stats, error) {
	query := `SELECT "d" AS "day",
		COALESCE(SUM("visitor_stats".visitors), 0) "visitors"
		FROM (
			SELECT * FROM generate_series(
				$2::date,
				$3::date,
				INTERVAL '1 day'
			) "d"
		) AS date_series
		LEFT JOIN "visitor_stats" ON ($1::bigint IS NULL OR tenant_id = $1) AND "visitor_stats"."day" = "d"
		GROUP BY "d"
		ORDER BY "d" ASC`
	var visitors []Stats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// PageVisitors implements the Store interface.
func (store *PostgresStore) PageVisitors(tenantID sql.NullInt64, path string, from, to time.Time) ([]Stats, error) {
	query := `SELECT "d" AS "day",
		CASE WHEN "path" IS NULL THEN '' ELSE "path" END,
		CASE WHEN "visitor_stats".visitors IS NULL THEN 0 ELSE "visitor_stats".visitors END
		FROM (
			SELECT * FROM generate_series(
				$2::date,
				$3::date,
				INTERVAL '1 day'
			) "d"
		) AS date_series
		LEFT JOIN "visitor_stats" ON ($1::bigint IS NULL OR tenant_id = $1) AND "visitor_stats"."day" = "d" AND LOWER("path") = LOWER($4)
		ORDER BY "d" ASC`
	var visitors []Stats

	if err := store.DB.Select(&visitors, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// Referrer implements the Store interface.
func (store *PostgresStore) Referrer(tenantID sql.NullInt64, path string, from, to time.Time) ([]ReferrerStats, error) {
	query := `SELECT "d" AS "day",
		CASE WHEN "path" IS NULL THEN '' ELSE "path" END,
		"referrer_stats"."referrer",
		CASE WHEN "referrer_stats".visitors IS NULL THEN 0 ELSE "referrer_stats".visitors END
		FROM (
			SELECT * FROM generate_series(
				$2::date,
				$3::date,
				INTERVAL '1 day'
			) "d"
		) AS date_series
		LEFT JOIN "referrer_stats" ON ($1::bigint IS NULL OR tenant_id = $1) AND "referrer_stats"."day" = "d" AND LOWER("path") = LOWER($4)
		ORDER BY "d" ASC`
	var visitors []ReferrerStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

func (store *PostgresStore) createUpdateEntity(tx *sqlx.Tx, entity, existing StatsEntity, found bool, insertQuery, updateQuery string) error {
	if found {
		visitors := existing.GetVisitors() + entity.GetVisitors()

		if _, err := tx.Exec(updateQuery, visitors, existing.GetID()); err != nil {
			return err
		}
	} else {
		rows, err := tx.NamedQuery(insertQuery, entity)

		if err != nil {
			return err
		}

		store.closeRows(rows)
	}

	return nil
}

func (store *PostgresStore) closeRows(rows *sqlx.Rows) {
	if err := rows.Close(); err != nil {
		store.logger.Printf("error closing rows: %s", err)
	}
}

/*
// Referrer implements the Store interface.
func (store *PostgresStore) Referrer(tenantID sql.NullInt64, from, to time.Time) ([]string, error) {
	query := `SELECT * FROM (
			SELECT DISTINCT "ref" FROM "visitors_per_referrer" WHERE ($1::bigint IS NULL OR tenant_id = $1) AND "day" >= $2 AND "day" <= $3
		) AS referrer ORDER BY length("ref"), "ref" ASC`
	var referrer []string

	if err := store.DB.Select(&referrer, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return referrer, nil
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

// ReferrerVisits implements the Store interface.
func (store *PostgresStore) ReferrerVisits(tenantID sql.NullInt64, referrer string, from, to time.Time) ([]VisitorsPerReferrer, error) {
	query := `SELECT tenant_id, "date" "day",
		CASE WHEN "visitors_per_referrer".visitors IS NULL THEN 0 ELSE "visitors_per_referrer".visitors END
		FROM (
			SELECT * FROM generate_series(
				$2::timestamp,
				$3::timestamp,
				INTERVAL '1 day'
			) "date"
		) AS date_series
		LEFT JOIN "visitors_per_referrer" ON ($1::bigint IS NULL OR tenant_id = $1) AND date("visitors_per_referrer"."day") = date("date") AND "visitors_per_referrer"."ref" = $4
		ORDER BY "date" ASC`
	var visitors []VisitorsPerReferrer

	if err := store.DB.Select(&visitors, query, tenantID, from, to, referrer); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorPages implements the Store interface.
func (store *PostgresStore) VisitorPages(tenantID sql.NullInt64, from time.Time, to time.Time) ([]Stats, error) {
	query := `SELECT * FROM (
			SELECT "path", sum("visitors") "visitors" FROM (
				SELECT "path", sum("visitors") "visitors" FROM "visitors_per_page"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= date($2::timestamp)
				AND "day" <= date($3::timestamp)
				GROUP BY "path"
				UNION
				SELECT "path", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				GROUP BY "path"
			) AS results
			GROUP BY "path"
		) AS pages
		ORDER BY "visitors" DESC`
	var pages []Stats

	if err := store.DB.Select(&pages, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return pages, nil
}

// VisitorLanguages implements the Store interface.
func (store *PostgresStore) VisitorLanguages(tenantID sql.NullInt64, from, to time.Time) ([]Stats, error) {
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
	var languages []Stats

	if err := store.DB.Select(&languages, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return languages, nil
}

// VisitorReferrer implements the Store interface.
func (store *PostgresStore) VisitorReferrer(tenantID sql.NullInt64, from time.Time, to time.Time) ([]Stats, error) {
	query := `SELECT * FROM (
			SELECT "ref", sum("visitors") "visitors" FROM (
				SELECT "ref", sum("visitors") "visitors" FROM "visitors_per_referrer"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= date($2::timestamp)
				AND "day" <= date($3::timestamp)
				GROUP BY "ref"
				UNION
				SELECT "ref", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				GROUP BY "ref"
			) AS results
			GROUP BY "ref"
		) AS referrer
		ORDER BY "visitors" DESC`
	var referrer []Stats

	if err := store.DB.Select(&referrer, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return referrer, nil
}

// VisitorOS implements the Store interface.
func (store *PostgresStore) VisitorOS(tenantID sql.NullInt64, from time.Time, to time.Time) ([]Stats, error) {
	query := `SELECT * FROM (
			SELECT "os", sum("visitors") "visitors" FROM (
				SELECT "os", sum("visitors") "visitors" FROM "visitors_per_os"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= date($2::timestamp)
				AND "day" <= date($3::timestamp)
				GROUP BY "os"
				UNION
				SELECT "os", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				GROUP BY "os"
			) AS results
			GROUP BY "os"
		) AS operating_systems
		ORDER BY "visitors" DESC`
	var os []Stats

	if err := store.DB.Select(&os, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return os, nil
}

// VisitorBrowser implements the Store interface.
func (store *PostgresStore) VisitorBrowser(tenantID sql.NullInt64, from time.Time, to time.Time) ([]Stats, error) {
	query := `SELECT * FROM (
			SELECT "browser", sum("visitors") "visitors" FROM (
				SELECT "browser", sum("visitors") "visitors" FROM "visitors_per_browser"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= date($2::timestamp)
				AND "day" <= date($3::timestamp)
				GROUP BY "browser"
				UNION
				SELECT "browser", count(DISTINCT fingerprint) "visitors" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				GROUP BY "browser"
			) AS results
			GROUP BY "browser"
		) AS browser
		ORDER BY "visitors" DESC`
	var browser []Stats

	if err := store.DB.Select(&browser, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return browser, nil
}

// VisitorPlatform implements the Store interface.
func (store *PostgresStore) VisitorPlatform(tenantID sql.NullInt64, from time.Time, to time.Time) (*Stats, error) {
	query := `SELECT sum("desktop") "platform_desktop_visitors",
				sum("mobile") "platform_mobile_visitors",
				sum("unknown") "platform_unknown_visitors" FROM (
				SELECT "desktop", "mobile", "unknown" FROM "visitor_platform"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= date($2::timestamp)
				AND "day" <= date($3::timestamp)
				UNION
				SELECT count(DISTINCT fingerprint) "desktop", 0 "mobile", 0 "unknown" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				AND "desktop" IS TRUE
				AND "mobile" IS FALSE
				UNION
				SELECT 0 "desktop", count(DISTINCT fingerprint) "mobile", 0 "unknown" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				AND "desktop" IS FALSE
				AND "mobile" IS TRUE
				UNION
				SELECT 0 "desktop", 0 "mobile", count(DISTINCT fingerprint) "unknown" FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND date("time") >= date($2::timestamp)
				AND date("time") <= date($3::timestamp)
				AND "desktop" IS FALSE
				AND "mobile" IS FALSE
			) AS results`
	platforms := new(Stats)

	if err := store.DB.Get(platforms, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return platforms, nil
}

// HourlyVisitors implements the Store interface.
func (store *PostgresStore) HourlyVisitors(tenantID sql.NullInt64, from, to time.Time) ([]Stats, error) {
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
	var visitors []Stats

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
func (store *PostgresStore) ActiveVisitorsPerPage(tenantID sql.NullInt64, from time.Time) ([]Stats, error) {
	query := `SELECT "path", count(DISTINCT fingerprint) AS "visitors"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "time" > $2
		GROUP BY "path"
		ORDER BY "visitors" DESC`
	var visitors []Stats

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

// VisitorsPerReferrer implements the Store interface.
func (store *PostgresStore) VisitorsPerReferrer(tenantID sql.NullInt64) []VisitorsPerReferrer {
	var entities []VisitorsPerReferrer

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_referrer" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day" ASC, "visitors" DESC`, tenantID); err != nil {
		return nil
	}

	return entities
}

// VisitorsPerOS implements the Store interface.
func (store *PostgresStore) VisitorsPerOS(tenantID sql.NullInt64) []VisitorsPerOS {
	var entities []VisitorsPerOS

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_os" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day" ASC, "visitors" DESC`, tenantID); err != nil {
		return nil
	}

	return entities
}

// VisitorsPerBrowser implements the Store interface.
func (store *PostgresStore) VisitorsPerBrowser(tenantID sql.NullInt64) []VisitorsPerBrowser {
	var entities []VisitorsPerBrowser

	if err := store.DB.Select(&entities, `SELECT * FROM "visitors_per_browser" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day" ASC, "visitors" DESC`, tenantID); err != nil {
		return nil
	}

	return entities
}

// VisitorsPerPlatform implements the Store interface.
func (store *PostgresStore) VisitorsPerPlatform(tenantID sql.NullInt64) []VisitorPlatform {
	var entities []VisitorPlatform

	if err := store.DB.Select(&entities, `SELECT * FROM "visitor_platform" WHERE ($1::bigint IS NULL OR tenant_id = $1) ORDER BY "day" ASC`, tenantID); err != nil {
		return nil
	}

	return entities
}*/
