package db

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/util"
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
		config.Logger = util.GetDefaultLogger()
	}
}

// Client is a ClickHouse database client.
type Client struct {
	*sql.DB
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

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Client{
		db,
		config.Logger,
		config.Debug,
	}, nil
}

// SavePageViews implements the Store interface.
func (client *Client) SavePageViews(pageViews []model.PageView) error {
	tx, err := client.Begin()

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
func (client *Client) SaveSessions(sessions []model.Session) error {
	tx, err := client.Begin()

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
func (client *Client) SaveEvents(events []model.Event) error {
	tx, err := client.Begin()

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
func (client *Client) SaveUserAgents(userAgents []model.UserAgent) error {
	tx, err := client.Begin()

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
func (client *Client) Session(clientID, fingerprint uint64, maxAge time.Time) (*model.Session, error) {
	query := `SELECT sign,
        client_id,
		visitor_id,
		session_id,
		time,
		start,
		duration_seconds,
		entry_path,
		exit_path,
		page_views,
		is_bounce,
		entry_title,
		exit_title,
		language,
		country_code,
		city,
		referrer,
		referrer_name,
		referrer_icon,
		os,
		os_version,
		browser,
		browser_version,
		desktop,
		mobile,
		screen_width,
		screen_height,
		screen_class,
		utm_source,
		utm_medium,
		utm_campaign,
		utm_content,
		utm_term,
		is_bot
		FROM session
		WHERE client_id = ?
		AND visitor_id = ?
		AND time > ?
		ORDER BY time DESC
		LIMIT 1`
	session := new(model.Session)
	err := client.QueryRow(query, clientID, fingerprint, maxAge).Scan(&session.Sign,
		&session.ClientID,
		&session.VisitorID,
		&session.SessionID,
		&session.Time,
		&session.Start,
		&session.DurationSeconds,
		&session.EntryPath,
		&session.ExitPath,
		&session.PageViews,
		&session.IsBounce,
		&session.EntryTitle,
		&session.ExitTitle,
		&session.Language,
		&session.CountryCode,
		&session.City,
		&session.Referrer,
		&session.ReferrerName,
		&session.ReferrerIcon,
		&session.OS,
		&session.OSVersion,
		&session.Browser,
		&session.BrowserVersion,
		&session.Desktop,
		&session.Mobile,
		&session.ScreenWidth,
		&session.ScreenHeight,
		&session.ScreenClass,
		&session.UTMSource,
		&session.UTMMedium,
		&session.UTMCampaign,
		&session.UTMContent,
		&session.UTMTerm,
		&session.IsBot)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			client.logger.Printf("error reading session: %s", err)
			return nil, err
		}
	}

	return session, nil
}

// Count implements the Store interface.
func (client *Client) Count(query string, args ...any) (int, error) {
	var count int

	if err := client.QueryRow(query, args...).Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		} else {
			client.logger.Printf("error counting results: %s", err)
			return 0, err
		}
	}

	return count, nil
}

