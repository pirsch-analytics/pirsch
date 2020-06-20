package pirsch

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// Hit represents a single data point/page view.
type Hit struct {
	ID          int64     `db:"id" json:"id"`
	Fingerprint string    `db:"fingerprint" json:"fingerprint"`
	Path        string    `db:"path" json:"path"`
	Query       string    `db:"query" json:"query"`
	Fragment    string    `db:"fragment" json:"fragment"`
	URL         string    `db:"url" json:"url"`
	Language    string    `db:"language" json:"language"`
	Browser     string    `db:"browser" json:"browser"`
	Ref         string    `db:"ref" json:"ref"`
	Time        time.Time `db:"time" json:"time"`
}

func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// SaveHit saves a request as a new data point.
func SaveHit(r *http.Request) {
	hit := Hit{
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

	// TODO use store to save
	log.Printf("%v", hit)
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
