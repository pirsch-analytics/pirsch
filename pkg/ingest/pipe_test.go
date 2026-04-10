package ingest

import (
	"net/http"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestPipe(t *testing.T) {
	// create a storage mock and a basic pipeline
	storage := db.NewMock()
	pipe := NewPipe(PipeOptions{
		Storage: storage,
	}).Use(func(request *Request) (bool, error) {
		request.session = new(model.Session)
		return false, nil
	})

	// process a sample request
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
	assert.NoError(t, pipe.Process(&Request{
		Request: req,
	}))

	// stop the pipeline to flush the results
	pipe.Stop()
	assert.Len(t, storage.GetSessions(), 1)
	assert.Len(t, storage.GetPageViews(), 1)
}

// TODO test concurrency, timeout, buffer limit, retry on save
