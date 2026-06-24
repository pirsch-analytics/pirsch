package db

import (
	"os"
	"testing"
)

var client *ClickHouse

func TestMain(m *testing.M) {
	client = Connect()
	defer Disconnect(client)
	os.Exit(m.Run())
}
