package pirsch

import (
	"database/sql"
	"encoding/json"
	iso6391 "github.com/emvi/iso-639-1"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var referrerQueryParams = []string{
	"ref",
	"referer",
	"referrer",
}

// Hit represents a single data point/page visit and is the central entity of pirsch.
type Hit struct {
	BaseEntity

	Fingerprint    string         `db:"fingerprint" json:"fingerprint"`
	Session        sql.NullTime   `db:"session" json:"session"`
	Path           sql.NullString `db:"path" json:"path,omitempty"`
	URL            sql.NullString `db:"url" json:"url,omitempty"`
	Language       sql.NullString `db:"language" json:"language,omitempty"`
	UserAgent      sql.NullString `db:"user_agent" json:"user_agent,omitempty"`
	Referrer       sql.NullString `db:"referrer" json:"referrer,omitempty"`
	OS             sql.NullString `db:"os" json:"os,omitempty"`
	OSVersion      sql.NullString `db:"os_version" json:"os_version,omitempty"`
	Browser        sql.NullString `db:"browser" json:"browser,omitempty"`
	BrowserVersion sql.NullString `db:"browser_version" json:"browser_version,omitempty"`
	CountryCode    sql.NullString `db:"country_code" json:"country_code"`
	Desktop        bool           `db:"desktop" json:"desktop"`
	Mobile         bool           `db:"mobile" json:"mobile"`
	ScreenWidth    int            `db:"screen_width" json:"screen_width"`
	ScreenHeight   int            `db:"screen_height" json:"screen_height"`
	Time           time.Time      `db:"time" json:"time"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// HitOptions is used to manipulate the data saved on a hit.
type HitOptions struct {
	// TenantID is optionally saved with a hit to split the data between multiple tenants.
	TenantID sql.NullInt64

	// URL can be set to manually overwrite the URL stored for this request.
	// This will also affect the Path, except it is set too.
	URL string

	// Path can be set to manually overwrite the path stored for the request.
	// This will also affect the URL.
	Path string

	// Referrer can be set to manually overwrite the referrer from the request.
	Referrer string

	// ReferrerDomainBlacklist is used to filter out unwanted referrer from the Referrer header.
	// This can be used to filter out traffic from your own site or subdomains.
	// To filter your own domain and subdomains, add your domain to the list and set ReferrerDomainBlacklistIncludesSubdomains to true.
	// This way the referrer for blog.mypage.com -> mypage.com won't be saved.
	ReferrerDomainBlacklist []string

	// ReferrerDomainBlacklistIncludesSubdomains set to true to include all subdomains in the ReferrerDomainBlacklist,
	// or else subdomains must explicitly be included in the blacklist.
	// If the blacklist contains domain.com, sub.domain.com and domain.com will be treated as equally.
	ReferrerDomainBlacklistIncludesSubdomains bool

	// ScreenWidth sets the screen width to be stored with the hit.
	ScreenWidth int

	// ScreenHeight sets the screen height to be stored with the hit.
	ScreenHeight int

	geoDB        *GeoDB
	sessionCache *sessionCache
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

	// shorten strings if required and parse User-Agent to extract more data (OS, Browser)
	getRequestURI(r, options)
	fingerprint := Fingerprint(r, salt)
	path := shortenString(options.Path, 2000)
	requestURL := shortenString(options.URL, 2000)
	ua := r.UserAgent()
	uaInfo := ParseUserAgent(ua)
	uaInfo.OS = shortenString(uaInfo.OS, 20)
	uaInfo.OSVersion = shortenString(uaInfo.OSVersion, 20)
	uaInfo.Browser = shortenString(uaInfo.Browser, 20)
	uaInfo.BrowserVersion = shortenString(uaInfo.BrowserVersion, 20)
	ua = shortenString(ua, 200)
	lang := shortenString(getLanguage(r), 10)
	referrer := shortenString(getReferrer(r, options.Referrer, options.ReferrerDomainBlacklist, options.ReferrerDomainBlacklistIncludesSubdomains), 200)
	countryCode := ""

	if options.geoDB != nil {
		countryCode = options.geoDB.CountryCode(getIP(r))
	}

	var session time.Time

	if options.sessionCache != nil {
		session = options.sessionCache.find(options.TenantID, fingerprint)
	}

	if options.ScreenWidth <= 0 || options.ScreenHeight <= 0 {
		options.ScreenWidth = 0
		options.ScreenHeight = 0
	}

	return Hit{
		BaseEntity:     BaseEntity{TenantID: options.TenantID},
		Fingerprint:    fingerprint,
		Session:        sql.NullTime{Time: session, Valid: !session.IsZero()},
		Path:           sql.NullString{String: path, Valid: path != ""},
		URL:            sql.NullString{String: requestURL, Valid: requestURL != ""},
		Language:       sql.NullString{String: lang, Valid: lang != ""},
		UserAgent:      sql.NullString{String: ua, Valid: ua != ""},
		Referrer:       sql.NullString{String: referrer, Valid: referrer != ""},
		OS:             sql.NullString{String: uaInfo.OS, Valid: uaInfo.OS != ""},
		OSVersion:      sql.NullString{String: uaInfo.OSVersion, Valid: uaInfo.OSVersion != ""},
		Browser:        sql.NullString{String: uaInfo.Browser, Valid: uaInfo.Browser != ""},
		BrowserVersion: sql.NullString{String: uaInfo.BrowserVersion, Valid: uaInfo.BrowserVersion != ""},
		CountryCode:    sql.NullString{String: countryCode, Valid: countryCode != ""},
		Desktop:        uaInfo.IsDesktop(),
		Mobile:         uaInfo.IsMobile(),
		ScreenWidth:    options.ScreenWidth,
		ScreenHeight:   options.ScreenHeight,
		Time:           now,
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

	// filter referrer spammers
	if ignoreReferrer(r) {
		return true
	}

	// filter for bot keywords (most expensive operation last)
	for _, botUserAgent := range userAgentBlacklist {
		if strings.Contains(userAgent, botUserAgent) {
			return true
		}
	}

	return false
}

// HitOptionsFromRequest returns the HitOptions for given client request.
// This function can be used to accept hits from pirsch.js. Invalid parameters are ignored and left empty.
// You might want to add additional checks before calling HitFromRequest afterwards (like for the HitOptions.TenantID).
func HitOptionsFromRequest(r *http.Request) *HitOptions {
	query := r.URL.Query()
	return &HitOptions{
		TenantID:     getNullInt64QueryParam(query.Get("tenantid")),
		URL:          getURLQueryParam(query.Get("url")),
		Referrer:     getURLQueryParam(query.Get("ref")),
		ScreenWidth:  getIntQueryParam(query.Get("w")),
		ScreenHeight: getIntQueryParam(query.Get("h")),
	}
}

func ignoreReferrer(r *http.Request) bool {
	referrer := getReferrerFromHeaderOrQuery(r)

	if referrer == "" {
		return false
	}

	u, err := url.ParseRequestURI(referrer)

	if err == nil {
		referrer = u.Hostname()
	}

	referrer = stripSubdomain(referrer)
	_, found := referrerBlacklist[referrer]
	return found
}

func getRequestURI(r *http.Request, options *HitOptions) {
	if options.URL == "" {
		options.URL = r.URL.String()
	}

	u, err := url.ParseRequestURI(options.URL)

	if err == nil {
		if options.Path != "" {
			// change path and re-assemble URL
			u.Path = options.Path
			options.URL = u.String()
		} else {
			options.Path = u.Path
		}
	}
}

func getLanguage(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")

	if lang != "" {
		langs := strings.Split(lang, ";")
		parts := strings.Split(langs[0], ",")
		parts = strings.Split(parts[0], "-")
		code := strings.ToLower(strings.TrimSpace(parts[0]))

		if iso6391.ValidCode(code) {
			return code
		}
	}

	return ""
}

func getReferrer(r *http.Request, ref string, domainBlacklist []string, ignoreSubdomain bool) string {
	referrer := ""

	if ref != "" {
		referrer = ref
	} else {
		referrer = getReferrerFromHeaderOrQuery(r)
	}

	if referrer == "" {
		return ""
	}

	u, err := url.ParseRequestURI(referrer)

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

	// remove query parameters and anchor
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func getReferrerFromHeaderOrQuery(r *http.Request) string {
	referrer := r.Header.Get("Referer")

	if referrer == "" {
		for _, param := range referrerQueryParams {
			referrer = r.URL.Query().Get(param)

			if referrer != "" {
				return referrer
			}
		}
	}

	return referrer
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

func shortenString(str string, n int) string {
	// we intentionally use len instead of utf8.RuneCountInString here
	if len(str) > n {
		return str[:n]
	}

	return str
}

func getNullInt64QueryParam(param string) sql.NullInt64 {
	i, err := strconv.Atoi(param)
	return sql.NullInt64{Int64: int64(i), Valid: err == nil}
}

func getIntQueryParam(param string) int {
	i, _ := strconv.Atoi(param)
	return i
}

func getURLQueryParam(param string) string {
	if _, err := url.ParseRequestURI(param); err != nil {
		return ""
	}

	return param
}
