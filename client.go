package pirsch

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

const (
	defaultMaxOpenConnections    = 20
	defaultMaxConnectionLifetime = 1800
	defaultMaxIdleConnections    = 5
	defaultMaxConnectionIdleTime = 300
)

// ClientConfig is the optional configuration for the Client.
type ClientConfig struct {
	// Hostname is the database hostname.
	Hostname string

	// Port is the database port.
	Port int

	// Database is the database schema.
	Database string

	// Username is the database user.
	Username string

	// Password is the database password.
	Password string

	// Secure enables TLS encryption.
	Secure bool

	// SSLSkipVerify skips the SSL verification if set to true.
	SSLSkipVerify bool

	// MaxOpenConnections sets the number of maximum open connections.
	// If set to <= 0, the default value of 20 will be used.
	MaxOpenConnections int

	// MaxConnectionLifetimeSeconds sets the maximum amount of time a connection will be reused.
	// If set to <= 0, the default value of 1800 will be used.
	MaxConnectionLifetimeSeconds int

	// MaxIdleConnections sets the number of maximum idle connections.
	// If set to <= 0, the default value of 5 will be used.
	MaxIdleConnections int

	// MaxConnectionIdleTimeSeconds sets the maximum amount of time a connection can be idle.
	// If set to <= 0, the default value of 300 will be used.
	MaxConnectionIdleTimeSeconds int

	// Logger is the log.Logger used for logging.
	// The default log will be used printing to os.Stdout with "pirsch" in its prefix in case it is not set.
	Logger *log.Logger

	// Debug will enable verbose logging.
	Debug bool
}

func (config *ClientConfig) validate() {
	if config.MaxOpenConnections <= 0 {
		config.MaxOpenConnections = defaultMaxOpenConnections
	}

	if config.MaxConnectionLifetimeSeconds <= 0 {
		config.MaxConnectionLifetimeSeconds = defaultMaxConnectionLifetime
	}

	if config.MaxIdleConnections <= 0 {
		config.MaxIdleConnections = defaultMaxIdleConnections
	}

	if config.MaxConnectionIdleTimeSeconds <= 0 {
		config.MaxConnectionIdleTimeSeconds = defaultMaxConnectionIdleTime
	}

	if config.Logger == nil {
		config.Logger = logger
	}
}

// Client is a ClickHouse database client.
type Client struct {
	sqlx.DB
	logger *log.Logger
	debug  bool
}

// NewClient returns a new client for given database connection string.
// Pass nil for the config to use the defaults.
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		return nil, errors.New("configuration missing")
	}

	config.validate()
	var tlsConn *tls.Config

	if config.Secure {
		tlsConn = &tls.Config{
			InsecureSkipVerify: config.SSLSkipVerify,
		}
	}

	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Hostname, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		TLS:         tlsConn,
		DialTimeout: time.Second * 30,
		Debug:       config.Debug,
	})
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetConnMaxLifetime(time.Duration(config.MaxConnectionLifetimeSeconds) * time.Second)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxIdleTime(time.Duration(config.MaxConnectionIdleTimeSeconds) * time.Second)
	c := sqlx.NewDb(db, "clickhouse")

	if err := c.Ping(); err != nil {
		return nil, err
	}

	return &Client{
		*c,
		config.Logger,
		config.Debug,
	}, nil
}

