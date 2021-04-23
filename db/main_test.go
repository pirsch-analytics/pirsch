package db

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if err := Migrate("clickhouse://127.0.0.1:9000?debug=true"); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func cleanupDB(client *Client) {
	client.MustExec(`ALTER TABLE "hit" DELETE WHERE 1=1`)
}
