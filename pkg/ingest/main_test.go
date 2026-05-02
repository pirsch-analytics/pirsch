package ingest

import (
	"os"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
)

var dbClient *db.ClickHouse

func TestMain(m *testing.M) {
	dbClient = db.Connect()
	defer db.Disconnect(dbClient)
	os.Exit(m.Run())
}
