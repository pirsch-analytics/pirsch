package pirsch

import (
	// ClickHouse is an essential part of Pirsch.
	_ "github.com/ClickHouse/clickhouse-go"

	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
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

	query, err := tx.Prepare(`INSERT INTO "hit" (client_id, fingerprint, time, session, previous_time_on_page_seconds,
		user_agent, path, url, title, language, country_code, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		return err
	}

	for _, hit := range hits {
		_, err := query.Exec(hit.ClientID,
			hit.Fingerprint,
			hit.Time,
			hit.Session,
			hit.PreviousTimeOnPageSeconds,
			hit.UserAgent,
			hit.Path,
			hit.URL,
			hit.Title,
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
			hit.ScreenClass,
			hit.UTMSource,
			hit.UTMMedium,
			hit.UTMCampaign,
			hit.UTMContent,
			hit.UTMTerm)

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

// SaveEvents implements the Store interface.
func (client *Client) SaveEvents(events []Event) error {
	tx, err := client.Beginx()

	if err != nil {
		return err
	}

	query, err := tx.Prepare(`INSERT INTO "event" (client_id, fingerprint, time, session, previous_time_on_page_seconds,
		user_agent, path, url, title, language, country_code, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term,
		event_name, event_duration_seconds, event_meta_keys, event_meta_values) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		return err
	}

	for _, event := range events {
		_, err := query.Exec(event.ClientID,
			event.Fingerprint,
			event.Time,
			event.Session,
			event.PreviousTimeOnPageSeconds,
			event.UserAgent,
			event.Path,
			event.URL,
			event.Title,
			event.Language,
			event.CountryCode,
			event.Referrer,
			event.ReferrerName,
			event.ReferrerIcon,
			event.OS,
			event.OSVersion,
			event.Browser,
			event.BrowserVersion,
			client.boolean(event.Desktop),
			client.boolean(event.Mobile),
			event.ScreenWidth,
			event.ScreenHeight,
			event.ScreenClass,
			event.UTMSource,
			event.UTMMedium,
			event.UTMCampaign,
			event.UTMContent,
			event.UTMTerm,
			event.Name,
			event.DurationSeconds,
			event.MetaKeys,
			event.MetaValues)

		if err != nil {
			if e := tx.Rollback(); e != nil {
				client.logger.Printf("error rolling back transaction to save events: %s", err)
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
func (client *Client) Session(clientID int64, fingerprint string, maxAge time.Time) (string, time.Time, time.Time, error) {
	query := `SELECT path, time, session FROM hit WHERE client_id = ? AND fingerprint = ? AND time > ? ORDER BY time DESC LIMIT 1`
	data := struct {
		Path    string
		Time    time.Time
		Session time.Time
	}{}

	if err := client.DB.Get(&data, query, clientID, fingerprint, maxAge); err != nil && err != sql.ErrNoRows {
		client.logger.Printf("error reading session timestamp: %s", err)
		return "", time.Time{}, time.Time{}, err
	}

	return data.Path, data.Time, data.Session, nil
}

// Count implements the Store interface.
func (client *Client) Count(query string, args ...interface{}) (int, error) {
	count := 0

	if err := client.DB.Get(&count, query, args...); err != nil {
		client.logger.Printf("error counting results: %s", err)
		return 0, err
	}

	return count, nil
}

// Get implements the Store interface.
func (client *Client) Get(result interface{}, query string, args ...interface{}) error {
	if err := client.DB.Get(result, query, args...); err != nil {
		client.logger.Printf("error getting result: %s", err)
		return err
	}

	return nil
}

// Select implements the Store interface.
func (client *Client) Select(results interface{}, query string, args ...interface{}) error {
	if err := client.DB.Select(results, query, args...); err != nil {
		client.logger.Printf("error selecting results: %s", err)
		return err
	}

	return nil
}

func (client *Client) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
