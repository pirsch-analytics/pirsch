package utm

import (
	"net/http"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestUTM(t *testing.T) {
	utm := NewUTM()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/?utm_source=Source&utm_medium=Medium&utm_campaign=Campaign&utm_content=Content&utm_term=Term", nil)
	r := &ingest.Request{
		Request: req,
	}
	cancel, err := utm.Step(r)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "Source", r.UTMSource)
	assert.Equal(t, "Medium", r.UTMMedium)
	assert.Equal(t, "Campaign", r.UTMCampaign)
	assert.Equal(t, "Content", r.UTMContent)
	assert.Equal(t, "Term", r.UTMTerm)
}

func TestUTMGclid(t *testing.T) {
	utm := NewUTM()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/?gclid=1234", nil)
	r := &ingest.Request{
		Request:      req,
		ReferrerName: "Google",
	}
	cancel, err := utm.Step(r)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "(gclid)", r.UTMMedium)
}

func TestUTMMsclkid(t *testing.T) {
	utm := NewUTM()
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/?msclkid=1234", nil)
	r := &ingest.Request{
		Request:      req,
		ReferrerName: "Bing",
	}
	cancel, err := utm.Step(r)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "(msclkid)", r.UTMMedium)
}
