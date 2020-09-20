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

	if _, err := postgresDB.Exec(`DELETE FROM "visitor_stats"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "visitor_time_stats"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "language_stats"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "referrer_stats"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "os_stats"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "browser_stats"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "screen_stats"`); err != nil {
		t.Fatal(err)
	}

	if _, err := postgresDB.Exec(`DELETE FROM "country_stats"`); err != nil {
		t.Fatal(err)
	}
}
