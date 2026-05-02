package ip

import (
	"net"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

var (
	// CFConnectingIP is a HeaderParser.
	// https://support.cloudflare.com/hc/en-us/articles/206776727-What-is-True-Client-IP-
	CFConnectingIP = HeaderParser{"CF-Connecting-IP", ParseXForwardedForHeader}

	// TrueClientIP is a HeaderParser.
	TrueClientIP = HeaderParser{"True-Client-IP", ParseXForwardedForHeader}

	// XForwardedFor is a HeaderParser.
	XForwardedFor = HeaderParser{"X-Forwarded-For", ParseXForwardedForHeader}

	// Forwarded is a HeaderParser.
	Forwarded = HeaderParser{"Forwarded", ParseForwardedHeader}

	// XRealIP is a HeaderParser.
	XRealIP = HeaderParser{"X-Real-IP", ParseXRealIPHeader}

	// DefaultHeaderParser is a list of headers and corresponding parsers to look up the real client IP.
	// They will be checked in order, the first non-empty one will be picked,
	// or else the remote address is selected.
	DefaultHeaderParser = []HeaderParser{
		CFConnectingIP,
		TrueClientIP,
		XForwardedFor,
		Forwarded,
		XRealIP,
	}
)

// ParseHeaderFunc parses and validates an IP address from a header.
// It must return an empty string if the header or contained IP address is invalid.
type ParseHeaderFunc func(string) string

// HeaderParser parses a header to extract the real client IP address.
type HeaderParser struct {
	Header string
	Parser ParseHeaderFunc
}

// IP extracts the real visitor IP address from the remote address or headers if configured.
type IP struct {
	parser         []HeaderParser
	allowedSubnets []net.IPNet
}

// NewIP returns a new IP for the given header parsers and allowed networks.
func NewIP(parser []HeaderParser, allowed []net.IPNet) *IP {
	return &IP{
		parser:         parser,
		allowedSubnets: allowed,
	}
}

// Step implements ingest.PipeStep to process a step.
// It will set the remote IP address or try to extract the real client IP from headers if configured.
func (i *IP) Step(request *ingest.Request) (bool, error) {
	ip := cleanIP(request.Request.RemoteAddr)

	if len(i.allowedSubnets) > 0 && !i.validProxySource(ip, i.allowedSubnets) {
		return true, nil
	}

	for _, header := range i.parser {
		value := request.Request.Header.Get(header.Header)

		if value != "" {
			parsedIP := header.Parser(value)

			if parsedIP != "" {
				request.IP = parsedIP
				return false, nil
			}
		}
	}

	request.IP = ip
	return false, nil
}

func (i *IP) validProxySource(address string, allowed []net.IPNet) bool {
	ip := net.ParseIP(address)

	if ip == nil {
		return false
	}

	for _, from := range allowed {
		if from.Contains(ip) {
			return true
		}
	}

	return false
}
