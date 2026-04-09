package db

import (
	"os"
	"testing"
)

var dbClient *ClickHouse

func TestMain(m *testing.M) {
	dbClient = Connect()
	defer Disconnect(dbClient)
	os.Exit(m.Run())
}
