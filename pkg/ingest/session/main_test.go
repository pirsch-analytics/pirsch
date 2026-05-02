package session

import (
	"os"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
)

var client *db.ClickHouse

func TestMain(m *testing.M) {
	client = db.Connect()
	defer db.Disconnect(client)
	os.Exit(m.Run())
}
