package integration

import "net/http"

type requestOptions struct {
	URL            string
	IP             *string
	UserAgent      *string
	AcceptLanguage *string
	AcceptEncoding *string
	Referrer       *string
}

func newRequest(options requestOptions) *http.Request {
	if options.URL == "" {
		options.URL = "https://example.com/?utm_source=Source&utm_medium=Medium&utm_campaign=Campaign&utm_content=Content&utm_term=Term"
	}

	req, _ := http.NewRequest(http.MethodGet, options.URL, nil)

	if options.IP != nil {
		req.RemoteAddr = *options.IP
	} else {
		req.RemoteAddr = "81.2.69.142"
	}

	if options.UserAgent != nil {
		req.Header.Set("User-Agent", *options.UserAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
	}

	if options.AcceptLanguage != nil {
		req.Header.Set("Accept-Language", *options.AcceptLanguage)
	} else {
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	}

	if options.AcceptEncoding != nil {
		req.Header.Set("Accept-Encoding", *options.AcceptEncoding)
	} else {
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}

	if options.Referrer != nil {
		req.Header.Set("Referer", *options.Referrer)
	} else {
		req.Header.Set("Referer", "https://google.com")
	}

	return req
}
