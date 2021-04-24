package pirsch

import (
	"os"
	"testing"
)

var dbClient *Client

func TestMain(m *testing.M) {
	if err := Migrate("clickhouse://127.0.0.1:9000?x-multi-statement=true"); err != nil {
		panic(err)
	}

	c, err := NewClient("tcp://127.0.0.1:9000", nil)

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
	dbClient.MustExec(`ALTER TABLE "hit" DELETE WHERE 1=1`)
}
