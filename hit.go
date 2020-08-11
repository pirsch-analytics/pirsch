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

	// RefererDomainBlacklist is used to filter out unwanted referers from the Referer header.
	// This can be used to filter out traffic from your own site or subdomains.
	// To filter your own domain and subdomains, add your domain to the list and set RefererDomainBlacklistIncludesSubdomains to true.
	// This way the referer for blog.mypage.com -> mypage.com won't be saved.
	RefererDomainBlacklist []string

	// RefererDomainBlacklistIncludesSubdomains set to true to include all subdomains in the RefererDomainBlacklist,
	// or else subdomains must explicitly be included in the blacklist.
	// If the blacklist contains domain.com, sub.domain.com and domain.com will be treated as equally.
	RefererDomainBlacklistIncludesSubdomains bool
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
		Ref:         getReferer(r, options.RefererDomainBlacklist, options.RefererDomainBlacklistIncludesSubdomains),
		Time:        now,
	}
}

// IgnoreHit returns true, if a hit should be ignored for given request, or false otherwise.
// The easiest way to track visitors is to use the Tracker.
func IgnoreHit(r *http.Request) bool {
	// empty User-Agents are usually bots
	userAgent := strings.TrimSpace(strings.ToLower(r.Header.Get("User-Agent")))

	if userAgent == "" {
		return true
	}

	// ignore browsers pre-fetching data
	xPurpose := r.Header.Get("X-Purpose")
	purpose := r.Header.Get("Purpose")

	if r.Header.Get("X-Moz") == "prefetch" ||
		xPurpose == "prefetch" ||
		xPurpose == "preview" ||
		purpose == "prefetch" ||
		purpose == "preview" {
		return true
	}

	// filter for bot keywords
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

func getReferer(r *http.Request, domainBlacklist []string, ignoreSubdomain bool) string {
	referer := r.Header.Get("Referer")

	if referer == "" {
		return ""
	}

	u, err := url.Parse(referer)

	if err != nil {
		return ""
	}

	hostname := u.Hostname()

	if ignoreSubdomain {
		hostname = stripSubdomain(hostname)
	}

	if containsString(domainBlacklist, hostname) {
		return ""
	}

	return referer
}

func stripSubdomain(hostname string) string {
	if hostname == "" {
		return ""
	}

	runes := []rune(hostname)
	index := len(runes) - 1
	dots := 0

	for i := index; i > 0; i-- {
		if runes[i] == '.' {
			dots++

			if dots == 2 {
				index++
				break
			}
		}

		index--
	}

	return hostname[index:]
}
