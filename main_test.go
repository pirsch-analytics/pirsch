package pirsch

import (
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

func cleanupDB() {
	dbClient.MustExec(`ALTER TABLE "page_view" DELETE WHERE 1=1`)
	dbClient.MustExec(`ALTER TABLE "session" DELETE WHERE 1=1`)
	dbClient.MustExec(`ALTER TABLE "event" DELETE WHERE 1=1`)
	dbClient.MustExec(`ALTER TABLE "user_agent" DELETE WHERE 1=1`)
	time.Sleep(time.Millisecond * 20)
}

func dropDB() {
	dbClient.MustExec(`DROP TABLE IF EXISTS "page_view"`)
	dbClient.MustExec(`DROP TABLE IF EXISTS "session"`)
	dbClient.MustExec(`DROP TABLE IF EXISTS "event"`)
	dbClient.MustExec(`DROP TABLE IF EXISTS "user_agent"`)
	dbClient.MustExec(`DROP TABLE IF EXISTS "schema_migrations"`)
}
