package ip

import (
	"net"
	"strings"
)

// ParseForwardedHeader parses the Forwarded header.
func ParseForwardedHeader(value string) string {
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

// ParseXForwardedForHeader parses the X-Forwarded header.
func ParseXForwardedForHeader(value string) string {
	parts := strings.Split(value, ",")

	if len(parts) > 0 {
		ip := cleanIP(strings.TrimSpace(parts[len(parts)-1]))

		if isValidIP(ip) {
			return ip
		}
	}

	return ""
}

// ParseXRealIPHeader parses the X-Real-IP header.
func ParseXRealIPHeader(value string) string {
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
