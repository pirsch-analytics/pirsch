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
	logPrefix = "[pirsch] "
)

// statsEntity is an interface for all statistics entities.
// This is used to simplify saving entities in the database.
type statsEntity interface {
	// GetID returns the ID.
	GetID() int64

	// GetVisitors returns the visitor count.
	GetVisitors() int
}

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
			Logger: logger,
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

// SaveHits implements the Store interface.
func (store *PostgresStore) SaveHits(hits []Hit) error {
	const hitParams = 21
	args := make([]interface{}, 0, len(hits)*hitParams)
	var query strings.Builder
	query.WriteString(`INSERT INTO "hit" (tenant_id, fingerprint, session, path, url, language, user_agent, referrer, referrer_name, referrer_icon, os, os_version, browser, browser_version, country_code, desktop, mobile, screen_width, screen_height, screen_class, time) VALUES `)

	for i, hit := range hits {
		args = append(args, hit.TenantID)
		args = append(args, hit.Fingerprint)
		args = append(args, hit.Session)
		args = append(args, hit.Path)
		args = append(args, hit.URL)
		args = append(args, hit.Language)
		args = append(args, hit.UserAgent)
		args = append(args, hit.Referrer)
		args = append(args, hit.ReferrerName)
		args = append(args, hit.ReferrerIcon)
		args = append(args, hit.OS)
		args = append(args, hit.OSVersion)
		args = append(args, hit.Browser)
		args = append(args, hit.BrowserVersion)
		args = append(args, hit.CountryCode)
		args = append(args, hit.Desktop)
		args = append(args, hit.Mobile)
		args = append(args, hit.ScreenWidth)
		args = append(args, hit.ScreenHeight)
		args = append(args, hit.ScreenClass)
		args = append(args, hit.Time)
		index := i * hitParams
		query.WriteString(fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d),`,
			index+1, index+2, index+3, index+4, index+5, index+6, index+7, index+8, index+9, index+10, index+11, index+12, index+13, index+14, index+15, index+16, index+17, index+18, index+19, index+20, index+21))
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
	err := tx.Get(existing, `SELECT id, visitors, sessions, bounces, views, platform_desktop, platform_mobile, platform_unknown, average_session_duration_seconds
		FROM "visitor_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND (LOWER("path") = LOWER($3) OR $3 IS NULL AND "path" IS NULL)`, entity.TenantID, entity.Day, entity.Path)

	if err == nil {
		existing.Visitors += entity.Visitors
		existing.Sessions += entity.Sessions
		existing.Bounces += entity.Bounces
		existing.Views += entity.Views
		existing.PlatformDesktop += entity.PlatformDesktop
		existing.PlatformMobile += entity.PlatformMobile
		existing.PlatformUnknown += entity.PlatformUnknown
		existing.AverageSessionDurationSeconds = addAverage(existing.AverageSessionDurationSeconds, entity.AverageSessionDurationSeconds, existing.Sessions)

		if _, err := tx.Exec(`UPDATE "visitor_stats" SET "visitors" = $1, "sessions" = $2, "bounces" = $3, "views" = $4, "platform_desktop" = $5, "platform_mobile" = $6, "platform_unknown" = $7, "average_session_duration_seconds" = $8 WHERE id = $9`,
			existing.Visitors,
			existing.Sessions,
			existing.Bounces,
			existing.Views,
			existing.PlatformDesktop,
			existing.PlatformMobile,
			existing.PlatformUnknown,
			existing.AverageSessionDurationSeconds,
			existing.ID); err != nil {
			return err
		}
	} else {
		rows, err := tx.NamedQuery(`INSERT INTO "visitor_stats" ("tenant_id", "day", "path", "visitors", "sessions", "bounces", "views", "platform_desktop", "platform_mobile", "platform_unknown", "average_session_duration_seconds") VALUES (:tenant_id, :day, :path, :visitors, :sessions, :bounces, :views, :platform_desktop, :platform_mobile, :platform_unknown, :average_session_duration_seconds)`, entity)

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
		AND "hour" = $3`, entity.TenantID, entity.Day, entity.Hour)

	if err == nil {
		existing.Visitors += entity.Visitors

		if _, err := tx.Exec(`UPDATE "visitor_time_stats" SET "visitors" = $1 WHERE id = $2`,
			existing.Visitors,
			existing.ID); err != nil {
			return err
		}
	} else {
		rows, err := tx.NamedQuery(`INSERT INTO "visitor_time_stats" ("tenant_id", "day", "hour", "visitors") VALUES (:tenant_id, :day, :hour, :visitors)`, entity)

		if err != nil {
			return err
		}

		store.closeRows(rows)
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
		AND (LOWER("path") = LOWER($3) OR $3 IS NULL AND "path" IS NULL)
		AND "language" = LOWER($4)`, entity.TenantID, entity.Day, entity.Path, entity.Language)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "language_stats" ("tenant_id", "day", "path", "language", "visitors") VALUES (:tenant_id, :day, :path, LOWER(:language), :visitors)`,
		`UPDATE "language_stats" SET "visitors" = $1 WHERE id = $2`); err != nil {
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
	err := tx.Get(existing, `SELECT id, visitors, bounces FROM "referrer_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND (LOWER("path") = LOWER($3) OR $3 IS NULL AND "path" IS NULL)
		AND LOWER("referrer") = LOWER($4)`, entity.TenantID, entity.Day, entity.Path, entity.Referrer)

	if err == nil {
		existing.Visitors += entity.Visitors
		existing.Bounces += entity.Bounces

		if _, err := tx.Exec(`UPDATE "referrer_stats" SET "visitors" = $1, "bounces" = $2 WHERE id = $3`,
			existing.Visitors,
			existing.Bounces,
			existing.ID); err != nil {
			return err
		}
	} else {
		rows, err := tx.NamedQuery(`INSERT INTO "referrer_stats" ("tenant_id", "day", "path", "referrer", "referrer_name", "referrer_icon", "visitors", "bounces") VALUES (:tenant_id, :day, :path, :referrer, :referrer_name, :referrer_icon, :visitors, :bounces)`, entity)

		if err != nil {
			return err
		}

		store.closeRows(rows)
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
		AND (LOWER("path") = LOWER($3) OR $3 IS NULL AND "path" IS NULL)
		AND "os" = $4
		AND "os_version" = $5`, entity.TenantID, entity.Day, entity.Path, entity.OS, entity.OSVersion)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "os_stats" ("tenant_id", "day", "path", "os", "os_version", "visitors") VALUES (:tenant_id, :day, :path, :os, :os_version, :visitors)`,
		`UPDATE "os_stats" SET "visitors" = $1 WHERE id = $2`); err != nil {
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
		AND (LOWER("path") = LOWER($3) OR $3 IS NULL AND "path" IS NULL)
		AND "browser" = $4
		AND "browser_version" = $5`, entity.TenantID, entity.Day, entity.Path, entity.Browser, entity.BrowserVersion)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "browser_stats" ("tenant_id", "day", "path", "browser", "browser_version", "visitors") VALUES (:tenant_id, :day, :path, :browser, :browser_version, :visitors)`,
		`UPDATE "browser_stats" SET "visitors" = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// SaveScreenStats implements the Store interface.
