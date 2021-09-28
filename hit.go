package pirsch

import (
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	minChromeVersion  = 71 // late 2018
	minFirefoxVersion = 63 // late 2018
	minSafariVersion  = 12 // late 2018
	minOperaVersion   = 57 // late 2018
	minEdgeVersion    = 88 // late 2020
	minIEVersion      = 11 // late 2013

	defaultSessionMaxAge = time.Minute * 15
)

// HitOptions is used to manipulate the data saved on a hit.
type HitOptions struct {
	// Client is the database client required to look up sessions.
	//Client Store

	// SessionCache is the cache to look up sessions.
	SessionCache SessionCache

	// ClientID is optionally saved with a hit to split the data between multiple clients.
	ClientID uint64

	// SessionMaxAge defines the maximum time a session stays active.
	// A session is kept active if requests are made within the time frame.
	// Set to 15 minutes by default.
	SessionMaxAge time.Duration

	// URL can be set to manually overwrite the URL stored for this request.
	// This will also affect the Path, except it is set too.
	URL string

	// Path can be set to manually overwrite the path stored for the request.
	// This will also affect the URL.
	Path string

	// Title is the page title.
	Title string

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
	ScreenWidth uint16

	// ScreenHeight sets the screen height to be stored with the hit.
	ScreenHeight uint16

	geoDB *GeoDB
}

// HitFromRequest returns a new Hit for given request, salt and HitOptions.
// The salt must stay consistent to track visitors across multiple calls.
// The easiest way to track visitors is to use the Tracker.
// The options must be set!
func HitFromRequest(r *http.Request, salt string, options *HitOptions) (*Hit, *UserAgent) {
	now := time.Now().UTC() // capture first to get as close as possible, hits and sessions use UTC

	if options == nil {
		return nil, nil
	}

	// set default options in case they're nil
	if options.SessionMaxAge.Seconds() == 0 {
		options.SessionMaxAge = defaultSessionMaxAge
	}

	fingerprint := Fingerprint(r, salt)
	getRequestURI(r, options)
	path := getPath(options.Path)
	title := shortenString(options.Title, 512)
	hit := options.SessionCache.Get(options.ClientID, fingerprint, time.Now().UTC().Add(-options.SessionMaxAge))
	var ua *UserAgent

	if hit == nil {
		// shorten strings if required and parse User-Agent to extract more data (OS, Browser)
		userAgent := r.UserAgent()
		uaInfo := ParseUserAgent(userAgent)
		ua = &uaInfo
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
		utm := getUTMParams(r)
		countryCode, city := "", ""

		if options.geoDB != nil {
			countryCode, city = options.geoDB.CountryCodeAndCity(getIP(r))
		}

		if options.ScreenWidth <= 0 || options.ScreenHeight <= 0 {
			options.ScreenWidth = 0
			options.ScreenHeight = 0
		}

		hit = &Hit{
			ClientID:       options.ClientID,
			Fingerprint:    fingerprint,
			Time:           now,
			SessionID:      rand.Uint32(),
			Path:           path,
			EntryPath:      path,
			PageViews:      1,
			IsBounce:       true,
			Title:          title,
			Language:       lang,
			CountryCode:    countryCode,
			City:           city,
			Referrer:       referrer,
			ReferrerName:   referrerName,
			ReferrerIcon:   referrerIcon,
			OS:             uaInfo.OS,
			OSVersion:      uaInfo.OSVersion,
			Browser:        uaInfo.Browser,
			BrowserVersion: uaInfo.BrowserVersion,
			Desktop:        uaInfo.IsDesktop(),
			Mobile:         uaInfo.IsMobile(),
			ScreenWidth:    options.ScreenWidth,
			ScreenHeight:   options.ScreenHeight,
			ScreenClass:    screen,
			UTMSource:      utm.source,
			UTMMedium:      utm.medium,
			UTMCampaign:    utm.campaign,
			UTMContent:     utm.content,
			UTMTerm:        utm.term,
		}
	} else {
		hit.DurationSeconds = uint32(now.Unix() - hit.Time.Unix())
		hit.IsBounce = hit.IsBounce && path == hit.Path
		hit.Time = now
		hit.Path = path
		hit.PageViews++
		hit.Title = title
	}

	options.SessionCache.Put(options.ClientID, fingerprint, hit)
	return hit, ua
}

// IgnoreHit returns true, if a hit should be ignored for given request, or false otherwise.
// The easiest way to track visitors is to use the Tracker.
func IgnoreHit(r *http.Request) bool {
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

	userAgentResult := ParseUserAgent(r.UserAgent())

	if ignoreBrowserVersion(userAgentResult.Browser, userAgentResult.BrowserVersion) {
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
// You might want to add additional checks before calling HitFromRequest afterwards (like for the HitOptions.ClientID).
func HitOptionsFromRequest(r *http.Request) *HitOptions {
	query := r.URL.Query()
	return &HitOptions{
		ClientID:     getUInt64QueryParam(query.Get("client_id")),
		URL:          getURLQueryParam(query.Get("url")),
		Title:        strings.TrimSpace(query.Get("t")),
		Referrer:     getURLQueryParam(query.Get("ref")),
		ScreenWidth:  getUInt16QueryParam(query.Get("w")),
		ScreenHeight: getUInt16QueryParam(query.Get("h")),
	}
}

func getPath(path string) string {
	path = shortenString(path, 2000)

	if path == "" {
		return "/"
	}

	return path
}

func ignoreBrowserVersion(browser, version string) bool {
	return version != "" &&
		browser == BrowserChrome && browserVersionBefore(version, minChromeVersion) ||
		browser == BrowserFirefox && browserVersionBefore(version, minFirefoxVersion) ||
		browser == BrowserSafari && browserVersionBefore(version, minSafariVersion) ||
		browser == BrowserOpera && browserVersionBefore(version, minOperaVersion) ||
		browser == BrowserEdge && browserVersionBefore(version, minEdgeVersion) ||
		browser == BrowserIE && browserVersionBefore(version, minIEVersion)
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

func shortenString(str string, n int) string {
	// we intentionally use len instead of utf8.RuneCountInString here
	if len(str) > n {
		return str[:n]
	}

	return str
}

func getUInt16QueryParam(param string) uint16 {
	i, _ := strconv.Atoi(param)
	return uint16(i)
}

func getUInt64QueryParam(param string) uint64 {
	i, _ := strconv.Atoi(param)
	return uint64(i)
}

func getURLQueryParam(param string) string {
	if _, err := url.ParseRequestURI(param); err != nil {
		return ""
	}

	return param
}
