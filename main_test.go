package pirsch

import (
	"database/sql"
	_ "github.com/lib/pq"
	"os"
	"testing"
)

var (
	postgresDB *sql.DB
)

func TestMain(m *testing.M) {
	connectDB()
	defer closeDB()
	os.Exit(m.Run())
}

// open test database connections
func connectDB() {
	connectPostgresDB()
}

// close test database connections
func closeDB() {
	closePostgresDB()
}

// clean up all test databases
func cleanupDB(t *testing.T) {
	cleanupPostgresDB(t)
}

func connectPostgresDB() {
	var err error
	postgresDB, err = sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=pirsch search_path=public sslmode=disable timezone=UTC")

	if err != nil {
		panic(err)
	}

	if err := postgresDB.Ping(); err != nil {
		panic(err)
	}

	postgresDB.SetMaxOpenConns(1)
}

func closePostgresDB() {
	if err := postgresDB.Close(); err != nil {
		panic(err)
	}
}

func cleanupPostgresDB(t *testing.T) {
	if _, err := postgresDB.Exec(`DELETE FROM "hit"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitors_per_day"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitors_per_hour"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitors_per_language"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitors_per_page"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitors_per_referrer"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitors_per_os"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitors_per_browser"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitor_platform"`); err != nil {
		t.Fatal(err)
	}
}
