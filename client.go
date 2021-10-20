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

	query, err := tx.Prepare(`INSERT INTO "session" (sign, client_id, visitor_id, session_id, time, start, duration_seconds,
		path, entry_path, page_views, is_bounce, title, language, country_code, city, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		return err
	}

	for _, hit := range hits {
		_, err := query.Exec(hit.Sign,
			hit.ClientID,
			hit.VisitorID,
			hit.SessionID,
			hit.Time,
			hit.Start,
			hit.DurationSeconds,
			hit.Path,
			hit.EntryPath,
			hit.PageViews,
			hit.IsBounce,
			hit.Title,
			hit.Language,
			hit.CountryCode,
			hit.City,
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

	query, err := tx.Prepare(`INSERT INTO "event" (client_id, visitor_id, time, session_id, event_name, event_meta_keys, event_meta_values, duration_seconds,
		path, title, language, country_code, city, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		return err
	}

	for _, event := range events {
		_, err := query.Exec(event.ClientID,
			event.VisitorID,
			event.Time,
			event.SessionID,
			event.Name,
			event.MetaKeys,
			event.MetaValues,
			event.DurationSeconds,
			event.Path,
			event.Title,
			event.Language,
			event.CountryCode,
			event.City,
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
			event.UTMTerm)

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

// SaveUserAgents implements the Store interface.
func (client *Client) SaveUserAgents(userAgents []UserAgent) error {
	tx, err := client.Beginx()

	if err != nil {
		return err
	}

	query, err := tx.Prepare(`INSERT INTO "user_agent" (time, user_agent) VALUES (?,?)`)

	if err != nil {
		return err
	}

	for _, ua := range userAgents {
		_, err := query.Exec(ua.Time, ua.UserAgent)

		if err != nil {
			if e := tx.Rollback(); e != nil {
				client.logger.Printf("error rolling back transaction to save user agents: %s", err)
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
func (client *Client) Session(clientID, fingerprint uint64, maxAge time.Time) (*Hit, error) {
	query := `SELECT * FROM session WHERE client_id = ? AND visitor_id = ? AND time > ? ORDER BY time DESC LIMIT 1`
	hit := new(Hit)

	if err := client.DB.Get(hit, query, clientID, fingerprint, maxAge); err != nil && err != sql.ErrNoRows {
		client.logger.Printf("error reading session: %s", err)
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return hit, nil
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
