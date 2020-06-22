package pirsch

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
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
	UserAgent   string    `db:"user_agent" json:"user_agent,omitempty"`
	Ref         string    `db:"ref" json:"ref,omitempty"`
	Time        time.Time `db:"time" json:"time"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

func hitFromRequest(r *http.Request) Hit {
	return Hit{
		Fingerprint: Fingerprint(r),
		Path:        r.URL.Path,
		Query:       r.URL.RawQuery,
		Fragment:    r.URL.Fragment,
		URL:         r.URL.String(),
		Language:    getLanguage(r),
		UserAgent:   r.UserAgent(),
		Ref:         r.Header.Get("Referer"),
		Time:        time.Now(),
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
