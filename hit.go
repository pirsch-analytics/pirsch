package pirsch

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Hit represents a single data point/page visit.
type Hit struct {
	ID          int64         `db:"id" json:"id"`
	TenantID    sql.NullInt64 `db:"tenant_id" json:"tenant_id"`
	Fingerprint string        `db:"fingerprint" json:"fingerprint"`
	Path        string        `db:"path" json:"path,omitempty"`
	URL         string        `db:"url" json:"url,omitempty"`
	Language    string        `db:"language" json:"language,omitempty"`
	UserAgent   string        `db:"user_agent" json:"user_agent,omitempty"`
	Ref         string        `db:"ref" json:"ref,omitempty"`
	Time        time.Time     `db:"time" json:"time"`
}

// HitOptions is used to manipulate the data saved on a hit.
type HitOptions struct {
	// TenantID is optionally saved with a hit to split the data between multiple tenants.
	TenantID sql.NullInt64

	// Path can be specified to manually overwrite the path stored for the request.
	// This will also affect the URL.
	Path string
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// HitFromRequest returns a new Hit for given request, salt and HitOptions.
// The salt must stay consistent to track visitors across multiple calls.
// The easiest way to track visitors is to use the Tracker.
func HitFromRequest(r *http.Request, salt string, options *HitOptions) Hit {
	now := time.Now().UTC() // capture first to get as close as possible

	// set default options in case they're nil
	if options == nil {
		options = &HitOptions{}
	}

	// manually overwrite path if set
	requestURL := r.URL.String()

	if options.Path != "" {
		u, err := url.Parse(r.RequestURI)

		if err == nil {
			// change path and re-assemble URL
			u.Path = options.Path
			requestURL = u.String()
		}
	} else {
		options.Path = r.URL.Path
	}

	return Hit{
		TenantID:    options.TenantID,
		Fingerprint: Fingerprint(r, salt),
		Path:        options.Path,
		URL:         requestURL,
		Language:    getLanguage(r),
		UserAgent:   r.UserAgent(),
		Ref:         r.Header.Get("Referer"),
		Time:        now,
	}
}

// IgnoreHit returns true, if a hit should be ignored for given request, or false otherwise.
// The easiest way to track visitors is to use the Tracker.
func IgnoreHit(r *http.Request) bool {
	userAgent := strings.TrimSpace(strings.ToLower(r.Header.Get("User-Agent")))

	if userAgent == "" {
		return true
	}

	xPurpose := r.Header.Get("X-Purpose")
	purpose := r.Header.Get("Purpose")

	if r.Header.Get("X-Moz") == "prefetch" || // ignore browsers pre-fetching data
		xPurpose == "prefetch" ||
		xPurpose == "preview" ||
		purpose == "prefetch" ||
		purpose == "preview" {
		return true
	}

	for _, botUserAgent := range userAgentBlacklist {
		if strings.Contains(userAgent, botUserAgent) {
			return true
		}
	}

	return false
}

func getLanguage(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")

	if lang != "" {
		langs := strings.Split(lang, ";")
		parts := strings.Split(langs[0], ",")
		return strings.ToLower(parts[0])
	}

	return ""
}
