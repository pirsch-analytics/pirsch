package ua

import "unicode"

// ContainsNonASCIICharacters returns true if the string only consists out of ASCII characters.
func ContainsNonASCIICharacters(ua string) bool {
	for i := 0; i < len(ua); i++ {
		if ua[i] > unicode.MaxASCII {
			return true
		}
	}

	return false
}
