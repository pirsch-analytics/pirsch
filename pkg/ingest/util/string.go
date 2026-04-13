package util

import "strings"

// StripWWW strips the www prefix if it is the first and only subdomain.
func StripWWW(hostname string) string {
	if strings.Count(hostname, ".") == 2 && strings.HasPrefix(hostname, "www.") {
		return hostname[4:]
	}

	return hostname
}
