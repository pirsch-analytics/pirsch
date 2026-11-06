package db

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Connect connects to the test database.
func Connect() *ClickHouse {
	cfg := &ClickHouseConfig{
		Hostnames:          []string{"127.0.0.1"},
		Port:               9000,
		Database:           "pirschtest",
		Password:           "default",
		SSLSkipVerify:      true,
		Debug:              true,
		MaxOpenConnections: 1,
		dev:                true,
	}

	if err := Migrate(cfg); err != nil {
		panic(err)
	}

	c, err := NewClickHouse(cfg)

	if err != nil {
		panic(err)
	}

	return c
}

// Disconnect disconnects from the test database.
func Disconnect(client *ClickHouse) {
	if err := client.Close(); err != nil {
		panic(err)
	}
}

// CleanupDB clears all test database tables.
func CleanupDB(t *testing.T, client *ClickHouse) {
	if !client.dev {
		panic("the client is not in dev mode")
	}

	tables := []string{
		"session_v7",
		"page_view_v7",
		"event_v7",
		"request_v7",
		"session",
		"page_view",
		"event",
		"request",
		"imported_browser",
		"imported_utm_campaign",
		"imported_city",
		"imported_country",
		"imported_device",
		"imported_entry_page",
		"imported_exit_page",
		"imported_language",
		"imported_utm_medium",
		"imported_os",
		"imported_page",
		"imported_referrer",
		"imported_region",
		"imported_utm_source",
		"imported_visitors",
	}

	for _, table := range tables {
		if err := client.Exec(context.Background(), fmt.Sprintf(`TRUNCATE TABLE IF EXISTS "%s"`, table)); err != nil {
			t.Fatalf("error truncating table '%s': %s", table, err)
		}
	}

	slog.Debug("Finished cleaning up test data")
}

// DropDB drops all test database tables.
func DropDB(t *testing.T, client *ClickHouse) {
	if !client.dev {
		panic("the client is not in dev mode")
	}

	tables := []string{
		"session_v7",
		"page_view_v7",
		"event_v7",
		"request_v7",
		"session",
		"page_view",
		"event",
		"request",
		"schema_migrations",
		"imported_browser",
		"imported_utm_campaign",
		"imported_city",
		"imported_country",
		"imported_device",
		"imported_entry_page",
		"imported_exit_page",
		"imported_language",
		"imported_utm_medium",
		"imported_os",
		"imported_page",
		"imported_referrer",
		"imported_region",
		"imported_utm_source",
		"imported_visitors",
	}

	for _, table := range tables {
		err := client.Exec(context.Background(), fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, table))
		assert.NoError(t, err)
	}

	slog.Debug("Finished dropping test data")
}
