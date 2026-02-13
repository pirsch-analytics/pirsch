package tracker

import (
	"os"
	"testing"

	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
)

var dbClient *db.Client

func TestMain(m *testing.M) {
	dbClient = db.Connect()
	defer db.Disconnect(dbClient)
	os.Exit(m.Run())
}
