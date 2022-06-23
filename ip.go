package pirsch

import (
	"net"
	"net/http"
	"strings"
)

// Headers and corresponding parser to look up the real client IP.
// They will be check in order, the first non-empty one will be picked,
// or else the remote address is selected.
// CF-Connecting-IP is a header added by Cloudflare: https://support.cloudflare.com/hc/en-us/articles/206776727-What-is-True-Client-IP-
var ipHeaders = []IPHeader{
	{"CF-Connecting-IP", ParseXForwardedForHeader},
	{"True-Client-IP", ParseXForwardedForHeader},
	{"X-Forwarded-For", ParseXForwardedForHeader},
	{"Forwarded", ParseForwardedHeader},
	{"X-Real-IP", ParseXRealIPHeader},
}

// ParseHeaderFunc parses and validates an IP address from a header.
// It must return an empty string if the header or contained IP address is invalid.
type ParseHeaderFunc func(string) string

// IPHeader parses a header to extract the real client IP address.
type IPHeader struct {
	Header string
	Parser ParseHeaderFunc
}

// ParseForwardedHeader parses the Forwarded header to extract the real client IP.
func ParseForwardedHeader(value string) string {
	parts := strings.Split(value, ",")

	if len(parts) > 0 {
		parts = strings.Split(parts[len(parts)-1], ";")

		for _, part := range parts {
			k, ip, found := strings.Cut(part, "=")

			if found && strings.TrimSpace(k) == "for" {
				if isValidIP(ip) {
					return ip
				}
			}
		}
	}

	return ""
}

// ParseXForwardedForHeader parses the X-Forwarded-For header to extract the real client IP.
func ParseXForwardedForHeader(value string) string {
	parts := strings.Split(value, ",")

	if len(parts) > 0 {
		ip := strings.TrimSpace(parts[len(parts)-1])

		if isValidIP(ip) {
			return ip
		}
	}

	return ""
}

// ParseXRealIPHeader parses the X-Real-IP header to extract the real client IP.
func ParseXRealIPHeader(value string) string {
	if isValidIP(value) {
		return strings.TrimSpace(value)
	}

	return ""
}

// getIP returns the IP from given request.
// It will try to extract the real client IP from headers if possible.
func getIP(r *http.Request) string {
	ip := r.RemoteAddr

	for _, header := range ipHeaders {
		value := r.Header.Get(header.Header)

		if value != "" {
			parsedIP := header.Parser(value)

			if parsedIP != "" {
				ip = parsedIP
				break
			}
		}
	}

	if strings.Contains(ip, ":") {
		host, _, err := net.SplitHostPort(ip)

		if err != nil {
			return ip
		}

		return host
	}

	return ip
}

func isValidIP(value string) bool {
	ip := net.ParseIP(value)
	return ip != nil &&
		!ip.IsPrivate() &&
		!ip.IsLoopback() &&
		!ip.IsUnspecified()
}
