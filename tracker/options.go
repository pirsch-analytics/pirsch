package tracker

import (
	"github.com/pirsch-analytics/pirsch/v4/util"
	"net/http"
	"net/url"
)

// Options are optional parameters for page views and events.
type Options struct {
	URL          string
	Path         string
	Title        string
	Referrer     string
	ScreenWidth  uint16
	ScreenHeight uint16
}

func (options *Options) validate(r *http.Request) {
	if options.URL == "" {
		options.URL = r.URL.String()
	}

	u, err := url.ParseRequestURI(options.URL)

	if err == nil {
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
