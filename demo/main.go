package main

import (
	"database/sql"
	"github.com/emvi/pirsch"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	db := connectToDB()

	// Create a new Postgres store to save statistics and hits.
	store := pirsch.NewPostgresStore(db, nil)

	// Set up a default tracker with a salt.
	// This will buffer and store hits and generate sessions by default.
	tracker := pirsch.NewTracker(store, "salt", nil)

	// Create a new process and run it each day on midnight (UTC) to process the stored hits.
	// The processor also cleans up the hits.
	processor := pirsch.NewProcessor(store)
	pirsch.RunAtMidnight(func() {
		if err := processor.Process(); err != nil {
			panic(err)
		}
	})

	// Create a handler to serve traffic.
	// We prevent tracking resources by checking the path. So a file on /my-file.txt won't create a new hit
	// but all page calls will be tracked.
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			tracker.Hit(r, nil)
		}

		w.Write([]byte("<h1>Hello World!</h1>"))
	}))

	// And finally, start the server.
	// We don't flush hits on shutdown but you should add that in a real application by calling Tracker.Flush().
	log.Println("Starting server on port 8080...")
	http.ListenAndServe(":8080", nil)
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
