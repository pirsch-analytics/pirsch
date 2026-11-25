package gad

import (
	"strings"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

// GAD maps query parameters to gad_* parameters.
// They overwrite the UTM source and campaign parameters.
type GAD struct{}

// NewGAD returns a new GAD.
func NewGAD() *GAD {
	return &GAD{}
}

// Step implements ingest.PipeStep to process a step.
// It sets the gad_* parameters for the request.
func (u *GAD) Step(request *ingest.Request) (bool, error) {
	query := request.Request.URL.Query()

	if request.UTMSource == "" {
		request.UTMSource = strings.TrimSpace(query.Get("gad_source"))
	}

	if request.UTMCampaign == "" {
		request.UTMCampaign = strings.TrimSpace(query.Get("gad_campaignid"))
	}

	return false, nil
}
