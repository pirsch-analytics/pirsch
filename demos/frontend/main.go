package main

import (
	"database/sql"
	"github.com/emvi/pirsch"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
)

// For more details, take a look at the backend demo and documentation.
func main() {
	copyPirschJs()
	db := connectToDB()
	store := pirsch.NewPostgresStore(db, nil)
	tracker := pirsch.NewTracker(store, "salt", nil)

	// Create an endpoint to handle client tracking requests.
	// HitOptionsFromRequest is a utility function to process the required parameters.
	// You might want to additional checks, like for the tenant ID.
	http.Handle("/count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracker.Hit(r, pirsch.HitOptionsFromRequest(r))
		log.Println("Counted one hit")
	}))

	// Add a handler to serve index.html and pirsch.js.
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Println("Starting server on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func copyPirschJs() {
	content, err := ioutil.ReadFile("../../pirsch.js")

	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile("pirsch.js", content, 0755); err != nil {
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
