package referrer

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"strings"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/util"
)

var (
	isDomain = regexp.MustCompile("^.*\\.[a-zA-Z]+$")
)

// Referrer maps referrers to groups and filters bot requests based on the referrer.
type Referrer struct {
	groups map[string]string
}

// NewReferrer returns a new Referrer for the given group list.
func NewReferrer(groups map[string]string) *Referrer {
	return &Referrer{
		groups: groups,
	}
}

// Step implements ingest.PipeStep to process a step.
// It sets the referrer for the request.
func (r *Referrer) Step(request *ingest.Request) (bool, error) {
	referrer := ""

	if request.Referrer != "" {
		referrer = request.Referrer
	} else {
		referrer = fromHeaderOrQuery(request.Request)
	}

	if referrer == "" {
		r.unset(request)
		return false, nil
	}

	if strings.HasPrefix(strings.ToLower(referrer), androidAppPrefix) {
		name, icon := androidAppCache.get(referrer)
		request.Referrer = util.Shorten(referrer, 200)
		request.ReferrerName = util.Shorten(name, 200)
		request.ReferrerIcon = util.Shorten(icon, 2000)
		return false, nil
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
		if r.isIP(referrer) {
			r.unset(request)
			return false, nil
		}

		// accept non-url referrers (from utm_source, for example)
		r.unset(request)
		request.ReferrerName = util.Shorten(strings.TrimSpace(referrer), 200)
		return false, nil
	}

	// the subdomain for requestHostname is already stripped at this point (any, not just www)
	hostname := util.StripWWW(strings.ToLower(u.Hostname()))

	if hostname == request.Hostname {
		r.unset(request)
		return false, nil
	}

	if u.Path == "/" {
		u.Path = ""
	}

	if r.isIP(hostname) {
		return false, nil
	}

	name := r.groups[hostname+u.Path]

	if name == "" {
		name = r.groups[hostname]

		if name == "" {
			name = hostname
		}
	}

	// remove query parameters and anchor
	u.Host = strings.ToLower(u.Host)
	u.RawQuery = ""
	u.Fragment = ""
	request.Referrer = util.Shorten(u.String(), 200)
	request.ReferrerName = util.Shorten(name, 200)
	request.ReferrerIcon = ""
	return false, nil
}

func (r *Referrer) unset(request *ingest.Request) {
	request.Referrer = ""
	request.ReferrerName = ""
	request.ReferrerIcon = ""
}

func (r *Referrer) isIP(referrer string) bool {
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
