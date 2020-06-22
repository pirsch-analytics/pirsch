package pirsch

import (
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"testing"
)

var (
	db *sql.DB
)

func TestMain(m *testing.M) {
	connectDB()
	os.Exit(m.Run())
}

func connectDB() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=pirsch sslmode=disable")

	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}
}

func cleanupDB(t *testing.T) {
	if _, err := db.Exec(`DELETE FROM "hit"`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`DELETE FROM "visitors_per_day"`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`DELETE FROM "visitors_per_hour"`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`DELETE FROM "visitors_per_language"`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`DELETE FROM "visitors_per_page"`); err != nil {
		t.Fatal(err)
	}
}
