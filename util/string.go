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
