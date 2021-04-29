package main

/*
import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pirsch-analytics/pirsch"
	"log"
	"net/http"
	"os"
)

// For more details, take a look at the backend demo and documentation.
func main() {
	copyPirschJs()
	db := connectToDB()
	store := pirsch.NewPostgresStore(db, nil)
	tracker := pirsch.NewTracker(store, "salt", nil)

	// Create an endpoint to handle client tracking requests.
	// HitOptionsFromRequest is a utility function to process the required parameters.
	// You might want to additional checks, like for the client ID.
	http.Handle("/count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We don't need to call Hit in a new goroutine, as this is the only call in the handler
		// (running in its own goroutine already).
		tracker.Hit(r, pirsch.HitOptionsFromRequest(r))
		log.Println("Counted one hit")
	}))

	// Add a handler to serve index.html and pirsch.js.
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Println("Starting server on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func copyPirschJs() {
	content, err := os.ReadFile("../../js/pirsch.js")

	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("pirsch.js", content, 0755); err != nil {
		panic(err)
	}
}

func connectToDB() *sql.DB {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=pirsch sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
*/
