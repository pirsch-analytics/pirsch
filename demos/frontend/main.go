package main

import (
	"encoding/json"
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
	copyPirschEventsJs()

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

	// Create an endpoint to handle client event requests.
	http.Handle("/event", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string            `json:"event_name"`
			Duration uint32            `json:"event_duration"`
			Meta     map[string]string `json:"event_meta"`
		}
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&req); err != nil {
			log.Printf("Error decoding event request: %s", err)
			return
		}

		data := tracker.EventOptions{
			Name:     req.Name,
			Duration: req.Duration,
			Meta:     req.Meta,
		}

		pirschTracker.Event(r, data, tracker.HitOptionsFromRequest(r))
		log.Println("Received event")
	}))

	// Add a handler to serve index.html and pirsch.js.
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Println("Started server on http://localhost:8080")
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

func copyPirschEventsJs() {
	content, err := os.ReadFile("../../js/pirsch-events.js")

	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("pirsch-events.js", content, 0755); err != nil {
		panic(err)
	}
}