func (store *PostgresStore) SaveScreenStats(tx *sqlx.Tx, entity *ScreenStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(ScreenStats)
	err := tx.Get(existing, `SELECT id, visitors
		FROM "screen_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND "width" = $3
		AND "height" = $4
		AND ("class" = $5 OR $5 IS NULL AND "class" IS NULL)`, entity.TenantID, entity.Day, entity.Width, entity.Height, entity.Class)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "screen_stats" ("tenant_id", "day", "width", "height", "class", "visitors") VALUES (:tenant_id, :day, :width, :height, :class, :visitors)`,
		`UPDATE "screen_stats" SET "visitors" = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// SaveCountryStats implements the Store interface.
func (store *PostgresStore) SaveCountryStats(tx *sqlx.Tx, entity *CountryStats) error {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	existing := new(CountryStats)
	err := tx.Get(existing, `SELECT id, visitors FROM "country_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" = $2
		AND "country_code" = $3`, entity.TenantID, entity.Day, entity.CountryCode)

	if err := store.createUpdateEntity(tx, entity, existing, err == nil,
		`INSERT INTO "country_stats" ("tenant_id", "day", "country_code", "visitors") VALUES (:tenant_id, :day, LOWER(:country_code), :visitors)`,
		`UPDATE "country_stats" SET "visitors" = $1 WHERE id = $2`); err != nil {
		return err
	}

	return nil
}

// Session implements the Store interface.
func (store *PostgresStore) Session(tenantID sql.NullInt64, fingerprint string, maxAge time.Time) time.Time {
	query := `SELECT "session"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND fingerprint = $2
	    AND "time" > $3
		LIMIT 1`
	var session time.Time

	if err := store.DB.Get(&session, query, tenantID, fingerprint, maxAge); err != nil && err != sql.ErrNoRows {
		store.logger.Printf("error reading session timestamp: %s", err)
	}

	return session
}

// HitDays implements the Store interface.
func (store *PostgresStore) HitDays(tenantID sql.NullInt64) ([]time.Time, error) {
	query := `SELECT DISTINCT DATE("time") AS "day"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") < current_date AT TIME ZONE 'UTC'
		ORDER BY "day" ASC`
	var days []time.Time

	if err := store.DB.Select(&days, query, tenantID); err != nil {
		return nil, err
	}

	return days, nil
}

// HitPaths implements the Store interface.
func (store *PostgresStore) HitPaths(tenantID sql.NullInt64, day time.Time) ([]string, error) {
	query := `SELECT DISTINCT "path"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2
		ORDER BY "path" ASC`
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
			AND DATE("time") >= $2::date
			AND DATE("time") <= $3::date
			UNION
			SELECT "path"
			FROM "visitor_stats"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "day" >= $2::date
			AND "day" <= $3::date
    		AND "path" IS NOT NULL
		) AS results
		ORDER BY "path" ASC`
	var paths []string

	if err := store.DB.Select(&paths, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return paths, nil
}

// CountVisitors implements the Store interface.
func (store *PostgresStore) CountVisitors(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) *Stats {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT DATE("time") "day",
        count(DISTINCT "fingerprint") "visitors",
        count(DISTINCT("fingerprint", "session")) "sessions",
        count(1) "views"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		GROUP BY "day"`
	visitors := new(Stats)

	if err := tx.Get(visitors, query, tenantID, day); err != nil && err != sql.ErrNoRows {
		store.logger.Printf("error counting visitors: %s", err)
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

	query := `SELECT * FROM (
    	SELECT "tenant_id",
		$2::date "day",
	    $3::varchar "path",
	    count(DISTINCT "fingerprint") "visitors",
		count(DISTINCT("fingerprint", "session")) "sessions",
		count(1) "views" `

	if includePlatform {
		query += `, (
				SELECT COUNT(DISTINCT "fingerprint") FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "time" >= $2::date
				AND "time" < $2::date + INTERVAL '1 day'
				AND LOWER("path") = LOWER($3)
				AND desktop IS TRUE
				AND mobile IS FALSE
			) AS "platform_desktop",
			(
				SELECT COUNT(DISTINCT "fingerprint") FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "time" >= $2::date
				AND "time" < $2::date + INTERVAL '1 day'
				AND LOWER("path") = LOWER($3)
				AND desktop IS FALSE
				AND mobile IS TRUE
			) AS "platform_mobile",
			(
				SELECT COUNT(DISTINCT "fingerprint") FROM "hit"
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

// CountVisitorsByHour implements the Store interface.
func (store *PostgresStore) CountVisitorsByHour(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]VisitorTimeStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT $1::bigint AS "tenant_id",
		$2::date AS "day",
		EXTRACT(HOUR FROM "day_and_hour") "hour",
		(
			SELECT count(DISTINCT "fingerprint") FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= "day_and_hour"
			AND "time" < "day_and_hour" + INTERVAL '1 hour'
		) "visitors",
       (
			SELECT count(DISTINCT("fingerprint", "session")) FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= "day_and_hour"
			AND "time" < "day_and_hour" + INTERVAL '1 hour'
		) "sessions"
		FROM (
			SELECT * FROM generate_series(
				$2::timestamp,
				$2::timestamp + INTERVAL '23 hours',
				interval '1 hour'
			) "day_and_hour"
		) AS hours`
	var visitors []VisitorTimeStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
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
		) AS results
		ORDER BY "day" ASC`
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
			SELECT "tenant_id", $2::date "day", $3::varchar "path", "referrer", "referrer_name", "referrer_icon", count(DISTINCT fingerprint) "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" >= $2::date
			AND "time" < $2::date + INTERVAL '1 day'
			AND LOWER("path") = LOWER($3)
			GROUP BY tenant_id, "referrer", "referrer_name", "referrer_icon"
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

// CountVisitorsByLanguage implements the Store interface.
func (store *PostgresStore) CountVisitorsByLanguage(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]LanguageStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT "language", count(DISTINCT fingerprint) "visitors", $2::date "day"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		GROUP BY "language"`
	var visitors []LanguageStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByReferrer implements the Store interface.
func (store *PostgresStore) CountVisitorsByReferrer(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]ReferrerStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT h1."referrer",
       	h1."referrer_name",
        h1."referrer_icon",
        count(DISTINCT h1.fingerprint) "visitors",
	    (
			SELECT count(DISTINCT h2."fingerprint")
			FROM "hit" h2
			WHERE ($1::bigint IS NULL OR h2.tenant_id = $1)
			AND DATE(h2."time") = $2::date
			AND h2."referrer" = h1."referrer"
			AND (
				SELECT COUNT(DISTINCT h3."path")
				FROM "hit" h3
				WHERE h3."fingerprint" = h2."fingerprint"
			) = 1
	    ) "bounces",
        $2::date "day"
		FROM "hit" h1
		WHERE ($1::bigint IS NULL OR h1.tenant_id = $1)
		AND DATE(h1."time") = $2::date
		GROUP BY h1."referrer", h1."referrer_name", h1."referrer_icon"`
	var visitors []ReferrerStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByOS implements the Store interface.
func (store *PostgresStore) CountVisitorsByOS(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]OSStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT "os", count(DISTINCT fingerprint) "visitors", $2::date "day"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		GROUP BY "os"`
	var visitors []OSStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByBrowser implements the Store interface.
func (store *PostgresStore) CountVisitorsByBrowser(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]BrowserStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT "browser", count(DISTINCT fingerprint) "visitors", $2::date "day"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		GROUP BY "browser"`
	var visitors []BrowserStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByScreenSize implements the Store interface.
func (store *PostgresStore) CountVisitorsByScreenSize(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]ScreenStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT "tenant_id", $2::date "day", "screen_width" "width", "screen_height" "height", count(DISTINCT fingerprint) "visitors"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		AND "screen_width" != 0
		AND "screen_height" != 0
		GROUP BY "tenant_id", "width", "height"`
	var visitors []ScreenStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByScreenClass implements the Store interface.
func (store *PostgresStore) CountVisitorsByScreenClass(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]ScreenStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT "tenant_id", $2::date "day", "screen_class" "class", count(DISTINCT fingerprint) "visitors"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		AND "screen_class" IS NOT NULL
		GROUP BY "tenant_id", "screen_class"`
	var visitors []ScreenStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByCountryCode implements the Store interface.
func (store *PostgresStore) CountVisitorsByCountryCode(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) ([]CountryStats, error) {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT "tenant_id", $2::date "day", "country_code", count(DISTINCT fingerprint) "visitors"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		GROUP BY "tenant_id", "country_code"`
	var visitors []CountryStats

	if err := tx.Select(&visitors, query, tenantID, day); err != nil {
		return nil, err
	}

	return visitors, nil
}

// CountVisitorsByPlatform implements the Store interface.
func (store *PostgresStore) CountVisitorsByPlatform(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) *VisitorStats {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT (
				SELECT COUNT(DISTINCT "fingerprint")
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") = $2::date
				AND desktop IS TRUE
				AND mobile IS FALSE
			) AS "platform_desktop",
			(
				SELECT COUNT(DISTINCT "fingerprint")
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") = $2::date
				AND desktop IS FALSE
				AND mobile IS TRUE
			) AS "platform_mobile",
			(
				SELECT COUNT(DISTINCT "fingerprint")
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") = $2::date
				AND desktop IS FALSE
				AND mobile IS FALSE
			) AS "platform_unknown"`
	visitors := new(VisitorStats)

	if err := tx.Get(visitors, query, tenantID, day); err != nil && err != sql.ErrNoRows {
		store.logger.Printf("error counting visitor platforms: %s", err)
		return nil
	}

	return visitors
}

// CountVisitorsByPathAndMaxOneHit implements the Store interface.
func (store *PostgresStore) CountVisitorsByPathAndMaxOneHit(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) int {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	args := make([]interface{}, 0, 3)
	args = append(args, tenantID)
	args = append(args, day)
	query := `SELECT count(DISTINCT "fingerprint")
		FROM "hit" h
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date `

	if path != "" {
		args = append(args, path)
		query += `AND LOWER("path") = LOWER($3) `
	}

	query += `AND (
			SELECT COUNT(DISTINCT "path")
			FROM "hit"
			WHERE "fingerprint" = h."fingerprint"
		) = 1`
	var visitors int

	if err := tx.Get(&visitors, query, args...); err != nil {
		store.logger.Printf("error counting visitors for path with a maximum of one hit: %s", err)
	}

	return visitors
}

// CountVisitorsByPathAndReferrerAndMaxOneHit implements the Store interface.
func (store *PostgresStore) CountVisitorsByPathAndReferrerAndMaxOneHit(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path, referrer string) int {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	args := make([]interface{}, 0, 4)
	args = append(args, tenantID)
	args = append(args, day)
	args = append(args, referrer)
	query := `SELECT count(DISTINCT "fingerprint")
		FROM "hit" h
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND DATE("time") = $2::date
		AND LOWER("referrer") = LOWER($3) `

	if path != "" {
		args = append(args, path)
		query += `AND LOWER("path") = LOWER($4) `
	}

	query += `AND (
			SELECT COUNT(DISTINCT "path")
			FROM "hit"
			WHERE "fingerprint" = h."fingerprint"
		) = 1`
	var visitors int

	if err := tx.Get(&visitors, query, args...); err != nil {
		store.logger.Printf("error counting visitors for referrer with a maximum of one hit: %s", err)
	}

	return visitors
}

// ActiveVisitors implements the Store interface.
func (store *PostgresStore) ActiveVisitors(tenantID sql.NullInt64, from time.Time) int {
	query := `SELECT count(DISTINCT fingerprint) "visitors"
		FROM "hit"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "time" > $2`
	visitors := 0

	if err := store.DB.Get(&visitors, query, tenantID, from); err != nil {
		store.logger.Printf("error counting active visitors: %s", err)
		return 0
	}

	return visitors
}

// ActivePageVisitors implements the Store interface.
func (store *PostgresStore) ActivePageVisitors(tenantID sql.NullInt64, from time.Time) ([]Stats, error) {
	query := `SELECT * FROM (
			SELECT "tenant_id", "path", count(DISTINCT fingerprint) "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "time" > $2
			GROUP BY tenant_id, "path"
		) AS results
		ORDER BY "visitors" DESC, "path" ASC`
	var visitors []Stats

	if err := store.DB.Select(&visitors, query, tenantID, from); err != nil {
		return nil, err
	}

	return visitors, nil
}

// Visitors implements the Store interface.
func (store *PostgresStore) Visitors(tenantID sql.NullInt64, from, to time.Time) ([]Stats, error) {
	// summing up the average session duration isn't an issue here, since there should be only one row per day
	query := `SELECT "d" AS "day",
		COALESCE(SUM("visitor_stats".visitors), 0) "visitors",
        COALESCE(SUM("visitor_stats".sessions), 0) "sessions",
        COALESCE(SUM("visitor_stats".bounces), 0) "bounces",
        COALESCE(SUM("visitor_stats".views), 0) "views",
        COALESCE(SUM("visitor_stats".average_session_duration_seconds), 0) "average_session_duration_seconds"
		FROM (
			SELECT * FROM generate_series(
				$2::date,
				$3::date,
				INTERVAL '1 day'
			) "d"
		) AS date_series
		LEFT JOIN "visitor_stats" ON ($1::bigint IS NULL OR tenant_id = $1)
		AND "visitor_stats"."day" = "d"
		AND "visitor_stats"."path" IS NULL
		GROUP BY "d"
		ORDER BY "d" ASC`
	var visitors []Stats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorHours implements the Store interface.
func (store *PostgresStore) VisitorHours(tenantID sql.NullInt64, from time.Time, to time.Time) ([]VisitorTimeStats, error) {
	query := `SELECT "day_and_hour" "hour",
        COALESCE(sum("visitors"), 0) "visitors"
		FROM generate_series(0, 23, 1) "day_and_hour"
		LEFT JOIN (
			SELECT "hour", sum("visitors") "visitors"
			FROM "visitor_time_stats"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "day" >= DATE($2::timestamp)
			AND "day" <= DATE($3::timestamp)
			GROUP BY "hour"
			UNION
			SELECT EXTRACT(HOUR FROM "time") "hour", count(DISTINCT "fingerprint") "visitors"
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND DATE("time") >= DATE($2::timestamp)
			AND DATE("time") <= DATE($3::timestamp)
			GROUP BY "hour"
		) AS results ON "hour" = "day_and_hour"
		GROUP BY "day_and_hour"
		ORDER BY "day_and_hour" ASC`
	var visitors []VisitorTimeStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorLanguages implements the Store interface.
func (store *PostgresStore) VisitorLanguages(tenantID sql.NullInt64, from, to time.Time) ([]LanguageStats, error) {
	query := `SELECT "language", COALESCE(SUM("visitors"), 0) "visitors"
		FROM "language_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		AND "path" IS NULL
		GROUP BY "language"
		ORDER BY "visitors" DESC`
	var visitors []LanguageStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorReferrer implements the Store interface.
func (store *PostgresStore) VisitorReferrer(tenantID sql.NullInt64, from, to time.Time) ([]ReferrerStats, error) {
	query := `SELECT "referrer",
       	"referrer_name",
	    "referrer_icon",
	    COALESCE(SUM("visitors"), 0) "visitors",
        COALESCE(SUM("bounces"), 0) "bounces"
		FROM "referrer_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		AND "path" IS NULL
		GROUP BY "referrer", "referrer_name", "referrer_icon"
		ORDER BY "visitors" DESC`
	var visitors []ReferrerStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorOS implements the Store interface.
func (store *PostgresStore) VisitorOS(tenantID sql.NullInt64, from, to time.Time) ([]OSStats, error) {
	query := `SELECT "os", COALESCE(SUM("visitors"), 0) "visitors"
		FROM "os_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		AND "path" IS NULL
		GROUP BY "os"
		ORDER BY "visitors" DESC`
	var visitors []OSStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorBrowser implements the Store interface.
func (store *PostgresStore) VisitorBrowser(tenantID sql.NullInt64, from, to time.Time) ([]BrowserStats, error) {
	query := `SELECT "browser", COALESCE(SUM("visitors"), 0) "visitors"
		FROM "browser_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		AND "path" IS NULL
		GROUP BY "browser"
		ORDER BY "visitors" DESC`
	var visitors []BrowserStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorPlatform implements the Store interface.
func (store *PostgresStore) VisitorPlatform(tenantID sql.NullInt64, from, to time.Time) *VisitorStats {
	query := `SELECT COALESCE(SUM("platform_desktop"), 0) "platform_desktop",
		COALESCE(SUM("platform_mobile"), 0) "platform_mobile",
		COALESCE(SUM("platform_unknown"), 0) "platform_unknown"
		FROM "visitor_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		AND "path" IS NULL`
	visitors := new(VisitorStats)

	if err := store.DB.Get(visitors, query, tenantID, from, to); err != nil && err != sql.ErrNoRows {
		store.logger.Printf("error reading visitor platforms: %s", err)
		return nil
	}

	return visitors
}

// VisitorScreenSize implements the Store interface.
func (store *PostgresStore) VisitorScreenSize(tenantID sql.NullInt64, from, to time.Time) ([]ScreenStats, error) {
	query := `SELECT "width", "height", COALESCE(SUM("visitors"), 0) "visitors"
		FROM "screen_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		AND "class" IS NULL
		GROUP BY "width", "height"
		ORDER BY "visitors" DESC`
	var visitors []ScreenStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorScreenClass implements the Store interface.
func (store *PostgresStore) VisitorScreenClass(tenantID sql.NullInt64, from, to time.Time) ([]ScreenStats, error) {
	query := `SELECT "class", COALESCE(SUM("visitors"), 0) "visitors"
		FROM "screen_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		AND "class" IS NOT NULL
		GROUP BY "class"
		ORDER BY "visitors" DESC`
	var visitors []ScreenStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// VisitorCountry implements the Store interface.
func (store *PostgresStore) VisitorCountry(tenantID sql.NullInt64, from, to time.Time) ([]CountryStats, error) {
	query := `SELECT "country_code", COALESCE(SUM("visitors"), 0) "visitors"
		FROM "country_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date
		GROUP BY "country_code"
		ORDER BY "visitors" DESC`
	var visitors []CountryStats

	if err := store.DB.Select(&visitors, query, tenantID, from, to); err != nil {
		return nil, err
	}

	return visitors, nil
}

// PageVisitors implements the Store interface.
func (store *PostgresStore) PageVisitors(tenantID sql.NullInt64, path string, from, to time.Time) ([]Stats, error) {
	query := `SELECT "d" AS "day",
		COALESCE("path", '') "path",
		COALESCE("visitor_stats".visitors, 0) "visitors",
		COALESCE("visitor_stats".sessions, 0) "sessions",
        COALESCE("visitor_stats".bounces, 0) "bounces",
        COALESCE("visitor_stats".views, 0) "views"
		FROM (
			SELECT * FROM generate_series(
				$2::date,
				$3::date,
				INTERVAL '1 day'
			) "d"
		) AS date_series
		LEFT JOIN "visitor_stats" ON ($1::bigint IS NULL OR tenant_id = $1)
		AND "visitor_stats"."day" = "d"
		AND LOWER("path") = LOWER($4)
		ORDER BY "d" ASC`
	var visitors []Stats

	if err := store.DB.Select(&visitors, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return visitors, nil
}

// PageLanguages implements the Store interface.
func (store *PostgresStore) PageLanguages(tenantID sql.NullInt64, path string, from time.Time, to time.Time) ([]LanguageStats, error) {
	query := `SELECT * FROM (
			SELECT "language", sum("visitors") "visitors" FROM (
				SELECT "language", sum("visitors") "visitors"
				FROM "language_stats"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= DATE($2::timestamp)
				AND "day" <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "language"
				UNION
				SELECT "language", count(DISTINCT fingerprint) "visitors"
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") >= DATE($2::timestamp)
				AND DATE("time") <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "language"
			) AS results
			GROUP BY "language"
		) AS languages
		ORDER BY "visitors" DESC`
	var languages []LanguageStats

	if err := store.DB.Select(&languages, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return languages, nil
}

// PageReferrer implements the Store interface.
func (store *PostgresStore) PageReferrer(tenantID sql.NullInt64, path string, from time.Time, to time.Time) ([]ReferrerStats, error) {
	query := `SELECT * FROM (
			SELECT "referrer", "referrer_name", "referrer_icon", sum("visitors") "visitors", sum("bounces") "bounces"
			FROM (
				SELECT "referrer",
				       "referrer_name",
				       "referrer_icon",
				       sum("visitors") "visitors",
				       sum("bounces") "bounces"
				FROM "referrer_stats"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= DATE($2::timestamp)
				AND "day" <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "referrer", "referrer_name", "referrer_icon"
				UNION
				SELECT "referrer",
				       "referrer_name",
				       "referrer_icon",
				       count(DISTINCT fingerprint) "visitors",
					   0 "bounces"
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") >= DATE($2::timestamp)
				AND DATE("time") <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "referrer", "referrer_name", "referrer_icon"
			) AS results
			GROUP BY "referrer", "referrer_name", "referrer_icon"
		) AS referrer
		ORDER BY "visitors" DESC`
	var referrer []ReferrerStats

	if err := store.DB.Select(&referrer, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return referrer, nil
}

// PageOS implements the Store interface.
func (store *PostgresStore) PageOS(tenantID sql.NullInt64, path string, from time.Time, to time.Time) ([]OSStats, error) {
	query := `SELECT * FROM (
			SELECT "os", sum("visitors") "visitors" FROM (
				SELECT "os", sum("visitors") "visitors"
				FROM "os_stats"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= DATE($2::timestamp)
				AND "day" <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "os"
				UNION
				SELECT "os", count(DISTINCT fingerprint) "visitors"
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") >= DATE($2::timestamp)
				AND DATE("time") <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "os"
			) AS results
			GROUP BY "os"
		) AS os
		ORDER BY "visitors" DESC`
	var osStats []OSStats

	if err := store.DB.Select(&osStats, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return osStats, nil
}

// PageBrowser implements the Store interface.
func (store *PostgresStore) PageBrowser(tenantID sql.NullInt64, path string, from time.Time, to time.Time) ([]BrowserStats, error) {
	query := `SELECT * FROM (
			SELECT "browser", sum("visitors") "visitors" FROM (
				SELECT "browser", sum("visitors") "visitors"
				FROM "browser_stats"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND "day" >= DATE($2::timestamp)
				AND "day" <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "browser"
				UNION
				SELECT "browser", count(DISTINCT fingerprint) "visitors"
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") >= DATE($2::timestamp)
				AND DATE("time") <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				GROUP BY "browser"
			) AS results
			GROUP BY "browser"
		) AS browser
		ORDER BY "visitors" DESC`
	var browser []BrowserStats

	if err := store.DB.Select(&browser, query, tenantID, from, to, path); err != nil {
		return nil, err
	}

	return browser, nil
}

// PagePlatform implements the Store interface.
func (store *PostgresStore) PagePlatform(tenantID sql.NullInt64, path string, from time.Time, to time.Time) *VisitorStats {
	query := `SELECT COALESCE(SUM("platform_desktop"), 0) "platform_desktop",
		COALESCE(SUM("platform_mobile"), 0) "platform_mobile",
		COALESCE(SUM("platform_unknown"), 0) "platform_unknown"
		FROM (
			SELECT (
				SELECT COUNT(DISTINCT "fingerprint")
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") >= DATE($2::timestamp)
				AND DATE("time") <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				AND desktop IS TRUE
				AND mobile IS FALSE
			) AS "platform_desktop",
			(
				SELECT COUNT(DISTINCT "fingerprint")
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") >= DATE($2::timestamp)
				AND DATE("time") <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				AND desktop IS FALSE
				AND mobile IS TRUE
			) AS "platform_mobile",
			(
				SELECT COUNT(DISTINCT "fingerprint")
				FROM "hit"
				WHERE ($1::bigint IS NULL OR tenant_id = $1)
				AND DATE("time") >= DATE($2::timestamp)
				AND DATE("time") <= DATE($3::timestamp)
				AND LOWER("path") = LOWER($4)
				AND desktop IS FALSE
				AND mobile IS FALSE
			) AS "platform_unknown"
			UNION
			SELECT COALESCE(SUM("platform_desktop"), 0) "platform_desktop",
			COALESCE(SUM("platform_mobile"), 0) "platform_mobile",
			COALESCE(SUM("platform_unknown"), 0) "platform_unknown"
			FROM "visitor_stats"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND "day" >= $2::date
			AND "day" <= $3::date
			AND LOWER("path") = LOWER($4)
		) AS platforms`
	visitors := new(VisitorStats)

	if err := store.DB.Get(visitors, query, tenantID, from, to, path); err != nil && err != sql.ErrNoRows {
		store.logger.Printf("error reading page platforms: %s", err)
		return nil
	}

	return visitors
}

// VisitorsSum implements the Store interface.
func (store *PostgresStore) VisitorsSum(tenantID sql.NullInt64, from, to time.Time, path string) (*Stats, error) {
	args := make([]interface{}, 0, 4)
	args = append(args, tenantID)
	args = append(args, from)
	args = append(args, to)
	query := `SELECT COALESCE(SUM("visitors"), 0) "visitors",
        COALESCE(SUM("sessions"), 0) "sessions",
        COALESCE(SUM("bounces"), 0) "bounces",
        COALESCE(SUM("views"), 0) "views"
		FROM "visitor_stats"
		WHERE ($1::bigint IS NULL OR tenant_id = $1)
		AND "day" >= $2::date
		AND "day" <= $3::date `

	if path != "" {
		args = append(args, path)
		query += `AND LOWER("path") = LOWER($4) `
	} else {
		query += `AND "path" IS NULL `
	}

	visitors := new(Stats)

	if err := store.DB.Get(visitors, query, args...); err != nil {
		return nil, err
	}

	return visitors, nil
}

// SessionDurationSum implements the Store interface.
func (store *PostgresStore) SessionDurationSum(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) int {
	if tx == nil {
		tx = store.NewTx()
		defer store.Commit(tx)
	}

	query := `SELECT EXTRACT(EPOCH FROM sum("duration"))::integer FROM (
			SELECT max("time")-min("time") "duration" 
			FROM "hit"
			WHERE ($1::bigint IS NULL OR tenant_id = $1)
			AND DATE("time") = $2::date
			GROUP BY fingerprint, "session"
		) AS results`
	var duration int

	if err := tx.Get(&duration, query, tenantID, day); err != nil && err != sql.ErrNoRows {
		store.logger.Printf("error calculating session duration: %s", err)
		return 0
	}

	return duration
}

func (store *PostgresStore) createUpdateEntity(tx *sqlx.Tx, entity, existing statsEntity, found bool, insertQuery, updateQuery string) error {
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
