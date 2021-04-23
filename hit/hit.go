package hit

import (
	"database/sql"
	"encoding/json"
	iso6391 "github.com/emvi/iso-639-1"
	"github.com/pirsch-analytics/pirsch/geodb"
	"github.com/pirsch-analytics/pirsch/ua"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// version numbers from early 2018
	minChromeVersion  = 64
	minFirefoxVersion = 58
	minSafariVersion  = 11
	minOperaVersion   = 50
	minEdgeVersion    = 80
	minIEVersion      = 11
)

// Hit represents a single data point/page visit and is the central entity of Pirsch.
type Hit struct {
	TenantID       sql.NullInt64  `db:"tenant_id" json:"tenant_id,omitempty"`
	Fingerprint    string         `db:"fingerprint" json:"fingerprint"`
	Time           time.Time      `db:"time" json:"time"`
	Session        sql.NullTime   `db:"session" json:"session,omitempty"`
	UserAgent      string         `db:"user_agent" json:"user_agent"`
	Path           string         `db:"path" json:"path"`
	URL            string         `db:"url" json:"url"`
	Language       sql.NullString `db:"language" json:"language,omitempty"`
	CountryCode    sql.NullString `db:"country_code" json:"country_code,omitempty"`
	Referrer       sql.NullString `db:"referrer" json:"referrer,omitempty"`
	ReferrerName   sql.NullString `db:"referrer_name" json:"referrer_name,omitempty"`
	ReferrerIcon   sql.NullString `db:"referrer_icon" json:"referrer_icon,omitempty"`
	OS             sql.NullString `db:"os" json:"os,omitempty"`
	OSVersion      sql.NullString `db:"os_version" json:"os_version,omitempty"`
	Browser        sql.NullString `db:"browser" json:"browser,omitempty"`
	BrowserVersion sql.NullString `db:"browser_version" json:"browser_version,omitempty"`
	Desktop        bool           `db:"desktop" json:"desktop"`
	Mobile         bool           `db:"mobile" json:"mobile"`
	ScreenWidth    int            `db:"screen_width" json:"screen_width"`
	ScreenHeight   int            `db:"screen_height" json:"screen_height"`
	ScreenClass    sql.NullString `db:"screen_class" json:"screen_class,omitempty"`
}

// String implements the Stringer interface.
func (hit Hit) String() string {
	out, _ := json.Marshal(hit)
	return string(out)
}

// Options is used to manipulate the data saved on a hit.
type Options struct {
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

	// ReferrerDomainBlacklist is used to filter out unwanted referrers from the Referrer header.
	// This can be used to filter out traffic from your own site or subdomains.
	// To filter your own domain and subdomains, add your domain to the list and set ReferrerDomainBlacklistIncludesSubdomains to true.
	// This way the referrer for blog.mypage.com -> mypage.com won't be saved.
	ReferrerDomainBlacklist []string

	// ReferrerDomainBlacklistIncludesSubdomains set to true to include all subdomains in the ReferrerDomainBlacklist,
	// or else subdomains must explicitly be included in the blacklist.
	// If the blacklist contains domain.com, sub.domain.com and domain.com will be treated as equals.
	ReferrerDomainBlacklistIncludesSubdomains bool

	// ScreenWidth sets the screen width to be stored with the hit.
	ScreenWidth int

	// ScreenHeight sets the screen height to be stored with the hit.
	ScreenHeight int

	geoDB *geodb.GeoDB
	//sessionCache *sessionCache
}

