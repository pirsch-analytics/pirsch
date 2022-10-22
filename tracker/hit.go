package tracker

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/tracker/geodb"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ip"
	"github.com/pirsch-analytics/pirsch/v4/tracker/language"
	"github.com/pirsch-analytics/pirsch/v4/tracker/referrer"
	"github.com/pirsch-analytics/pirsch/v4/tracker/screen"
	"github.com/pirsch-analytics/pirsch/v4/tracker/session"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ua"
	"github.com/pirsch-analytics/pirsch/v4/tracker/utm"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"net"
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

// SessionState is the state and cancellation for a session.
// The sessions must be inserted together to ensure sessions collapse.
type SessionState struct {
	// State is the new state for the session.
	State model.Session

	// Cancel is the state to cancel.
	// On session creation, this field is nil.
	Cancel *model.Session
}

// HitOptions is used to manipulate the data saved on a hit.
type HitOptions struct {
	// Salt is used to generate a fingerprint (optional).
	// It can be different for every request.
	Salt string

	// SessionCache is the cache to look up sessions.
	SessionCache session.Cache

	// HeaderParser is an (optional) list of parsers to extract the real client IP from request headers.
	HeaderParser []ip.HeaderParser

	// AllowedProxySubnets is an (optional) list of subnets to trust when extracting the real client IP from request headers.
	AllowedProxySubnets []net.IPNet

	// ClientID is optionally saved with a hit to split the data between multiple clients.
	ClientID uint64

	// SessionMaxAge defines the maximum time a session stays active.
	// A session is kept active if requests are made within the time frame.
	// Set to 15 minutes by default.
	SessionMaxAge time.Duration

	// MinDelay defines the minimum time in milliseconds between two page views before the session is flagged as a bot request.
	// This will update the Session.IsBot counter, which can later be used to filter upon.
	// Set it to 0 to disable flagging bots.
	MinDelay int64

	// IsBotThreshold sets the threshold before a request is ignored.
	// If Session.IsBot is larger or equal to the configured value, the request won't be accepted.
	// Set it to 0 to disable ignoring requests.
	IsBotThreshold uint8

	// MaxPageViews defines the maximum number of page views a session might have.
	// Once the maximum is reached, the session will be flagged as a bot, which can later be used to filter upon.
	// Only when IsBotThreshold is set this will have an effect on whether the session is updated or not.
	// Set it to 0 to disable flagging bots.
	MaxPageViews uint16

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

	event bool
	geoDB *geodb.GeoDB
}

