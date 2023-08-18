package analyzer

import (
	db2 "github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"os"
	"testing"
)

var dbClient *db2.Client

func TestMain(m *testing.M) {
	dbClient = db2.Connect()
	defer db2.Disconnect(dbClient)
	os.Exit(m.Run())
}
