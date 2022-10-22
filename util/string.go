package util

import "unicode"

// ContainsNonASCIICharacters returns whether the string only consists out of ASCII characters or not.
func ContainsNonASCIICharacters(ua string) bool {
	for i := 0; i < len(ua); i++ {
		if ua[i] > unicode.MaxASCII {
			return true
		}
	}

	return false
}

// ShortenString shortens given string to the given maximum number of bytes.
func ShortenString(str string, n int) string {
	if len(str) > n {
		return str[:n]
	}

	return str
}
