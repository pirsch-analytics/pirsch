package header

import (
	"strings"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

// Header filters bot traffic based on HTTP headers.
type Header struct{}

// NewHeader returns a new Header.
func NewHeader() *Header {
	return new(Header)
}

// Step implements ingest.PipeStep to process a step.
func (h *Header) Step(request *ingest.Request) (bool, error) {
	if request.DisableBotFilter {
		return false, nil
	}

	// ignore User-Agent missing
	if request.Request.UserAgent() == "" {
		request.BotReason = "ua-missing"
		return true, nil
	}

	// ignore Accept-Language missing
	if request.Request.Header.Get("Accept-Language") == "" {
		request.BotReason = "al-missing"
		return true, nil
	}

	// ignore Accept-Encoding missing
	if request.Request.Header.Get("Accept-Encoding") == "" {
		request.BotReason = "ae-missing"
		return true, nil
	}

	// ignore HTTP/2 requests with connection header set to close
	connection := strings.ToLower(strings.TrimSpace(request.Request.Header.Get("Connection")))

	if request.Request.ProtoMajor == 2 && connection == "close" {
		request.BotReason = "http2-close"
		return true, nil
	}

	// ignore HTTP/2 requests with connection header set to keep-alive
	if request.Request.ProtoMajor == 2 && connection == "keep-alive" {
		request.BotReason = "http2-alive"
		return true, nil
	}

	// ignore TE header set
	if request.Request.Header.Get("TE") != "" {
		request.BotReason = "te"
		return true, nil
	}

	// ignore Pragma set without Cache-Control
	pragma := request.Request.Header.Get("Pragma")
	cacheControl := request.Request.Header.Get("Cache-Control")

	if pragma != "" && cacheControl == "" {
		request.BotReason = "pragma-cc"
		return true, nil
	}

	// ignore Sec-Fetch-Site: none with referrer set
	secFetchSite := strings.ToLower(strings.TrimSpace(request.Request.Header.Get("Sec-Fetch-Site")))

	if secFetchSite == "none" && request.Request.Referer() != "" {
		request.BotReason = "sfs-referrer"
		return true, nil
	}

	// ignore Upgrade-Insecure-Requests for CORS requests
	upgradeInsecureRequests := strings.TrimSpace(request.Request.Header.Get("Upgrade-Insecure-Requests"))
	secFetchMode := strings.ToLower(strings.TrimSpace(request.Request.Header.Get("Sec-Fetch-Mode")))

	if upgradeInsecureRequests == "1" && secFetchMode == "cors" {
		request.BotReason = "ui-cors"
		return true, nil
	}

	// ignore Sec-Fetch-Dest: empty in combination with Upgrade-Insecure-Requests
	secFetchDest := request.Request.Header.Get("Sec-Fetch-Dest")

	if secFetchDest == "empty" && upgradeInsecureRequests == "1" {
		request.BotReason = "sfd-ui"
		return true, nil
	}

	// ignore modern header with HTTP/1.1
	if request.Request.ProtoMajor == 1 &&
		request.Request.ProtoMinor == 1 &&
		(secFetchSite != "" || secFetchMode != "" || secFetchDest != "") {
		request.BotReason = "http11-sf"
		return true, nil
	}

	return false, nil
}