// HitFromRequest returns a new PageView and Session for given request, salt and HitOptions.
// The salt must stay consistent to track visitors across multiple calls.
// The easiest way to track visitors is to use the Tracker.
func HitFromRequest(r *http.Request, salt string, options *HitOptions) (*model.PageView, SessionState, *model.UserAgent) {
	now := time.Now().UTC() // capture first to get as close as possible

	if options == nil {
		return nil, SessionState{}, nil
	}

	// set default options in case they're nil
	if options.SessionMaxAge.Seconds() == 0 {
		options.SessionMaxAge = defaultSessionMaxAge
	}

	getRequestURI(r, options)
	path := getPath(options.Path)
	title := shortenString(options.Title, 512)

	// find session for today and maximum session age
	salt += options.Salt
	fingerprint := Fingerprint(r, salt, now, options.HeaderParser, options.AllowedProxySubnets)
	m := options.SessionCache.NewMutex(options.ClientID, fingerprint)
	m.Lock()
	defer m.Unlock()
	sessionMaxAge := now.Add(-options.SessionMaxAge)
	s := options.SessionCache.Get(options.ClientID, fingerprint, sessionMaxAge)

	// if the maximum session age reaches yesterday, we also need to check for the previous day
	if s == nil && sessionMaxAge.Day() != now.Day() {
		fingerprintYesterday := Fingerprint(r, salt, sessionMaxAge, options.HeaderParser, options.AllowedProxySubnets)
		my := options.SessionCache.NewMutex(options.ClientID, fingerprintYesterday)
		my.Lock()
		defer my.Unlock()
		s = options.SessionCache.Get(options.ClientID, fingerprintYesterday, sessionMaxAge)

		if s != nil {
			if s.Start.Before(now.Add(-time.Hour * 24)) {
				s = nil
			} else {
				fingerprint = fingerprintYesterday
			}
		}
	}

	var sessionState SessionState
	var timeOnPage uint32
	var userAgent *model.UserAgent

	if s == nil || referrerOrCampaignChanged(s, r, options) {
		s, userAgent = newSession(r, options, fingerprint, now, path, title)
		sessionState.State = *s
		options.SessionCache.Put(options.ClientID, fingerprint, s)
	} else {
		if options.IsBotThreshold > 0 && s.IsBot >= options.IsBotThreshold {
			return nil, SessionState{}, nil
		}

		s.Sign = -1
		sessionState.Cancel = s
		state := *s
		timeOnPage = updateSession(options, &state, options.event, now, path, title)
		sessionState.State = state
		options.SessionCache.Put(options.ClientID, fingerprint, &state)
	}

	return &model.PageView{
		ClientID:        sessionState.State.ClientID,
		VisitorID:       sessionState.State.VisitorID,
		SessionID:       sessionState.State.SessionID,
		Time:            sessionState.State.Time,
		DurationSeconds: timeOnPage,
		Path:            sessionState.State.ExitPath,
		Title:           sessionState.State.ExitTitle,
		Language:        sessionState.State.Language,
		CountryCode:     sessionState.State.CountryCode,
		City:            sessionState.State.City,
		Referrer:        sessionState.State.Referrer,
		ReferrerName:    sessionState.State.ReferrerName,
		ReferrerIcon:    sessionState.State.ReferrerIcon,
		OS:              sessionState.State.OS,
		OSVersion:       sessionState.State.OSVersion,
		Browser:         sessionState.State.Browser,
		BrowserVersion:  sessionState.State.BrowserVersion,
		Desktop:         sessionState.State.Desktop,
		Mobile:          sessionState.State.Mobile,
		ScreenWidth:     sessionState.State.ScreenWidth,
		ScreenHeight:    sessionState.State.ScreenHeight,
		ScreenClass:     sessionState.State.ScreenClass,
		UTMSource:       sessionState.State.UTMSource,
		UTMMedium:       sessionState.State.UTMMedium,
		UTMCampaign:     sessionState.State.UTMCampaign,
		UTMContent:      sessionState.State.UTMContent,
		UTMTerm:         sessionState.State.UTMTerm,
	}, sessionState, userAgent
}

