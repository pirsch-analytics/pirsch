package pirsch

import (
	"database/sql"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"strings"
	"time"
)

// Client is a ClickHouse database client.
type Client struct {
	sqlx.DB
	logger *log.Logger
}

// NewClient returns a new client for given database connection string.
// The logger is optional.
func NewClient(connection string, logger *log.Logger) (*Client, error) {
	c, err := sqlx.Open("clickhouse", connection)

	if err != nil {
		return nil, err
	}

	if err := c.Ping(); err != nil {
		return nil, err
	}

	if logger == nil {
		logger = log.New(os.Stdout, "[pirsch] ", log.LstdFlags)
	}

	return &Client{
		*c,
		logger,
	}, nil
}

// SaveHits implements the Store interface.
func (client *Client) SaveHits(hits []Hit) error {
	tx, err := client.Beginx()

	if err != nil {
		return err
	}

	query, err := tx.Prepare(`INSERT INTO "hit" (tenant_id, fingerprint, time, session, user_agent,
		path, url, language, country_code, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		return err
	}

	for _, hit := range hits {
		_, err := query.Exec(hit.TenantID,
			hit.Fingerprint,
			hit.Time,
			hit.Session,
			hit.UserAgent,
			hit.Path,
			hit.URL,
			hit.Language,
			hit.CountryCode,
			hit.Referrer,
			hit.ReferrerName,
			hit.ReferrerIcon,
			hit.OS,
			hit.OSVersion,
			hit.Browser,
			hit.BrowserVersion,
			client.boolean(hit.Desktop),
			client.boolean(hit.Mobile),
			hit.ScreenWidth,
			hit.ScreenHeight,
			hit.ScreenClass)

		if err != nil {
			if e := tx.Rollback(); e != nil {
				client.logger.Printf("error rolling back transaction to save hits: %s", err)
			}

			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Session implements the Store interface.
func (client *Client) Session(tenantID sql.NullInt64, fingerprint string, maxAge time.Time) (time.Time, error) {
	args := make([]interface{}, 0, 3)

	if tenantID.Valid {
		args = append(args, tenantID.Int64)
	}

	args = append(args, fingerprint)
	args = append(args, maxAge)
	query := `SELECT "session" FROM "hit" WHERE ` +
		client.tenant(tenantID) +
		`AND fingerprint = ? AND "time" > ? LIMIT 1`
	var session time.Time

	if err := client.Get(&session, query, args...); err != nil && err != sql.ErrNoRows {
		client.logger.Printf("error reading session timestamp: %s", err)
		return session, err
	}

	return session, nil
}

// CountActiveVisitors implements the Store interface.
func (client *Client) CountActiveVisitors(filter *Filter) int {
	args, filterQuery := client.filter(filter)
	query := `SELECT count(DISTINCT fingerprint) "visitors" FROM "hit" WHERE ` + filterQuery
	visitors := 0

	if err := client.Get(&visitors, query, args...); err != nil {
		client.logger.Printf("error counting active visitors: %s", err)
		return 0
	}

	return visitors
}

// ActiveVisitors implements the Store interface.
func (client *Client) ActiveVisitors(filter *Filter) ([]Stats, error) {
	args, filterQuery := client.filter(filter)
	query := `SELECT "path", count(DISTINCT fingerprint) "visitors" FROM "hit" WHERE ` +
		filterQuery +
		`GROUP BY "path"
		ORDER BY "visitors" DESC, "path" ASC`
	var stats []Stats

	if err := client.Select(&stats, query, args...); err != nil {
		client.logger.Printf("error reading active visitors: %s", err)
		return nil, err
	}

	return stats, nil
}

// VisitorLanguages implements the Store interface.
func (client *Client) VisitorLanguages(filter *Filter) ([]Stats, error) {
	args, filterQuery := client.filter(filter)
	query := `SELECT "language", count(DISTINCT fingerprint) "visitors", toDate("time") "day" FROM "hit" WHERE ` +
		filterQuery +
		`GROUP BY "day", "language"
		ORDER BY "visitors" DESC, "language" ASC`
	var stats []Stats

	if err := client.Select(&stats, query, args...); err != nil {
		client.logger.Printf("error reading visitor languages: %s", err)
		return nil, err
	}

	return stats, nil
}

func (client *Client) filter(filter *Filter) ([]interface{}, string) {
	args := make([]interface{}, 0, 5)
	var query strings.Builder

	if filter.TenantID.Valid {
		args = append(args, filter.TenantID)
		query.WriteString("tenant_id = ? ")
	} else {
		query.WriteString("tenant_id IS NULL ")
	}

	if !filter.From.IsZero() {
		args = append(args, filter.From)
		query.WriteString("AND time >= ? ")
	}

	if !filter.To.IsZero() {
		args = append(args, filter.To)
		query.WriteString("AND time <= ? ")
	}

	if !filter.Day.IsZero() {
		args = append(args, filter.Day)
		query.WriteString("AND toDate(time) = ?")
	}

	if filter.Path != "" {
		args = append(args, filter.Path)
		query.WriteString("AND path = ? ")
	}

	return args, query.String()
}

func (client *Client) tenant(tenantID sql.NullInt64) string {
	if tenantID.Valid {
		return "tenant_id = ? "
	}

	return "tenant_id IS NULL "
}

func (client *Client) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
