package gad

import (
	"net/http"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestGAD(t *testing.T) {
	gad := NewGAD()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/?gad_source=Source&gad_campaignid=Campaign", nil)
	r := &ingest.Request{
		Request: req,
	}
	cancel, err := gad.Step(r)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "Source", r.UTMSource)
	assert.Equal(t, "Campaign", r.UTMCampaign)
}
