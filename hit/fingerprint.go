package hit

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"
)

// Fingerprint returns a hash for given request and salt.
// The hash is unique for the visitor.
func Fingerprint(r *http.Request, salt string) string {
	var sb strings.Builder
	sb.WriteString(r.Header.Get("User-Agent"))
	sb.WriteString(getIP(r))
	sb.WriteString(time.Now().UTC().Format("20060102"))
	sb.WriteString(salt)
	hash := md5.New()

	if _, err := io.WriteString(hash, sb.String()); err != nil {
		return "" // this should never fail actually...
	}

	return hex.EncodeToString(hash.Sum(nil))
}
