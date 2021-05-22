package pirsch

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"
)

// Fingerprint returns a hash for given request, salt, and timezone.
// The hash is unique for the visitor.
func Fingerprint(r *http.Request, salt string, timezone *time.Location) string {
	var sb strings.Builder
	sb.WriteString(r.Header.Get("User-Agent"))
	sb.WriteString(getIP(r))
	sb.WriteString(time.Now().In(timezone).Format("20060102"))
	sb.WriteString(salt)
	hash := md5.New()

	if _, err := io.WriteString(hash, sb.String()); err != nil {
		return "" // this should never fail actually...
	}

	return hex.EncodeToString(hash.Sum(nil))
}
