package ua

import (
	"net"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/util"
)

const (
	minUserAgentLength = 16
	maxUserAgentLength = 500

	minChromeVersion  = 70 // late 2019
	minFirefoxVersion = 68 // mid 2019
	minSafariVersion  = 12 // late 2018
	minOperaVersion   = 65 // late 2019
	minEdgeVersion    = 88 // late 2020
	minIEVersion      = 11 // late 2013
)

// BotFilter filters bot requests based on their User-Agent information.
// This step operates on the ingest.Request User-Agent fields, which must be set before this step is run.
type BotFilter struct{}

// NewBotFilter creates a new BotFilter.
func NewBotFilter() *BotFilter {
	return new(BotFilter)
}

// Step implements the ingest.PipeStep interface.
func (f *BotFilter) Step(request *ingest.Request) (bool, error) {
	if request.DisableBotFilter {
		return false, nil
	}

	// empty User-Agents are usually bots
	rawUserAgent := request.Request.UserAgent()
	userAgent := strings.TrimSpace(strings.ToLower(rawUserAgent))

	if userAgent == "" ||
		len(userAgent) <= minUserAgentLength ||
		len(userAgent) > maxUserAgentLength ||
		util.ContainsNonASCIICharacters(userAgent) {
		request.BotReason = "ua-chars"
		return true, nil
	}

	// ignore User-Agents that are an IP address
	host := rawUserAgent

	if net.ParseIP(host) != nil {
		request.BotReason = "ua-ip"
		return true, nil
	}

	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(rawUserAgent)
	}

	if net.ParseIP(host) != nil {
		request.BotReason = "ua-ip"
		return true, nil
	}

	// filter UUIDs
	if _, err := uuid.Parse(rawUserAgent); err == nil {
		request.BotReason = "ua-uuid"
		return true, nil
	}

	// filter User-Agent
	if f.ignoreBrowserVersion(request.Browser, request.BrowserVersion) {
		request.BotReason = "browser"
		return true, nil
	}

	if request.Browser == pkg.BrowserFirefox && request.BrowserRevision != request.BrowserVersion {
		request.BotReason = "ua-rv-mismatch"
		return true, nil
	}

	// filter for bot keywords
	browser := strings.ToLower(request.Browser)

	for _, botBrowser := range BrowserBlacklist {
		if strings.Contains(browser, botBrowser) {
			request.BotReason = "ch-browser"
			return true, nil
		}
	}

	// filter for bot keywords
	for _, botUserAgent := range UserAgentBlacklist {
		if strings.Contains(userAgent, botUserAgent) {
			request.BotReason = "ua-keyword"
			return true, nil
		}
	}

	// filter for bot regex
	for _, botUserAgent := range UserAgentRegexBlacklist {
		if botUserAgent.MatchString(userAgent) {
			request.BotReason = "ua-regex"
			return true, nil
		}
	}

	return false, nil
}

func (f *BotFilter) ignoreBrowserVersion(browser, version string) bool {
	return version != "" &&
		browser == pkg.BrowserChrome && f.browserVersionBefore(version, minChromeVersion) ||
		browser == pkg.BrowserFirefox && f.browserVersionBefore(version, minFirefoxVersion) ||
		browser == pkg.BrowserSafari && f.browserVersionBefore(version, minSafariVersion) ||
		browser == pkg.BrowserOpera && f.browserVersionBefore(version, minOperaVersion) ||
		browser == pkg.BrowserEdge && f.browserVersionBefore(version, minEdgeVersion) ||
		browser == pkg.BrowserIE && f.browserVersionBefore(version, minIEVersion)
}

func (f *BotFilter) browserVersionBefore(version string, min int) bool {
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
