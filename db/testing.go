package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// Connect connects to the database.
func Connect() *Client {
	dbConfig := &ClientConfig{
		Hostname:           "127.0.0.1",
		Port:               9000,
		Database:           "pirschtest",
		SSLSkipVerify:      true,
		Debug:              false,
		MaxOpenConnections: 1,
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
	_, err := client.Exec(`ALTER TABLE "page_view" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	_, err = client.Exec(`ALTER TABLE "session" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	_, err = client.Exec(`ALTER TABLE "event" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	_, err = client.Exec(`ALTER TABLE "user_agent" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 20)
}

// DropDB drops all database tables.
func DropDB(t *testing.T, client *Client) {
	_, err := client.Exec(`DROP TABLE IF EXISTS "page_view"`)
	assert.NoError(t, err)
	_, err = client.Exec(`DROP TABLE IF EXISTS "session"`)
	assert.NoError(t, err)
	_, err = client.Exec(`DROP TABLE IF EXISTS "event"`)
	assert.NoError(t, err)
	_, err = client.Exec(`DROP TABLE IF EXISTS "user_agent"`)
	assert.NoError(t, err)
	_, err = client.Exec(`DROP TABLE IF EXISTS "schema_migrations"`)
	assert.NoError(t, err)
}
