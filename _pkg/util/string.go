package util

import (
	"bytes"
	"unicode"
	"unicode/utf8"
)

// ContainsNonASCIICharacters returns whether the string only consists out of ASCII characters or not.
func ContainsNonASCIICharacters(ua string) bool {
	for _, c := range ua {
		if c > unicode.MaxASCII {
			return true
		}
	}

	return false
}

// ShortenString shortens given string to the given maximum number of bytes.
func ShortenString(str string, n int) string {
	if utf8.RuneCountInString(str) > n {
		runes := bytes.Runes([]byte(str))
		return string(runes[:n])
	}

	return str
}
