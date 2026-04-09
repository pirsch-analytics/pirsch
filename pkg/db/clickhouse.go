package db

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

const (
	defaultMaxOpenConnections    = 20
	defaultMaxConnectionLifetime = 1800
	defaultMaxIdleConnections    = 5
	defaultMaxConnectionIdleTime = 300
)

// ClickHouse implements the Storage interface.
type ClickHouse struct {
	*sql.DB
	logger *slog.Logger
	debug  bool
	dev    bool
}

// NewClickHouse returns a new ClickHouse client for the given configuration.
func NewClickHouse(config *ClickHouseConfig) (*ClickHouse, error) {
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

	addr := make([]string, len(config.Hostnames))

	for i, hostname := range config.Hostnames {
		addr[i] = fmt.Sprintf("%s:%d", hostname, config.Port)
	}

	var logger *slog.Logger

	if config.Debug {
		logger = config.Logger
	}

	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: addr,
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		TLS:         tlsConn,
		DialTimeout: time.Second * 30,
		Logger:      logger,
	})
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(config.MaxConnectionLifetimeSeconds) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(config.MaxConnectionIdleTimeSeconds) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &ClickHouse{
		db,
		config.Logger,
		config.Debug,
		config.dev,
	}, nil
}

// SaveSessions implements the Storage interface.
func (ch *ClickHouse) SaveSessions(sessions []model.Session) error {
	values := make([]string, 0, len(sessions))
	args := make([]any, 0, len(sessions)*35)

	for _, session := range sessions {
		values = append(values, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args,
			session.Sign,
			session.Version,
			session.ClientID,
			session.VisitorID,
			session.SessionID,
			session.Time.UnixMilli(),
			session.Start,
			session.Hostname,
			session.EntryPath,
			session.ExitPath,
			session.PageViews,
			ch.boolean(session.IsBounce),
			session.EntryTitle,
			session.ExitTitle,
			session.Language,
			session.CountryCode,
			session.Region,
			session.City,
			session.Referrer,
			session.ReferrerName,
			session.ReferrerIcon,
			session.OS,
			session.OSVersion,
			session.Browser,
			session.BrowserVersion,
			ch.boolean(session.Desktop),
			ch.boolean(session.Mobile),
			session.ScreenClass,
			session.UTMSource,
			session.UTMMedium,
			session.UTMCampaign,
			session.UTMContent,
			session.UTMTerm,
			session.Channel,
			session.Extended)
	}

	query := fmt.Sprintf(`INSERT INTO "session_v7" (sign,
		version,
		client_id,
		visitor_id,
		session_id,
		time,
		start,
		hostname,
		entry_path,
		exit_path,
		page_views,
		is_bounce,
		entry_title,
		exit_title,
		language,
		country_code,
		region,
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
		screen_class,
		utm_source,
		utm_medium,
		utm_campaign,
		utm_content,
		utm_term,
		channel,
		extended) VALUES %s`, strings.Join(values, ","))

	if _, err := ch.Exec(query, args...); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("saved sessions", "count", len(sessions))
	}

	return nil
}

// SavePageViews implements the Storage interface.
func (ch *ClickHouse) SavePageViews(pageViews []model.PageView) error {
	values := make([]string, 0, len(pageViews))
	args := make([]any, 0, len(pageViews)*28)

	for _, pageView := range pageViews {
		values = append(values, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args,
			pageView.ClientID,
			pageView.VisitorID,
			pageView.SessionID,
			pageView.Time.UnixMilli(),
			pageView.Hostname,
			pageView.Path,
			pageView.Title,
			pageView.Language,
			pageView.CountryCode,
			pageView.Region,
			pageView.City,
			pageView.Referrer,
			pageView.ReferrerName,
			pageView.ReferrerIcon,
			pageView.OS,
			pageView.OSVersion,
			pageView.Browser,
			pageView.BrowserVersion,
			ch.boolean(pageView.Desktop),
			ch.boolean(pageView.Mobile),
			pageView.ScreenClass,
			pageView.UTMSource,
			pageView.UTMMedium,
			pageView.UTMCampaign,
			pageView.UTMContent,
			pageView.UTMTerm,
			pageView.Channel,
			pageView.Tags)
	}

	query := fmt.Sprintf(`INSERT INTO "page_view_v7" (client_id,
		visitor_id,
		session_id,
		time,
		hostname,
		path,
		title,
		language,
		country_code,
		region,
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
		screen_class,
		utm_source,
		utm_medium,
		utm_campaign,
		utm_content,
		utm_term,
		channel,
		tags) VALUES %s`, strings.Join(values, ","))

	if _, err := ch.Exec(query, args...); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("saved page views", "count", len(pageViews))
	}

	return nil
}

// SaveEvents implements the Storage interface.
func (ch *ClickHouse) SaveEvents(events []model.Event) error {
	values := make([]string, 0, len(events))
	args := make([]any, 0, len(events)*29)

	for _, event := range events {
		meta, _ := json.Marshal(event)
		values = append(values, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args,
			event.ClientID,
			event.VisitorID,
			event.Time.UnixMilli(),
			event.SessionID,
			event.Name,
			string(meta),
			event.Hostname,
			event.Path,
			event.Title,
			event.Language,
			event.CountryCode,
			event.Region,
			event.City,
			event.Referrer,
			event.ReferrerName,
			event.ReferrerIcon,
			event.OS,
			event.OSVersion,
			event.Browser,
			event.BrowserVersion,
			ch.boolean(event.Desktop),
			ch.boolean(event.Mobile),
			event.ScreenClass,
			event.UTMSource,
			event.UTMMedium,
			event.UTMCampaign,
			event.UTMContent,
			event.UTMTerm,
			event.Channel)
	}

	query := fmt.Sprintf(`INSERT INTO "event_v7" (client_id,
		visitor_id,
		time,
		session_id,
		name,
		meta_data,
		hostname, 
		path, 
		title, 
		language, 
		country_code, 
		region, 
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
		screen_class,
		utm_source, 
		utm_medium, 
		utm_campaign, 
		utm_content,
		utm_term, 
		channel) VALUES %s`, strings.Join(values, ","))

	if _, err := ch.Exec(query, args...); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("saved events", "count", len(events))
	}

	return nil
}

// SaveRequests implements the Storage interface.
func (ch *ClickHouse) SaveRequests(requests []model.Request) error {
	values := make([]string, 0, len(requests))
	args := make([]any, 0, len(requests)*14)

	for _, req := range requests {
		values = append(values, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args,
			req.ClientID,
			req.VisitorID,
			req.Time.UnixMilli(),
			req.IP,
			req.UserAgent,
			req.Hostname,
			req.Path,
			req.Event,
			req.Referrer,
			req.UTMSource,
			req.UTMMedium,
			req.UTMCampaign,
			req.Bot,
			req.BotReason)
	}

	query := fmt.Sprintf(`INSERT INTO "request" (client_id,
		visitor_id,
		time,
		ip,
		user_agent,
		hostname,
		path,
		event_name,
		referrer,
		utm_source,
		utm_medium,
		utm_campaign,
		bot,
		bot_reason) VALUES %s`, strings.Join(values, ","))

	if _, err := ch.Exec(query, args...); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("saved requests", "count", len(requests))
	}

	return nil
}

func (ch *ClickHouse) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
