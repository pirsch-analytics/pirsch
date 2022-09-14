package tracker

import (
	"github.com/pirsch-analytics/pirsch/v4/tracker/ip"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/dchest/siphash"
)

var (
	hashKey0, hashKey1 uint64
)

// Fingerprint returns a hash for given request and salt.
// The hash is unique for the visitor and day.
func Fingerprint(r *http.Request, salt string, date time.Time, headerParser []ip.HeaderParser, allowed []net.IPNet) uint64 {
	var sb strings.Builder
	sb.WriteString(r.Header.Get("User-Agent"))
	sb.WriteString(ip.Get(r, headerParser, allowed))
	sb.WriteString(salt)
	sb.WriteString(date.Format("20060102"))
	return siphash.Hash(hashKey0, hashKey1, []byte(sb.String()))
}

// SetFingerprintKeys used to set the SipHash keys for fingerprints.
// This function is NOT concurrency save and should be called once on startup (before generating the first fingerprint).
func SetFingerprintKeys(key0, key1 uint64) {
	hashKey0 = key0
	hashKey1 = key1
}
