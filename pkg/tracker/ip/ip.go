package ip

import (
	"net"
	"net/http"
	"strings"
)

var (
	// CFConnectingIP is an HeaderParser.
	// https://support.cloudflare.com/hc/en-us/articles/206776727-What-is-True-Client-IP-
	CFConnectingIP = HeaderParser{"CF-Connecting-IP", parseXForwardedForHeader}

	// TrueClientIP is an HeaderParser.
	TrueClientIP = HeaderParser{"True-Client-IP", parseXForwardedForHeader}

	// XForwardedFor is an HeaderParser.
	XForwardedFor = HeaderParser{"X-Forwarded-For", parseXForwardedForHeader}

	// Forwarded is an HeaderParser.
	Forwarded = HeaderParser{"Forwarded", parseForwardedHeader}

	// XRealIP is an HeaderParser.
	XRealIP = HeaderParser{"X-Real-IP", parseXRealIPHeader}

	// DefaultHeaderParser is a list of headers and corresponding parsers to look up the real client IP.
	// They will be check in order, the first non-empty one will be picked,
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

// Get returns the IP from given request.
// It will try to extract the real client IP from headers if possible.
func Get(r *http.Request, parser []HeaderParser, allowed []net.IPNet) string {
	ip := cleanIP(r.RemoteAddr)

	if len(allowed) > 0 && !validProxySource(ip, allowed) {
		return ip
	}

	for _, header := range parser {
		value := r.Header.Get(header.Header)

		if value != "" {
			parsedIP := header.Parser(value)

			if parsedIP != "" {
				return parsedIP
			}
		}
	}

	return ip
}

// cleanIP returns the ip without port, if any.
func cleanIP(ip string) string {
	if strings.Contains(ip, ":") {
		host, _, err := net.SplitHostPort(ip)

		if err != nil {
			return ip
		}

		return host
	}

	return ip
}

// validProxySource returns whether the IP is in the allowed proxy subnet list.
func validProxySource(address string, allowed []net.IPNet) bool {
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

// parseForwardedHeader parses the Forwarded header to extract the real client IP.
func parseForwardedHeader(value string) string {
	parts := strings.Split(value, ",")

	if len(parts) > 0 {
		parts = strings.Split(parts[len(parts)-1], ";")

		for _, part := range parts {
			k, ip, found := strings.Cut(part, "=")

			if found && strings.TrimSpace(k) == "for" {
				ip = cleanIP(ip)

				if isValidIP(ip) {
					return ip
				}
			}
		}
	}

	return ""
}

// parseXForwardedForHeader parses the X-Forwarded-For header to extract the real client IP.
func parseXForwardedForHeader(value string) string {
	parts := strings.Split(value, ",")

	if len(parts) > 0 {
		ip := cleanIP(strings.TrimSpace(parts[len(parts)-1]))

		if isValidIP(ip) {
			return ip
		}
	}

	return ""
}

// parseXRealIPHeader parses the X-Real-IP header to extract the real client IP.
func parseXRealIPHeader(value string) string {
	value = cleanIP(strings.TrimSpace(value))

	if isValidIP(value) {
		return value
	}

	return ""
}

func isValidIP(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil &&
		!ip.IsPrivate() &&
		!ip.IsLoopback() &&
		!ip.IsUnspecified()
}