// SavePageViews implements the Store interface.
func (client *Client) SavePageViews(pageViews []PageView) error {
	tx, err := client.Beginx()

	if err != nil {
		return err
	}

	query, err := tx.Prepare(`INSERT INTO "page_view" (client_id, visitor_id, session_id, time, duration_seconds,
		path, title, language, country_code, city, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		return err
	}

	for _, pageView := range pageViews {
		_, err := query.Exec(pageView.ClientID,
			pageView.VisitorID,
			pageView.SessionID,
			pageView.Time,
			pageView.DurationSeconds,
			pageView.Path,
			pageView.Title,
			pageView.Language,
			pageView.CountryCode,
			pageView.City,
			pageView.Referrer,
			pageView.ReferrerName,
			pageView.ReferrerIcon,
			pageView.OS,
			pageView.OSVersion,
			pageView.Browser,
			pageView.BrowserVersion,
			client.boolean(pageView.Desktop),
			client.boolean(pageView.Mobile),
			pageView.ScreenWidth,
			pageView.ScreenHeight,
			pageView.ScreenClass,
			pageView.UTMSource,
			pageView.UTMMedium,
			pageView.UTMCampaign,
			pageView.UTMContent,
			pageView.UTMTerm)

		if err != nil {
			if e := tx.Rollback(); e != nil {
				client.logger.Printf("error rolling back transaction to save page views: %s", err)
			}

			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if client.debug {
		client.logger.Printf("saved %d page views", len(pageViews))
	}

	return nil
}

// SaveSessions implements the Store interface.
func (client *Client) SaveSessions(sessions []Session) error {
	tx, err := client.Beginx()

	if err != nil {
		return err
	}

	query, err := tx.Prepare(`INSERT INTO "session" (sign, client_id, visitor_id, session_id, time, start, duration_seconds,
		entry_path, exit_path, page_views, is_bounce, entry_title, exit_title, language, country_code, city, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class,
		utm_source, utm_medium, utm_campaign, utm_content, utm_term,
        is_bot) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	if err != nil {
		return err
	}

	for _, session := range sessions {
		_, err := query.Exec(session.Sign,
			session.ClientID,
			session.VisitorID,
			session.SessionID,
			session.Time,
			session.Start,
			session.DurationSeconds,
			session.EntryPath,
			session.ExitPath,
			session.PageViews,
			client.boolean(session.IsBounce),
			session.EntryTitle,
			session.ExitTitle,
			session.Language,
			session.CountryCode,
			session.City,
			session.Referrer,
			session.ReferrerName,
			session.ReferrerIcon,
			session.OS,
			session.OSVersion,
			session.Browser,
			session.BrowserVersion,
			client.boolean(session.Desktop),
			client.boolean(session.Mobile),
			session.ScreenWidth,
			session.ScreenHeight,
			session.ScreenClass,
			session.UTMSource,
			session.UTMMedium,
			session.UTMCampaign,
			session.UTMContent,
			session.UTMTerm,
			session.IsBot)

		if err != nil {
			if e := tx.Rollback(); e != nil {
				client.logger.Printf("error rolling back transaction to save sessions: %s", err)
			}

			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if client.debug {
		client.logger.Printf("saved %d sessions", len(sessions))
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

	if client.debug {
		client.logger.Printf("saved %d events", len(events))
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

	if client.debug {
		client.logger.Printf("saved %d user agents", len(userAgents))
	}

	return nil
}

// Session implements the Store interface.
func (client *Client) Session(clientID, fingerprint uint64, maxAge time.Time) (*Session, error) {
	query := `SELECT * FROM session WHERE client_id = ? AND visitor_id = ? AND time > ? ORDER BY time DESC LIMIT 1`
	session := new(Session)

	if err := client.DB.Get(session, query, clientID, fingerprint, maxAge); err != nil && err != sql.ErrNoRows {
		client.logger.Printf("error reading session: %s", err)
		return nil, err
	} else if err == sql.ErrNoRows {
		return nil, nil
	}

	return session, nil
}

// Count implements the Store interface.
func (client *Client) Count(query string, args ...any) (int, error) {
	count := 0

	if err := client.DB.Get(&count, query, args...); err != nil {
		client.logger.Printf("error counting results: %s", err)
		return 0, err
	}

	return count, nil
}

// Get implements the Store interface.
func (client *Client) Get(result any, query string, args ...any) error {
	// don't return an error if nothing was found
	if err := client.DB.Get(result, query, args...); err != nil && err != sql.ErrNoRows {
		client.logger.Printf("error getting result: %s", err)
		return err
	}

	return nil
}

// Select implements the Store interface.
func (client *Client) Select(results any, query string, args ...any) error {
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
