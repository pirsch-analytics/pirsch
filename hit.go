package pirsch

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	hitsBufferSize      = 200
	hitsBufferThreshold = 100 // save at 50% full
	hitsBufferTimeout   = time.Second * 15
)

var (
	m         sync.Mutex
	hits      = make(chan Hit, hitsBufferSize)
	lastSave  = time.Now()
	saveTimer *time.Timer
)

// Hit represents a single data point/page view.
type Hit struct {
	ID          int64     `db:"id" json:"id"`
	Fingerprint string    `db:"fingerprint" json:"fingerprint"`
	Path        string    `db:"path" json:"path,omitempty"`
	Query       string    `db:"query" json:"query,omitempty"`
	Fragment    string    `db:"fragment" json:"fragment,omitempty"`
	URL         string    `db:"url" json:"url,omitempty"`
	Language    string    `db:"language" json:"language,omitempty"`
	Browser     string    `db:"browser" json:"browser,omitempty"`
	Ref         string    `db:"ref" json:"ref,omitempty"`
	Time        time.Time `db:"time" json:"time"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// SaveHit saves a request as a new data point.
// It's recommended to call this function within its own goroutine.
func SaveHit(r *http.Request) {
	hits <- Hit{
		Fingerprint: Fingerprint(r),
		Path:        r.URL.Path,
		Query:       r.URL.RawQuery,
		Fragment:    r.URL.Fragment,
		URL:         r.URL.String(),
		Language:    getLanguage(r),
		Browser:     r.UserAgent(),
		Ref:         r.Header.Get("Referer"),
		Time:        time.Now(),
	}
	saveInStore()
}

func saveInStore() {
	if len(hits) > hitsBufferSize-hitsBufferThreshold || (len(hits) > 0 && lastSave.Before(time.Now().Add(-hitsBufferTimeout))) {
		m.Lock()
		defer m.Unlock()
		buffer := make([]Hit, 0, len(hits))

		for i := 0; i < len(hits); i++ {
			buffer = append(buffer, <-hits)
		}

		lastSave = time.Now()

		if saveTimer != nil {
			saveTimer.Stop()
		}

		saveTimer = nil
		go store.Save(buffer)
	} else if saveTimer == nil {
		m.Lock()
		defer m.Unlock()
		saveTimer = time.AfterFunc(hitsBufferTimeout, saveInStore)
	}
}

func getLanguage(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")

	if lang != "" {
		langs := strings.Split(lang, ";")
		parts := strings.Split(langs[0], ",")

		if len(parts) > 1 {
			return parts[1]
		}

		return parts[0]
	}

	return ""
}
