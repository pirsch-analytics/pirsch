package pirsch

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var dbClient *Client

func TestMain(m *testing.M) {
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

	dbClient = c
	defer func() {
		if err := dbClient.DB.Close(); err != nil {
			panic(err)
		}
	}()
	os.Exit(m.Run())
}

func cleanupDB(t *testing.T) {
	_, err := dbClient.Exec(`ALTER TABLE "page_view" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	_, err = dbClient.Exec(`ALTER TABLE "session" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	_, err = dbClient.Exec(`ALTER TABLE "event" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	_, err = dbClient.Exec(`ALTER TABLE "user_agent" DELETE WHERE 1=1`)
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 20)
}

func dropDB(t *testing.T) {
	_, err := dbClient.Exec(`DROP TABLE IF EXISTS "page_view"`)
	assert.NoError(t, err)
	_, err = dbClient.Exec(`DROP TABLE IF EXISTS "session"`)
	assert.NoError(t, err)
	_, err = dbClient.Exec(`DROP TABLE IF EXISTS "event"`)
	assert.NoError(t, err)
	_, err = dbClient.Exec(`DROP TABLE IF EXISTS "user_agent"`)
	assert.NoError(t, err)
	_, err = dbClient.Exec(`DROP TABLE IF EXISTS "schema_migrations"`)
	assert.NoError(t, err)
}
