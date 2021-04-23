package hit

import (
	"github.com/pirsch-analytics/pirsch/db"
	"os"
	"testing"
)

var dbClient *db.Client

func TestMain(m *testing.M) {
	if err := db.Migrate("clickhouse://127.0.0.1:9000?debug=true"); err != nil {
		panic(err)
	}

	c, err := db.NewClient("tcp://127.0.0.1:9000?debug=true")

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
