package db

import (
	"context"
	"fmt"
	"sync"
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
		Debug:              false,
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
	var wg sync.WaitGroup
	wg.Add(len(tables))

	for _, table := range tables {
		go func() {
			err := client.Exec(context.Background(), fmt.Sprintf(`ALTER TABLE "%s" DELETE WHERE 1=1`, table))
			wg.Done()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	for _, table := range tables {
		count := uint64(1)

		for count > 0 {
			row := client.QueryRow(context.Background(), fmt.Sprintf(`SELECT count(*) FROM "%s"`, table))
			assert.NoError(t, row.Scan(&count))
		}
	}
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
}
