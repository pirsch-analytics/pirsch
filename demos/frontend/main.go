package main

import (
	"encoding/json"
	_ "github.com/lib/pq"
	"github.com/pirsch-analytics/pirsch/v5/db"
	"github.com/pirsch-analytics/pirsch/v5/tracker"
	"github.com/pirsch-analytics/pirsch/v5/tracker/session"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// For more details, take a look at the backend demo and documentation.
func main() {
	compileJs()
	copyPirschJs()
	copyPirschEventsJs()
	copyPirschSessionsJs()

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

	pirschTracker := tracker.NewTracker(tracker.Config{
		Store:           store,
		SessionCache:    session.NewMemCache(store, 100),
		FingerprintKey0: 42,
		FingerprintKey1: 123,
	})

	// Create an endpoint to handle client tracking requests.
	// HitOptionsFromRequest is a utility function to process the required parameters.
	// You might want to additional checks, like for the client ID.
	http.Handle("/count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We don't need to call PageView in a new goroutine, as this is the only call in the handler
		// (running in its own goroutine already).
		pirschTracker.PageView(r, 0, tracker.OptionsFromRequest(r))
		log.Println("Counted page view")
	}))

	// Create an endpoint to handle client event requests.
	http.Handle("/event", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name         string            `json:"event_name"`
			Duration     uint32            `json:"event_duration"`
			Meta         map[string]string `json:"event_meta"`
			URL          string            `json:"url"`
			Title        string            `json:"title"`
			Referrer     string            `json:"referrer"`
			ScreenWidth  uint16            `json:"screen_width"`
			ScreenHeight uint16            `json:"screen_height"`
		}
		decoder := json.NewDecoder(r.Body)

		if err := decoder.Decode(&req); err != nil {
			log.Printf("Error decoding event request: %s", err)
			return
		}

		pirschTracker.Event(r, 0, tracker.EventOptions{
			Name:     req.Name,
			Duration: req.Duration,
			Meta:     req.Meta,
		}, tracker.Options{
			URL:          req.URL,
			Title:        req.Title,
			Referrer:     req.Referrer,
			ScreenWidth:  req.ScreenWidth,
			ScreenHeight: req.ScreenHeight,
		})
		log.Println("Received event")
	}))

	// Create an endpoint to handle session requests.
	http.Handle("/session", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pirschTracker.ExtendSession(r, 0, tracker.OptionsFromRequest(r))
		log.Println("Kept session alive")
	}))

	// Add a handler to serve index.html and pirsch.js.
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Println("Started server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func compileJs() {
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = "../../js"

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func copyPirschJs() {
	content, err := os.ReadFile("../../js/pirsch.min.js")

	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("pirsch.min.js", content, 0755); err != nil {
		panic(err)
	}
}

func copyPirschEventsJs() {
	content, err := os.ReadFile("../../js/pirsch-events.min.js")

	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("pirsch-events.min.js", content, 0755); err != nil {
		panic(err)
	}
}

func copyPirschSessionsJs() {
	content, err := os.ReadFile("../../js/pirsch-sessions.min.js")

	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("pirsch-sessions.min.js", content, 0755); err != nil {
		panic(err)
	}
}
