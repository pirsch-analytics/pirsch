package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

// Connect connects to the database.
func Connect() *Client {
	dbConfig := &ClientConfig{
		Hostnames:          []string{"127.0.0.1"},
		Port:               9000,
		Database:           "pirschtest",
		SSLSkipVerify:      true,
		Debug:              false,
		MaxOpenConnections: 1,
		dev:                true,
	}

	if err := Migrate(dbConfig); err != nil {
		panic(err)
	}

	c, err := NewClient(dbConfig)

	if err != nil {
		panic(err)
	}

	return c
}

// Disconnect disconnects from the database.
func Disconnect(client *Client) {
	if err := client.DB.Close(); err != nil {
		panic(err)
	}
}

// CleanupDB clears all database tables.
func CleanupDB(t *testing.T, client *Client) {
	if !client.dev {
		panic("client not in dev mode")
	}

	tables := []string{
		"page_view",
		"session",
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
			_, err := client.Exec(fmt.Sprintf(`ALTER TABLE "%s" DELETE WHERE 1=1`, table))
			wg.Done()
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	for _, table := range tables {
		count := 1

		for count > 0 {
			row := client.QueryRow(fmt.Sprintf(`SELECT count(*) FROM "%s"`, table))
			assert.NoError(t, row.Scan(&count))
		}
	}
}

// DropDB drops all database tables.
func DropDB(t *testing.T, client *Client) {
	if !client.dev {
		panic("client not in dev mode")
	}

	tables := []string{
		"page_view",
		"session",
		"event",
		"event_new",
		"event_backup",
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
		_, err := client.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, table))
		assert.NoError(t, err)
	}
}
