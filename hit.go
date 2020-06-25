package pirsch

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Hit represents a single data point/page visit.
type Hit struct {
	ID          int64     `db:"id" json:"id"`
	Fingerprint string    `db:"fingerprint" json:"fingerprint"`
	Path        string    `db:"path" json:"path,omitempty"`
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
	now := time.Now().UTC() // capture first to get as close as possible
	return Hit{
		Fingerprint: Fingerprint(r),
		Path:        r.URL.Path,
		URL:         r.URL.String(),
		Language:    getLanguage(r),
		UserAgent:   r.UserAgent(),
		Ref:         r.Header.Get("Referer"),
		Time:        now,
	}
}

func getLanguage(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")

	if lang != "" {
		langs := strings.Split(lang, ";")
		parts := strings.Split(langs[0], ",")
		return parts[0]
	}

	return ""
}
