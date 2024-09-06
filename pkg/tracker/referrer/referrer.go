package referrer

import (
	"errors"
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
)

// QueryParams is a list of query parameters to set the referrer.
var QueryParams = []struct {
	param        string
	preferHeader bool
}{
	{"ref", false},
	{"referer", false},
	{"referrer", false},
	{"source", true},
	{"utm_source", true},
}

var isDomain = regexp.MustCompile("^.*\\.[a-zA-Z]+$")

// Ignore returns whether a referrer should be ignored or not.
func Ignore(r *http.Request) bool {
	referrer := getFromHeaderOrQuery(r)

	if referrer == "" {
		return false
	}

	u, err := url.ParseRequestURI(referrer)

	if err == nil {
		referrer = u.Hostname()
	}

	referrer = stripSubdomain(referrer)
	_, found := hostnameBlacklist[referrer]

	// filter for bot keywords
	referrer = strings.ToLower(referrer)

	for _, botReferrer := range Blacklist {
		if strings.Contains(referrer, botReferrer) {
			return true
		}
	}

	return found
}

// Get returns the referrer for given request.
func Get(r *http.Request, ref, requestHostname string) (string, string, string) {
	referrer := ""

	if ref != "" {
		referrer = ref
	} else {
		referrer = getFromHeaderOrQuery(r)
	}

	if referrer == "" {
		return "", "", ""
	}

	if strings.HasPrefix(strings.ToLower(referrer), androidAppPrefix) {
		name, icon := androidAppCache.get(referrer)
		return referrer, name, icon
	}

	var u *url.URL
	var err error

	if strings.HasPrefix(strings.ToLower(referrer), "http") {
		referrer, err = url.QueryUnescape(referrer)

		if err == nil {
			u, err = url.ParseRequestURI(referrer)
		}
	} else if isDomain.MatchString(referrer) {
		u, err = url.ParseRequestURI(fmt.Sprintf("https://%s", referrer))
	} else {
		err = errors.New("not a URI")
	}

	if u == nil || err != nil {
		if isIP(referrer) {
			return "", "", ""
		}

		// accept non-url referrers (from utm_source for example)
		return "", strings.TrimSpace(referrer), ""
	}

	// the subdomain for requestHostname is already stripped at this point (any, not just www)
	hostname := util.StripWWW(strings.ToLower(u.Hostname()))

	if hostname == requestHostname {
		return "", "", ""
	}

	if u.Path == "/" {
		u.Path = ""
	}

	if isIP(hostname) {
		return "", "", ""
	}

	name := groups[hostname+u.Path]

	if name == "" {
		name = groups[hostname]

		if name == "" {
			name = hostname
		}
	}

	// remove query parameters and anchor
	u.Host = strings.ToLower(u.Host)
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), name, ""
}

func getFromHeaderOrQuery(r *http.Request) string {
	fromHeader := strings.TrimSpace(r.Header.Get("Referer"))

	for _, param := range QueryParams {
		referrer := r.URL.Query().Get(param.param)

		if referrer != "" && (!param.preferHeader || param.preferHeader && fromHeader == "") {
			return referrer
		}
	}

	return fromHeader
}

func isIP(referrer string) bool {
	referrer = strings.Trim(referrer, "/")

	if strings.Contains(referrer, ":") {
		var err error
		referrer, _, err = net.SplitHostPort(referrer)

		if err != nil {
			return false
		}
	}

	_, err := netip.ParseAddr(referrer)
	return err == nil
}

func stripSubdomain(hostname string) string {
	if hostname == "" {
		return ""
	}

	runes := []rune(hostname)
	index := len(runes) - 1
	dots := 0

	for i := index; i > 0; i-- {
		if runes[i] == '.' {
			dots++

			if dots == 2 {
				index++
				break
			}
		}

		index--
	}

	return hostname[index:]
}
