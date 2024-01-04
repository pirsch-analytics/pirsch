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
	URL          string
	Hostname     string
	Path         string
	Title        string
	Referrer     string
	ScreenWidth  uint16
	ScreenHeight uint16
	Time         time.Time

	// Tags are optional fields used to break down page views into segments.
	Tags map[string]string
}

func (options *Options) validate(r *http.Request) {
	if options.URL == "" {
		options.URL = r.URL.String()
	}

	u, err := url.ParseRequestURI(options.URL)

	if err == nil {
		options.Hostname = strings.ToLower(u.Hostname())

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
