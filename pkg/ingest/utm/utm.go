package utm

import (
	"net/http"
	"strings"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

type UTM struct{}

// NewUTM returns a new UTM.
func NewUTM() *UTM {
	return &UTM{}
}

// Step implements ingest.PipeStep to process a step.
// It sets the UTM parameters for the request.
func (u *UTM) Step(request *ingest.Request) (bool, error) {
	query := request.Request.URL.Query()
	request.UTMSource = strings.TrimSpace(query.Get("utm_source"))
	request.UTMMedium = u.medium(request.Request, request.ReferrerName)
	request.UTMCampaign = strings.TrimSpace(query.Get("utm_campaign"))
	request.UTMContent = strings.TrimSpace(query.Get("utm_content"))
	request.UTMTerm = strings.TrimSpace(query.Get("utm_term"))
	return false, nil
}

func (u *UTM) medium(r *http.Request, referrerName string) string {
	query := r.URL.Query()
	medium := strings.TrimSpace(query.Get("utm_medium"))

	if medium != "" {
		return medium
	}

	if referrerName == "Google" && query.Get("gclid") != "" {
		medium = "(gclid)"
	} else if referrerName == "Bing" && query.Get("msclkid") != "" {
		medium = "(msclkid)"
	}

	return medium
}
