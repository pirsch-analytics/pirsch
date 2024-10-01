package tracker

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Options are optional parameters for page views and events.
type Options struct {
	// URL is the full request URL.
	URL string

	// Hostname is used to check the Referrer. If it's the same as the hostname, it will be ignored.
	Hostname string

	// Path sets the path.
	Path string

	// Title sets the page title.
	Title string

	// Referrer overrides the referrer. If set to Hostname it will be ignored.
	Referrer string

	// ScreenWidth is the screen width which will be translated to a screen class.
	ScreenWidth uint16

	// ScreenHeight is the screen height which will be translated to a screen class.
	ScreenHeight uint16

	// Time overrides the time the page view should be recorded for.
	// Usually this is set to the time the request arrives at the Tracker.
	Time time.Time

	// Tags are optional fields used to break down page views into segments.
	Tags map[string]string
}

func (options *Options) validate(r *http.Request) {
	if options.URL == "" {
		options.URL = r.URL.String()
	}

	u, err := url.ParseRequestURI(options.URL)

	if err == nil {
		if options.Hostname == "" {
			options.Hostname = u.Hostname()
		}

		if options.Path != "" {
			// change path and re-assemble URL
			u.Path = options.Path
			options.URL = u.String()
		} else {
			options.Path = u.Path
		}
	}

	options.Title = util.ShortenString(options.Title, 512)
	options.Path = util.ShortenString(options.Path, 2000)

	if options.Path == "" {
		options.Path = "/"
	}
}

func (options *Options) getTags() ([]string, []string) {
	keys, values := make([]string, 0, len(options.Tags)), make([]string, 0, len(options.Tags))

	for k, v := range options.Tags {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)

		if k != "" && v != "" {
			keys = append(keys, k)
			values = append(values, v)
		}
	}

	return keys, values
}
