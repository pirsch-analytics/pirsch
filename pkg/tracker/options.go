package tracker

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"net/http"
	"net/url"
	"strconv"
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

// OptionsFromRequest returns Options for the client request.
func OptionsFromRequest(r *http.Request) Options {
	query := r.URL.Query()
	return Options{
		URL:          getURLQueryParam(query.Get("url")),
		Title:        strings.TrimSpace(query.Get("t")),
		Referrer:     strings.TrimSpace(query.Get("ref")),
		ScreenWidth:  getIntQueryParam[uint16](query.Get("w")),
		ScreenHeight: getIntQueryParam[uint16](query.Get("h")),
	}
}

func getIntQueryParam[T uint16 | uint64](param string) T {
	i, _ := strconv.ParseUint(param, 10, 64)
	return T(i)
}

func getURLQueryParam(param string) string {
	if _, err := url.ParseRequestURI(param); err != nil {
		return ""
	}

	return param
}