// ExtendSession looks up and extends the session for given request.
// This function does not Store a hit or event in database.
func ExtendSession(r *http.Request, salt string, options *HitOptions) {
	if options == nil {
		return
	}

	now := time.Now().UTC()
	fingerprint := Fingerprint(r, salt+options.Salt, now, options.HeaderParser, options.AllowedProxySubnets)
	s := options.SessionCache.Get(options.ClientID, fingerprint, now.Add(-options.SessionMaxAge))

	if s != nil {
		s.Time = now
		options.SessionCache.Put(options.ClientID, fingerprint, s)
	}
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

	if userAgent == "" || len(userAgent) < 10 || len(userAgent) > 300 || ua.ContainsNonASCIICharacters(userAgent) {
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
	if referrer.Ignore(r) {
		return true
	}

	userAgentResult := ua.Parse(r.UserAgent())

	if ignoreBrowserVersion(userAgentResult.Browser, userAgentResult.BrowserVersion) {
		return true
	}

	// filter for bot keywords (most expensive operation last)
	for _, botUserAgent := range ua.Blacklist {
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
		ClientID:     getIntQueryParam[uint64](query.Get("client_id")),
		URL:          getURLQueryParam(query.Get("url")),
		Title:        strings.TrimSpace(query.Get("t")),
		Referrer:     getURLQueryParam(query.Get("ref")),
		ScreenWidth:  getIntQueryParam[uint16](query.Get("w")),
		ScreenHeight: getIntQueryParam[uint16](query.Get("h")),
	}
}

func newSession(r *http.Request, options *HitOptions, fingerprint uint64, now time.Time, path, title string) (*model.Session, *model.UserAgent) {
	// shorten strings if required and parse User-Agent to extract more data (OS, Browser)
	userAgent := r.UserAgent()
	uaInfo := ua.Parse(userAgent)
	uaInfo.OS = shortenString(uaInfo.OS, 20)
	uaInfo.OSVersion = shortenString(uaInfo.OSVersion, 20)
	uaInfo.Browser = shortenString(uaInfo.Browser, 20)
	uaInfo.BrowserVersion = shortenString(uaInfo.BrowserVersion, 20)
	lang := shortenString(language.Get(r), 10)
	ref, referrerName, referrerIcon := referrer.Get(r, options.Referrer, options.ReferrerDomainBlacklist, options.ReferrerDomainBlacklistIncludesSubdomains)
	ref = shortenString(ref, 200)
	referrerName = shortenString(referrerName, 200)
	referrerIcon = shortenString(referrerIcon, 2000)
	screenClass := screen.GetClass(options.ScreenWidth)
	utmParams := utm.Get(r)
	countryCode, city := "", ""

	if options.geoDB != nil {
		countryCode, city = options.geoDB.CountryCodeAndCity(ip.Get(r, options.HeaderParser, options.AllowedProxySubnets))
	}

	if options.ScreenWidth <= 0 || options.ScreenHeight <= 0 {
		options.ScreenWidth = 0
		options.ScreenHeight = 0
	}

	return &model.Session{
		Sign:           1,
		ClientID:       options.ClientID,
		VisitorID:      fingerprint,
		SessionID:      util.RandUint32(),
		Time:           now,
		Start:          now,
		EntryPath:      path,
		ExitPath:       path,
		PageViews:      1,
		IsBounce:       true,
		EntryTitle:     title,
		ExitTitle:      title,
		Language:       lang,
		CountryCode:    countryCode,
		City:           city,
		Referrer:       ref,
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
		ScreenClass:    screenClass,
		UTMSource:      utmParams.Source,
		UTMMedium:      utmParams.Medium,
		UTMCampaign:    utmParams.Campaign,
		UTMContent:     utmParams.Content,
		UTMTerm:        utmParams.Term,
	}, &uaInfo
}

func updateSession(options *HitOptions, session *model.Session, event bool, now time.Time, path, title string) uint32 {
	if options.MaxPageViews > 0 && session.PageViews+1 >= options.MaxPageViews {
		session.IsBot = 255
	} else if options.MinDelay > 0 && now.UnixMilli()-session.Time.UnixMilli() < options.MinDelay {
		session.IsBot++
	}

	top := now.Unix() - session.Time.Unix()

	if top < 0 {
		top = 0
	}

	duration := now.Unix() - session.Start.Unix()

	if duration < 0 {
		duration = 0
	}

	session.DurationSeconds = uint32(duration)
	session.Sign = 1
	session.Time = now

	if event {
		session.IsBounce = false
	} else {
		session.IsBounce = session.IsBounce && path == session.ExitPath
		session.ExitPath = path
		session.ExitTitle = title
		session.PageViews++
	}

	return uint32(top)
}

func referrerOrCampaignChanged(session *model.Session, r *http.Request, options *HitOptions) bool {
	ref, _, _ := referrer.Get(r, options.Referrer, options.ReferrerDomainBlacklist, options.ReferrerDomainBlacklistIncludesSubdomains)

	if ref != "" && ref != session.Referrer {
		return true
	}

	utmParams := utm.Get(r)
	return (utmParams.Source != "" && utmParams.Source != session.UTMSource) ||
		(utmParams.Medium != "" && utmParams.Medium != session.UTMMedium) ||
		(utmParams.Campaign != "" && utmParams.Campaign != session.UTMCampaign) ||
		(utmParams.Content != "" && utmParams.Content != session.UTMContent) ||
		(utmParams.Term != "" && utmParams.Term != session.UTMTerm)
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
		browser == pirsch.BrowserChrome && browserVersionBefore(version, minChromeVersion) ||
		browser == pirsch.BrowserFirefox && browserVersionBefore(version, minFirefoxVersion) ||
		browser == pirsch.BrowserSafari && browserVersionBefore(version, minSafariVersion) ||
		browser == pirsch.BrowserOpera && browserVersionBefore(version, minOperaVersion) ||
		browser == pirsch.BrowserEdge && browserVersionBefore(version, minEdgeVersion) ||
		browser == pirsch.BrowserIE && browserVersionBefore(version, minIEVersion)
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

func getIntQueryParam[T uint16 | uint64](param string) T {
	i, _ := strconv.ParseUint(param, 10, 64)
	return T(i)
}

func getURLQueryParam(param string) string {
	if _, err := url.ParseRequestURI(param); err != nil {
		return ""
	}

	return param
}
