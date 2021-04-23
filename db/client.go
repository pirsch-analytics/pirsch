package db

import (
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
	"github.com/pirsch-analytics/pirsch/model"
	"strings"
)

// Client is a ClickHouse database client.
type Client struct {
	sqlx.DB
}

// NewClient returns a new client for given database connection string.
func NewClient(connection string) (*Client, error) {
	c, err := sqlx.Open("clickhouse", connection)

	if err != nil {
		return nil, err
	}

	if err := c.Ping(); err != nil {
		return nil, err
	}

	return &Client{
		*c,
	}, nil
}

// SaveHits saves new hits.
func (client *Client) SaveHits(hits []model.Hit) error {
	const hitParams = 21
	args := make([]interface{}, 0, len(hits)*hitParams)
	var query strings.Builder
	query.WriteString(`INSERT INTO "hit" (tenant_id, fingerprint, time, session, user_agent,
		path, url, language, country_code, referrer, referrer_name, referrer_icon, os, os_version,
		browser, browser_version, desktop, mobile, screen_width, screen_height, screen_class) VALUES `)

	for i, hit := range hits {
		args = append(args, hit.TenantID)
		args = append(args, hit.Fingerprint)
		args = append(args, hit.Time)
		args = append(args, hit.Session)
		args = append(args, hit.UserAgent)
		args = append(args, hit.Path)
		args = append(args, hit.URL)
		args = append(args, hit.Language)
		args = append(args, hit.CountryCode)
		args = append(args, hit.Referrer)
		args = append(args, hit.ReferrerName)
		args = append(args, hit.ReferrerIcon)
		args = append(args, hit.OS)
		args = append(args, hit.OSVersion)
		args = append(args, hit.Browser)
		args = append(args, hit.BrowserVersion)
		args = append(args, client.boolean(hit.Desktop))
		args = append(args, client.boolean(hit.Mobile))
		args = append(args, hit.ScreenWidth)
		args = append(args, hit.ScreenHeight)
		args = append(args, hit.ScreenClass)
		index := i * hitParams
		query.WriteString(fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d),`,
			index+1, index+2, index+3, index+4, index+5, index+6, index+7, index+8, index+9, index+10, index+11, index+12, index+13, index+14, index+15, index+16, index+17, index+18, index+19, index+20, index+21))
	}

	queryStr := query.String()
	tx, err := client.Beginx()

	if err != nil {
		return err
	}

	if _, err := tx.Exec(queryStr[:len(queryStr)-1], args...); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
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
