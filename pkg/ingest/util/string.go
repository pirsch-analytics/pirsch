package util

import (
	"strings"
)

// StripWWW strips the www prefix if it is the first and only subdomain.
func StripWWW(hostname string) string {
	if strings.Count(hostname, ".") == 2 && strings.HasPrefix(hostname, "www.") {
		return hostname[4:]
	}

	return hostname
}

// Shorten shortens given string to the maximum length (runes, not bytes).
func Shorten(s string, n int) string {
	runes := []rune(s)

	if len(runes) <= n {
		return s
	}

	return string(runes[:n])
}
