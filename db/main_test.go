package db

import (
	"os"
	"testing"
)

var dbClient *Client

func TestMain(m *testing.M) {
	dbClient = Connect()
	defer Disconnect(dbClient)
	os.Exit(m.Run())
}
