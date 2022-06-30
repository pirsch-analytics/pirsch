package main

import (
	_ "github.com/lib/pq"
	"github.com/pirsch-analytics/pirsch/v3"
	"log"
	"net/http"
)

func main() {
	// Set the key for SipHash.
	pirsch.SetFingerprintKeys(42, 123)

	db := &pirsch.ClientConfig{
		Hostname:      "127.0.0.1",
		Port:          9000,
		Database:      "pirschtest",
		SSLSkipVerify: true,
		Debug:         false,
	}

	if err := pirsch.Migrate(db); err != nil {
		panic(err)
	}

	// Create a new ClickHouse client to save hits.
	store, err := pirsch.NewClient(db)

	if err != nil {
		panic(err)
	}

	// Set up a default tracker with a salt.
	// This will buffer and store hits and generate sessions by default.
	tracker := pirsch.NewTracker(store, "salt", &pirsch.TrackerConfig{
		SessionCache: pirsch.NewSessionCacheMem(store, 100),
	})

	// Create a handler to serve traffic.
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// You can make sure only page calls get counted by checking the path: if r.URL.Path == "/" {...
		if r.URL.Path != "/favicon.ico" {
			go tracker.Hit(r, nil)
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
