package main

import (
	_ "github.com/lib/pq"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/tracker_"
	"github.com/pirsch-analytics/pirsch/v4/tracker_/session"
	"log"
	"net/http"
)

func main() {
	// Set the key for SipHash.
	tracker.SetFingerprintKeys(42, 123)

	dbConfig := &db.ClientConfig{
		Hostname:      "127.0.0.1",
		Port:          9000,
		Database:      "pirschtest",
		SSLSkipVerify: true,
		Debug:         false,
	}

	if err := db.Migrate(dbConfig); err != nil {
		panic(err)
	}

	// Create a new ClickHouse client to save hits.
	store, err := db.NewClient(dbConfig)

	if err != nil {
		panic(err)
	}

	// Set up a default tracker with a salt.
	// This will buffer and store hits and generate sessions by default.
	pirschTracker := tracker.NewTracker(store, "salt", &tracker.Config{
		SessionCache: session.NewMemCache(store, 100),
	})

	// Create a handler to serve traffic.
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// You can make sure only page calls get counted by checking the path: if r.URL.Path == "/" {...
		if r.URL.Path != "/favicon.ico" {
			go pirschTracker.Hit(r, nil)
		}

		// Send response.
		if _, err := w.Write([]byte("<h1>Hello World!</h1>")); err != nil {
			log.Fatal(err)
		}
	}))

	// And finally, start the server.
	// We don't flush hits on shutdown but you should add that in a real application by calling Tracker.Flush().
	log.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