// FromRequest returns a new Hit for given request, salt and Options.
// The salt must stay consistent to track visitors across multiple calls.
// The easiest way to track visitors is to use the Tracker.
func FromRequest(r *http.Request, salt string, options *Options) Hit {
	now := time.Now().UTC() // capture first to get as close as possible

	// set default options in case they're nil
	if options == nil {
		options = &Options{}
	}

	// shorten strings if required and parse User-Agent to extract more data (OS, Browser)
	getRequestURI(r, options)
	fingerprint := Fingerprint(r, salt)
	userAgent := r.UserAgent()
	path := shortenString(options.Path, 2000)
	requestURL := shortenString(options.URL, 2000)
	uaInfo := ua.ParseUserAgent(userAgent)
	uaInfo.OS = shortenString(uaInfo.OS, 20)
	uaInfo.OSVersion = shortenString(uaInfo.OSVersion, 20)
	uaInfo.Browser = shortenString(uaInfo.Browser, 20)
	uaInfo.BrowserVersion = shortenString(uaInfo.BrowserVersion, 20)
	userAgent = shortenString(userAgent, 200)
	lang := shortenString(getLanguage(r), 10)
	referrer, referrerName, referrerIcon := getReferrer(r, options.Referrer, options.ReferrerDomainBlacklist, options.ReferrerDomainBlacklistIncludesSubdomains)
	referrer = shortenString(referrer, 200)
	referrerName = shortenString(referrerName, 200)
	referrerIcon = shortenString(referrerIcon, 2000)
	screen := GetScreenClass(options.ScreenWidth)
	countryCode := ""

	if options.geoDB != nil {
		countryCode = options.geoDB.CountryCode(getIP(r))
	}

	var session time.Time

	/*if options.sessionCache != nil {
		session = options.sessionCache.find(options.TenantID, fingerprint)
	}*/

	if options.ScreenWidth <= 0 || options.ScreenHeight <= 0 {
		options.ScreenWidth = 0
		options.ScreenHeight = 0
	}

	if path == "" {
		path = "/"
	}

	return Hit{
		TenantID:       options.TenantID,
		Fingerprint:    fingerprint,
		Time:           now,
		Session:        sql.NullTime{Time: session, Valid: !session.IsZero()},
		UserAgent:      userAgent,
		Path:           path,
		URL:            requestURL,
		Language:       sql.NullString{String: lang, Valid: lang != ""},
		CountryCode:    sql.NullString{String: countryCode, Valid: countryCode != ""},
		Referrer:       sql.NullString{String: referrer, Valid: referrer != ""},
		ReferrerName:   sql.NullString{String: referrerName, Valid: referrerName != ""},
		ReferrerIcon:   sql.NullString{String: referrerIcon, Valid: referrerIcon != ""},
		OS:             sql.NullString{String: uaInfo.OS, Valid: uaInfo.OS != ""},
		OSVersion:      sql.NullString{String: uaInfo.OSVersion, Valid: uaInfo.OSVersion != ""},
		Browser:        sql.NullString{String: uaInfo.Browser, Valid: uaInfo.Browser != ""},
		BrowserVersion: sql.NullString{String: uaInfo.BrowserVersion, Valid: uaInfo.BrowserVersion != ""},
		Desktop:        uaInfo.IsDesktop(),
		Mobile:         uaInfo.IsMobile(),
		ScreenWidth:    options.ScreenWidth,
		ScreenHeight:   options.ScreenHeight,
		ScreenClass:    sql.NullString{String: screen, Valid: screen != ""},
	}
}

// Ignore returns true, if a hit should be ignored for given request, or false otherwise.
// The easiest way to track visitors is to use the Tracker.
func Ignore(r *http.Request) bool {
	// respect do not track header
	if r.Header.Get("DNT") == "1" {
		return true
	}

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

	userAgentResult := ua.ParseUserAgent(r.UserAgent())

	if ignoreBrowserVersion(userAgentResult.Browser, userAgentResult.BrowserVersion) {
		return true
	}

	// filter for bot keywords (most expensive operation last)
	for _, botUserAgent := range ua.UserAgentBlacklist {
		if strings.Contains(userAgent, botUserAgent) {
			return true
		}
	}

	return false
}

// OptionsFromRequest returns the Options for given client request.
// This function can be used to accept hits from pirsch.js. Invalid parameters are ignored and left empty.
// You might want to add additional checks before calling FromRequest afterwards (like for the Options.TenantID).
func OptionsFromRequest(r *http.Request) *Options {
	query := r.URL.Query()
	return &Options{
		TenantID:     getNullInt64QueryParam(query.Get("tenantid")),
		URL:          getURLQueryParam(query.Get("url")),
		Referrer:     getURLQueryParam(query.Get("ref")),
		ScreenWidth:  getIntQueryParam(query.Get("w")),
		ScreenHeight: getIntQueryParam(query.Get("h")),
	}
}

func ignoreBrowserVersion(browser, version string) bool {
	return version != "" &&
		browser == ua.BrowserChrome && browserVersionBefore(version, minChromeVersion) ||
		browser == ua.BrowserFirefox && browserVersionBefore(version, minFirefoxVersion) ||
		browser == ua.BrowserSafari && browserVersionBefore(version, minSafariVersion) ||
		browser == ua.BrowserOpera && browserVersionBefore(version, minOperaVersion) ||
		browser == ua.BrowserEdge && browserVersionBefore(version, minEdgeVersion) ||
		browser == ua.BrowserIE && browserVersionBefore(version, minIEVersion)
}

func browserVersionBefore(version string, min int) bool {
	i := strings.Index(version, ".")

	if i >= 0 {
		version = version[:i]
	}

	v, err := strconv.Atoi(version)

	if err != nil {
		return false
	}

	return v < min
}

func getRequestURI(r *http.Request, options *Options) {
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
