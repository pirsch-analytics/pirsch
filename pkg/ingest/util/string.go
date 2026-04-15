package util

import (
	"strings"
	"unicode"
)

// StripWWW strips the www prefix if it is the first and only subdomain.
func StripWWW(hostname string) string {
	if strings.Count(hostname, ".") == 2 && strings.HasPrefix(hostname, "www.") {
		return hostname[4:]
	}

	return hostname
}

// ContainsNonASCIICharacters returns true if the string only consists out of ASCII characters.
func ContainsNonASCIICharacters(ua string) bool {
	for i := 0; i < len(ua); i++ {
		if ua[i] > unicode.MaxASCII {
			return true
		}
	}

	return false
}
