package pirsch

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Fingerprint returns a hash for given request.
// The hash is unique for the visitor, not for the page.
func Fingerprint(r *http.Request) string {
	var sb strings.Builder
	sb.WriteString(r.Header.Get("User-Agent"))
	sb.WriteString(getIP(r))
	sb.WriteString(time.Now().UTC().Format("20060102"))
	hash := md5.New()

	if _, err := io.WriteString(hash, sb.String()); err != nil {
		return "" // this should never fail in reality...
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func getIP(r *http.Request) string {
	// read the IP from header in the case the request was forwarded by a proxy
	forwarded := r.Header.Get("X-Forwarded-For")

	if forwarded != "" {
		// we are just interested in the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// alternative header
	forwarded = r.Header.Get("Forwarded")

	if forwarded != "" {
		// pick the first part (for=ip)
		ips := strings.Split(forwarded, ",")
		f := strings.Split(ips[0], ";")
		return strings.TrimSpace(f[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return ""
	}

	return host
}
