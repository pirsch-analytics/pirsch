package db

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

const (
	defaultMaxOpenConnections    = 20
	defaultMaxConnectionLifetime = 1800
	defaultMaxIdleConnections    = 5
)

// ClickHouse implements the Storage interface.
type ClickHouse struct {
	clickhouse.Conn

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

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: addr,
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		TLS:             tlsConn,
		DialTimeout:     time.Second * 30,
		Logger:          logger,
		MaxOpenConns:    config.MaxOpenConnections,
		MaxIdleConns:    config.MaxIdleConnections,
		ConnMaxLifetime: time.Duration(config.MaxConnectionLifetimeSeconds) * time.Second,
	})

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	return &ClickHouse{
		conn,
		config.Logger,
		config.Debug,
		config.dev,
	}, nil
}

// SaveSessions implements the Storage interface.
func (ch *ClickHouse) SaveSessions(ctx context.Context, sessions []model.Session) error {
	stmt, err := ch.PrepareBatch(ctx, `INSERT INTO "session_v7" (sign,
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
		extended)`)

	if err != nil {
		return err
	}

	for _, session := range sessions {
		if err := stmt.Append(session.Sign,
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
			session.Extended); err != nil {
			return err
		}
	}

	if err := stmt.Send(); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("sessions saved", "count", len(sessions))
	}

	return nil
}

// SavePageViews implements the Storage interface.
func (ch *ClickHouse) SavePageViews(ctx context.Context, pageViews []model.PageView) error {
	stmt, err := ch.PrepareBatch(ctx, `INSERT INTO "page_view_v7" (client_id,
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
		tags)`)

	if err != nil {
		return err
	}

	for _, pageView := range pageViews {
		if err := stmt.Append(pageView.ClientID,
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
			pageView.Tags); err != nil {
			return err
		}
	}

	if err := stmt.Send(); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("page views saved", "count", len(pageViews))
	}

	return nil
}

// SaveEvents implements the Storage interface.
func (ch *ClickHouse) SaveEvents(ctx context.Context, events []model.Event) error {
	stmt, err := ch.PrepareBatch(ctx, `INSERT INTO "event_v7" (client_id,
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
		channel)`)

	if err != nil {
		return err
	}

	for _, event := range events {
		if err := stmt.Append(event.ClientID,
			event.VisitorID,
			event.Time.UnixMilli(),
			event.SessionID,
			event.Name,
			ch.json(event.MetaData),
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
			event.Channel); err != nil {
			return err
		}
	}

	if err := stmt.Send(); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("events saved", "count", len(events))
	}

	return nil
}

// SaveRequests implements the Storage interface.
func (ch *ClickHouse) SaveRequests(ctx context.Context, requests []model.Request) error {
	stmt, err := ch.PrepareBatch(ctx, `INSERT INTO "request" (client_id,
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
		bot_reason)`)

	if err != nil {
		return err
	}

	for _, req := range requests {
		if err := stmt.Append(req.ClientID,
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
			req.BotReason); err != nil {
			return err
		}
	}

	if err := stmt.Send(); err != nil {
		return err
	}

	if ch.debug {
		ch.logger.Debug("requests saved", "count", len(requests))
	}

	return nil
}

// Session implements the Storage interface.
func (ch *ClickHouse) Session(ctx context.Context, clientID, fingerprint uint64, maxAge time.Time) (*model.Session, error) {
	query := `SELECT sign,
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
		extended
		FROM "session_v7"
		WHERE client_id = ?
		AND visitor_id = ?
		AND time > ?
		ORDER BY time DESC
		LIMIT 1`
	session := new(model.Session)
	err := ch.QueryRow(ctx, query, clientID, fingerprint, maxAge).Scan(&session.Sign,
		&session.Version,
		&session.ClientID,
		&session.VisitorID,
		&session.SessionID,
		&session.Time,
		&session.Start,
		&session.Hostname,
		&session.EntryPath,
		&session.ExitPath,
		&session.PageViews,
		&session.IsBounce,
		&session.EntryTitle,
		&session.ExitTitle,
		&session.Language,
		&session.CountryCode,
		&session.Region,
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
		&session.ScreenClass,
		&session.UTMSource,
		&session.UTMMedium,
		&session.UTMCampaign,
		&session.UTMContent,
		&session.UTMTerm,
		&session.Channel,
		&session.Extended)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		if !errors.Is(err, context.Canceled) {
			ch.logger.Error("error reading session", "err", err)
		}

		return nil, err
	}

	return session, nil
}

func (ch *ClickHouse) json(s any) []byte {
	o, _ := json.Marshal(s)
	return o
}

func (ch *ClickHouse) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
