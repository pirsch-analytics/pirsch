package main

import (
	"database/sql"
	"github.com/emvi/pirsch"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=pirsch sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	// create a new tracker and processor using postgres as its data store
	store := pirsch.NewPostgresStore(db, nil)
	tracker := pirsch.NewTracker(store, "salt", nil)
	processor := pirsch.NewProcessor(store, nil)
	pirsch.RunAtMidnight(func() {
		if err := processor.Process(); err != nil {
			panic(err)
		}
	})

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// don't track resources, just the main page in this demo
		// in a real application we would split the handlers for pages and resource endpoints
		// and only track the page handlers
		if r.URL.Path == "/" {
			tracker.Hit(r, nil)
		}

		w.Write([]byte("<h1>Hello World!</h1>"))
	}))

	log.Println("Starting server...")
	http.ListenAndServe(":8080", nil)
}
