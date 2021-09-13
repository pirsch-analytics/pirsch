package pirsch

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

// Fingerprint returns a hash for given request and salt.
// The hash is unique for the visitor.
func Fingerprint(r *http.Request, salt string) string {
	var sb strings.Builder
	sb.WriteString(r.Header.Get("User-Agent"))
	sb.WriteString(getIP(r))
	sb.WriteString(salt)
	sha256Hash := sha256.New()

	if _, err := io.WriteString(sha256Hash, sb.String()); err != nil {
		return "" // this should not fail...
	}

	hash := sha256Hash.Sum(nil)
	md5Hash := md5.New()

	if _, err := md5Hash.Write(hash); err != nil {
		return "" // this should not fail...
	}

	return hex.EncodeToString(md5Hash.Sum(nil))
}
