package hit

import (
	"net/http"
	"strings"
)

type utmParams struct {
	source   string
	medium   string
	campaign string
	content  string
	term     string
}

func getUTMParams(r *http.Request) utmParams {
	query := r.URL.Query()
	return utmParams{
		source:   strings.TrimSpace(query.Get("utm_source")),
		medium:   strings.TrimSpace(query.Get("utm_medium")),
		campaign: strings.TrimSpace(query.Get("utm_campaign")),
		content:  strings.TrimSpace(query.Get("utm_content")),
		term:     strings.TrimSpace(query.Get("utm_term")),
	}
}
