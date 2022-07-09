package utm

import (
	"net/http"
	"strings"
)

// Params groups UTM query parameters.
type Params struct {
	Source   string
	Medium   string
	Campaign string
	Content  string
	Term     string
}

// Get returns the UTM parameters for given request.
func Get(r *http.Request) Params {
	query := r.URL.Query()
	return Params{
		Source:   strings.TrimSpace(query.Get("utm_source")),
		Medium:   strings.TrimSpace(query.Get("utm_medium")),
		Campaign: strings.TrimSpace(query.Get("utm_campaign")),
		Content:  strings.TrimSpace(query.Get("utm_content")),
		Term:     strings.TrimSpace(query.Get("utm_term")),
	}
}
