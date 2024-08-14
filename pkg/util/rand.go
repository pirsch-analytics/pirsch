package util

import "math/rand/v2"

const (
	randStrChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// RandUint32 returns a random uint32.
func RandUint32() uint32 {
	return rand.Uint32()
}

// RandUint64 returns a random uint64.
func RandUint64() uint64 {
	return rand.Uint64()
}

// RandString returns a random string for given length.
func RandString(n int) string {
	randStr := make([]byte, n)

	for i := range randStr {
		randStr[i] = randStrChars[rand.IntN(len(randStrChars))]
	}

	return string(randStr)
}
