package db

import (
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
)

// Client is a ClickHouse database client.
type Client struct {
	DB *sqlx.DB
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
		DB: c,
	}, nil
}
