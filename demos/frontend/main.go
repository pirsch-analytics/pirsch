package main

import (
	_ "github.com/lib/pq"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/tracker"
	"github.com/pirsch-analytics/pirsch/v4/tracker/session"
	"log"
	"net/http"
	"os"
)

// For more details, take a look at the backend demo and documentation.
func main() {
	copyPirschJs()

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

	store, err := db.NewClient(dbConfig)

	if err != nil {
		panic(err)
	}

	pirschTracker := tracker.NewTracker(store, "salt", &tracker.Config{
		SessionCache: session.NewMemCache(store, 100),
	})

	// Create an endpoint to handle client tracking requests.
	// HitOptionsFromRequest is a utility function to process the required parameters.
	// You might want to additional checks, like for the client ID.
	http.Handle("/count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We don't need to call Hit in a new goroutine, as this is the only call in the handler
		// (running in its own goroutine already).
		pirschTracker.Hit(r, tracker.HitOptionsFromRequest(r))
		log.Println("Counted one hit")
	}))

	// Add a handler to serve index.html and pirsch.js.
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
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
