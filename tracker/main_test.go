package tracker

import (
	"github.com/pirsch-analytics/pirsch/v5/db"
	"os"
	"testing"
)

var dbClient *db.Client

func TestMain(m *testing.M) {
	dbClient = db.Connect()
	defer db.Disconnect(dbClient)
	os.Exit(m.Run())
}
