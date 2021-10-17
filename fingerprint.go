package pirsch

import (
	"github.com/dchest/siphash"
	"net/http"
	"strings"
)

var (
	hashKey0, hashKey1 uint64
)

// Fingerprint returns a hash for given request and salt.
// The hash is unique for the visitor.
func Fingerprint(r *http.Request, salt string) uint64 {
	var sb strings.Builder
	sb.WriteString(r.Header.Get("User-Agent"))
	sb.WriteString(getIP(r))
	sb.WriteString(salt)
	return siphash.Hash(hashKey0, hashKey1, []byte(sb.String()))
}

// SetFingerprintKeys used to set the SipHash keys for fingerprints.
// This function is NOT concurrency save and should be called once on startup (before generating the first fingerprint).
func SetFingerprintKeys(key0, key1 uint64) {
	hashKey0 = key0
	hashKey1 = key1
}
