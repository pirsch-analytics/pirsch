package language

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestLanguage(t *testing.T) {
	input := []string{
		"",
		"  \t ",
		"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5",
		"en-us, en",
		"en-gb, en",
		"invalid",
	}
	expected := []string{
		"",
		"",
		"fr",
		"en",
		"en",
		"",
	}
	tracker := NewLanguage()

	for i, in := range input {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", in)
		r := &ingest.Request{
			Request: req,
		}
		cancel, err := tracker.Step(r)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.Equal(t, expected[i], r.Language)
	}
}