// SelectActiveVisitorStats implements the Store interface.
func (client *Client) SelectActiveVisitorStats(includeTitle bool, query string, args ...any) ([]model.ActiveVisitorStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.ActiveVisitorStats

	if includeTitle {
		for rows.Next() {
			var result model.ActiveVisitorStats

			if err := rows.Scan(&result.Path, &result.Title, &result.Visitors); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	} else {
		for rows.Next() {
			var result model.ActiveVisitorStats

			if err := rows.Scan(&result.Path, &result.Visitors); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// GetTotalVisitorStats implements the Store interface.
func (client *Client) GetTotalVisitorStats(query string, args ...any) (*model.TotalVisitorStats, error) {
	result := new(model.TotalVisitorStats)

	if err := client.QueryRow(query, args...).Scan(&result.Visitors,
		&result.Sessions,
		&result.Views,
		&result.Bounces,
		&result.BounceRate); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return result, nil
}

// SelectVisitorStats implements the Store interface.
func (client *Client) SelectVisitorStats(period pirsch.Period, query string, args ...any) ([]model.VisitorStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.VisitorStats

	switch period {
	case pirsch.PeriodWeek:
		for rows.Next() {
			var result model.VisitorStats

			if err := rows.Scan(&result.Week,
				&result.Visitors,
				&result.Sessions,
				&result.Views,
				&result.Bounces,
				&result.BounceRate); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	case pirsch.PeriodMonth:
		for rows.Next() {
			var result model.VisitorStats

			if err := rows.Scan(&result.Month,
				&result.Visitors,
				&result.Sessions,
				&result.Views,
				&result.Bounces,
				&result.BounceRate); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	case pirsch.PeriodYear:
		for rows.Next() {
			var result model.VisitorStats

			if err := rows.Scan(&result.Year,
				&result.Visitors,
				&result.Sessions,
				&result.Views,
				&result.Bounces,
				&result.BounceRate); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	default:
		for rows.Next() {
			var result model.VisitorStats

			if err := rows.Scan(&result.Day,
				&result.Visitors,
				&result.Sessions,
				&result.Views,
				&result.Bounces,
				&result.BounceRate); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SelectTimeSpentStats implements the Store interface.
func (client *Client) SelectTimeSpentStats(period pirsch.Period, query string, args ...any) ([]model.TimeSpentStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.TimeSpentStats

	switch period {
	case pirsch.PeriodWeek:
		for rows.Next() {
			var result model.TimeSpentStats

			if err := rows.Scan(&result.AverageTimeSpentSeconds, &result.Week); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	case pirsch.PeriodMonth:
		for rows.Next() {
			var result model.TimeSpentStats

			if err := rows.Scan(&result.AverageTimeSpentSeconds, &result.Month); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	case pirsch.PeriodYear:
		for rows.Next() {
			var result model.TimeSpentStats

			if err := rows.Scan(&result.AverageTimeSpentSeconds, &result.Year); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	default:
		for rows.Next() {
			var result model.TimeSpentStats

			if err := rows.Scan(&result.Day, &result.AverageTimeSpentSeconds); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// GetGrowthStats implements the Store interface.
func (client *Client) GetGrowthStats(query string, args ...any) (*model.GrowthStats, error) {
	result := new(model.GrowthStats)

	if err := client.QueryRow(query, args...).Scan(&result.Visitors,
		&result.Sessions,
		&result.Views,
		&result.Bounces,
		&result.BounceRate); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return result, nil
}

// SelectVisitorHourStats implements the Store interface.
func (client *Client) SelectVisitorHourStats(query string, args ...any) ([]model.VisitorHourStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.VisitorHourStats

	for rows.Next() {
		var result model.VisitorHourStats

		if err := rows.Scan(&result.Hour,
			&result.Visitors,
			&result.Sessions,
			&result.Views,
			&result.Bounces,
			&result.BounceRate); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectPageStats implements the Store interface.
func (client *Client) SelectPageStats(includeTitle, includeEvent bool, query string, args ...any) ([]model.PageStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.PageStats

	if includeTitle {
		if includeEvent {
			for rows.Next() {
				var result model.PageStats

				if err := rows.Scan(&result.Path,
					&result.Visitors,
					&result.Sessions,
					&result.RelativeVisitors,
					&result.Views,
					&result.RelativeViews,
					&result.Bounces,
					&result.BounceRate,
					&result.Title,
					&result.AverageTimeSpentSeconds); err != nil {
					return nil, err
				}

				results = append(results, result)
			}
		} else {
			for rows.Next() {
				var result model.PageStats

				if err := rows.Scan(&result.Path,
					&result.Visitors,
					&result.Sessions,
					&result.RelativeVisitors,
					&result.Views,
					&result.RelativeViews,
					&result.Bounces,
					&result.BounceRate,
					&result.Title); err != nil {
					return nil, err
				}

				results = append(results, result)
			}
		}
	} else {
		if includeEvent {
			for rows.Next() {
				var result model.PageStats

				if err := rows.Scan(&result.Path,
					&result.Visitors,
					&result.Sessions,
					&result.RelativeVisitors,
					&result.Views,
					&result.RelativeViews,
					&result.Bounces,
					&result.BounceRate,
					&result.AverageTimeSpentSeconds); err != nil {
					return nil, err
				}

				results = append(results, result)
			}
		} else {
			for rows.Next() {
				var result model.PageStats

				if err := rows.Scan(&result.Path,
					&result.Visitors,
					&result.Sessions,
					&result.RelativeVisitors,
					&result.Views,
					&result.RelativeViews,
					&result.Bounces,
					&result.BounceRate); err != nil {
					return nil, err
				}

				results = append(results, result)
			}
		}
	}

	return results, nil
}

// SelectAvgTimeSpentStats implements the Store interface.
func (client *Client) SelectAvgTimeSpentStats(query string, args ...any) ([]model.AvgTimeSpentStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.AvgTimeSpentStats

	for rows.Next() {
		var result model.AvgTimeSpentStats

		if err := rows.Scan(&result.Path, &result.AverageTimeSpentSeconds); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectEntryStats implements the Store interface.
func (client *Client) SelectEntryStats(includeTitle bool, query string, args ...any) ([]model.EntryStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.EntryStats

	if includeTitle {
		for rows.Next() {
			var result model.EntryStats

			if err := rows.Scan(&result.Path, &result.Entries, &result.Title); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	} else {
		for rows.Next() {
			var result model.EntryStats

			if err := rows.Scan(&result.Path, &result.Entries); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SelectExitStats implements the Store interface.
func (client *Client) SelectExitStats(includeTitle bool, query string, args ...any) ([]model.ExitStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.ExitStats

	if includeTitle {
		for rows.Next() {
			var result model.ExitStats

			if err := rows.Scan(&result.Path, &result.Exits, &result.Title); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	} else {
		for rows.Next() {
			var result model.ExitStats

			if err := rows.Scan(&result.Path, &result.Exits); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SelectTotalSessions implements the Store interface.
func (client *Client) SelectTotalSessions(query string, args ...any) (int, error) {
	var result int

	if err := client.QueryRow(query, args...).Scan(&result); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return result, nil
}

// SelectTotalVisitorSessionStats implements the Store interface.
func (client *Client) SelectTotalVisitorSessionStats(query string, args ...any) ([]model.TotalVisitorSessionStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.TotalVisitorSessionStats

	for rows.Next() {
		var result model.TotalVisitorSessionStats

		if err := rows.Scan(&result.Path, &result.Visitors, &result.Sessions, &result.Views); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// GetPageConversionsStats implements the Store interface.
func (client *Client) GetPageConversionsStats(query string, args ...any) (*model.PageConversionsStats, error) {
	result := new(model.PageConversionsStats)

	if err := client.QueryRow(query, args...).Scan(&result.Visitors,
		&result.Views,
		&result.CR); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return result, nil
}

// SelectEventStats implements the Store interface.
func (client *Client) SelectEventStats(breakdown bool, query string, args ...any) ([]model.EventStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.EventStats

	if breakdown {
		for rows.Next() {
			var result model.EventStats

			if err := rows.Scan(&result.Name,
				&result.Visitors,
				&result.Views,
				&result.CR,
				&result.AverageDurationSeconds,
				&result.MetaValue); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	} else {
		for rows.Next() {
			var result model.EventStats

			if err := rows.Scan(&result.Name,
				&result.Visitors,
				&result.Views,
				&result.CR,
				&result.AverageDurationSeconds,
				&result.MetaKeys); err != nil {
				return nil, err
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SelectEventListStats implements the Store interface.
func (client *Client) SelectEventListStats(query string, args ...any) ([]model.EventListStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.EventListStats

	for rows.Next() {
		var result model.EventListStats

		if err := rows.Scan(&result.Name, &result.Meta, &result.Visitors, &result.Count); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectReferrerStats implements the Store interface.
func (client *Client) SelectReferrerStats(query string, args ...any) ([]model.ReferrerStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.ReferrerStats

	for rows.Next() {
		var result model.ReferrerStats

		if err := rows.Scan(&result.ReferrerName,
			&result.ReferrerIcon,
			&result.Visitors,
			&result.Sessions,
			&result.RelativeVisitors,
			&result.Bounces,
			&result.BounceRate,
			&result.Referrer); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// GetPlatformStats implements the Store interface.
func (client *Client) GetPlatformStats(query string, args ...any) (*model.PlatformStats, error) {
	result := new(model.PlatformStats)

	if err := client.QueryRow(query, args...).Scan(&result.PlatformDesktop,
		&result.PlatformMobile,
		&result.PlatformUnknown,
		&result.RelativePlatformDesktop,
		&result.RelativePlatformMobile,
		&result.RelativePlatformUnknown); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return result, nil
}

// SelectLanguageStats implements the Store interface.
func (client *Client) SelectLanguageStats(query string, args ...any) ([]model.LanguageStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.LanguageStats

	for rows.Next() {
		var result model.LanguageStats

		if err := rows.Scan(&result.Language, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectCountryStats implements the Store interface.
func (client *Client) SelectCountryStats(query string, args ...any) ([]model.CountryStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.CountryStats

	for rows.Next() {
		var result model.CountryStats

		if err := rows.Scan(&result.CountryCode, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectCityStats implements the Store interface.
func (client *Client) SelectCityStats(query string, args ...any) ([]model.CityStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.CityStats

	for rows.Next() {
		var result model.CityStats

		if err := rows.Scan(&result.City, &result.CountryCode, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectBrowserStats implements the Store interface.
func (client *Client) SelectBrowserStats(query string, args ...any) ([]model.BrowserStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.BrowserStats

	for rows.Next() {
		var result model.BrowserStats

		if err := rows.Scan(&result.Browser, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectOSStats implements the Store interface.
func (client *Client) SelectOSStats(query string, args ...any) ([]model.OSStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.OSStats

	for rows.Next() {
		var result model.OSStats

		if err := rows.Scan(&result.OS, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectScreenClassStats implements the Store interface.
func (client *Client) SelectScreenClassStats(query string, args ...any) ([]model.ScreenClassStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.ScreenClassStats

	for rows.Next() {
		var result model.ScreenClassStats

		if err := rows.Scan(&result.ScreenClass, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectUTMSourceStats implements the Store interface.
func (client *Client) SelectUTMSourceStats(query string, args ...any) ([]model.UTMSourceStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.UTMSourceStats

	for rows.Next() {
		var result model.UTMSourceStats

		if err := rows.Scan(&result.UTMSource, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectUTMMediumStats implements the Store interface.
func (client *Client) SelectUTMMediumStats(query string, args ...any) ([]model.UTMMediumStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.UTMMediumStats

	for rows.Next() {
		var result model.UTMMediumStats

		if err := rows.Scan(&result.UTMMedium, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectUTMCampaignStats implements the Store interface.
func (client *Client) SelectUTMCampaignStats(query string, args ...any) ([]model.UTMCampaignStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.UTMCampaignStats

	for rows.Next() {
		var result model.UTMCampaignStats

		if err := rows.Scan(&result.UTMCampaign, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectUTMContentStats implements the Store interface.
func (client *Client) SelectUTMContentStats(query string, args ...any) ([]model.UTMContentStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.UTMContentStats

	for rows.Next() {
		var result model.UTMContentStats

		if err := rows.Scan(&result.UTMContent, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectUTMTermStats implements the Store interface.
func (client *Client) SelectUTMTermStats(query string, args ...any) ([]model.UTMTermStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.UTMTermStats

	for rows.Next() {
		var result model.UTMTermStats

		if err := rows.Scan(&result.UTMTerm, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectOSVersionStats implements the Store interface.
func (client *Client) SelectOSVersionStats(query string, args ...any) ([]model.OSVersionStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.OSVersionStats

	for rows.Next() {
		var result model.OSVersionStats

		if err := rows.Scan(&result.OS, &result.OSVersion, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectBrowserVersionStats implements the Store interface.
func (client *Client) SelectBrowserVersionStats(query string, args ...any) ([]model.BrowserVersionStats, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []model.BrowserVersionStats

	for rows.Next() {
		var result model.BrowserVersionStats

		if err := rows.Scan(&result.Browser, &result.BrowserVersion, &result.Visitors, &result.RelativeVisitors); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// SelectOptions implements the Store interface.
func (client *Client) SelectOptions(query string, args ...any) ([]string, error) {
	rows, err := client.Query(query, args...)

	if err != nil {
		return nil, err
	}

	defer client.closeRows(rows)
	var results []string

	for rows.Next() {
		var result string

		if err := rows.Scan(&result); err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (client *Client) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}

func (client *Client) closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		client.logger.Printf("error closing rows: %s", err)
	}
}
