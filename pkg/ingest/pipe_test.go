package ingest

import (
	"context"
	"errors"
	"math/rand/v2"
	"net/http"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestPipeSimple(t *testing.T) {
	// create a storage mock and a basic pipeline with a fake step updating the session
	storage := db.NewMock()
	pipe := NewPipe(PipeOptions{
		Storage: storage,
	}).Use(&sessionStep{})

	// process a sample request
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/?query=param", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	accepted, err := pipe.Process(&Request{Request: req})
	assert.NoError(t, err)
	assert.True(t, accepted)

	// stop the pipeline to flush the results
	pipe.Stop()
	assert.Len(t, storage.Sessions(), 1)
	assert.Len(t, storage.PageViews(), 1)
	assert.False(t, storage.PageViews()[0].Time.IsZero())
	requests := storage.Requests()
	assert.Len(t, requests, 1)
	assert.Equal(t, "query=param", requests[0].Query)
	assert.Equal(t, "gzip, deflate, br", requests[0].Headers["Accept-Encoding"])
	assert.Equal(t, "*/*", requests[0].Headers["Accept"])
	assert.Equal(t, "en-US,en;q=0.5", requests[0].Headers["Accept-Language"])
}

func TestPipeTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// create a pipeline with 5 seconds timeout
		storage := db.NewMock()
		pipe := NewPipe(PipeOptions{
			Storage:       storage,
			WorkerTimeout: time.Second * 5,
		}).Use(&sessionStep{})
		defer pipe.Stop()

		// create two sample requests
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		accepted, err := pipe.Process(&Request{Request: req})
		assert.NoError(t, err)
		assert.True(t, accepted)
		accepted, err = pipe.Process(&Request{Request: req})
		assert.NoError(t, err)
		assert.True(t, accepted)

		// nothing should have been stored immediately
		time.Sleep(time.Second * 3)
		synctest.Wait()
		assert.Empty(t, storage.Sessions())
		assert.Empty(t, storage.PageViews())

		// wait for the clock and check the result without stopping the pipeline
		time.Sleep(time.Second * 3)
		synctest.Wait()
		assert.Len(t, storage.Sessions(), 2)
		assert.Len(t, storage.PageViews(), 2)
	})
}

func TestPipeBufferLimit(t *testing.T) {
	// create a pipeline with one worker and a max channel size of 10 requests
	storage := db.NewMock()
	pipe := NewPipe(PipeOptions{
		Storage:          storage,
		Worker:           1,
		WorkerBufferSize: 10,
	}).Use(&sessionStep{})

	// ingest requests
	for range 11 {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		accepted, err := pipe.Process(&Request{Request: req})
		assert.NoError(t, err)
		assert.True(t, accepted)
	}

	// the buffer must have been flushed when it reached 10 requests
	assert.Len(t, storage.Sessions(), 10)
	assert.Len(t, storage.PageViews(), 10)

	// flush the remaining request
	pipe.Stop()
	assert.Len(t, storage.Sessions(), 11)
	assert.Len(t, storage.PageViews(), 11)
}

func TestPipeRetrySave(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// create a pipeline with failing storage
		storage := newStorageWithError(errors.New("error on save"))
		pipe := NewPipe(PipeOptions{
			Storage:       storage,
			Worker:        1,
			WorkerTimeout: time.Second * 5,
		}).Use(&sessionStep{})
		defer pipe.Stop()

		// create a sample request
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		accepted, err := pipe.Process(&Request{Request: req})
		assert.NoError(t, err)
		assert.True(t, accepted)

		// nothing must have been stored due to the storage error
		time.Sleep(time.Second * 6)
		synctest.Wait()
		assert.Empty(t, storage.Sessions())
		assert.Empty(t, storage.PageViews())

		// nothing must have been stored due to the storage error
		time.Sleep(time.Second * 30)
		synctest.Wait()
		assert.Empty(t, storage.Sessions())
		assert.Empty(t, storage.PageViews())

		// nothing must have been stored due to the storage error
		time.Sleep(time.Second * 50)
		synctest.Wait()
		assert.Empty(t, storage.Sessions())
		assert.Empty(t, storage.PageViews())

		// the data must have been stored after a successful retry
		storage.setErrorOnSave(nil)
		time.Sleep(time.Second * 66) // backup time + jitter
		synctest.Wait()
		assert.Len(t, storage.Sessions(), 1)
		assert.Len(t, storage.PageViews(), 1)
	})
}

func TestPipeRetryError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// create a pipeline with failing storage
		storage := newStorageWithError(errors.New("error on save"))
		pipe := NewPipe(PipeOptions{
			Storage:       storage,
			Worker:        1,
			WorkerTimeout: time.Second * 5,
		}).Use(&sessionStep{})

		// create a sample request
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		accepted, err := pipe.Process(&Request{Request: req})
		assert.NoError(t, err)
		assert.True(t, accepted)

		// nothing must have been stored after exhausting the maximum number of retries
		time.Sleep(time.Second * 130)
		synctest.Wait()
		assert.Empty(t, storage.Sessions())
		assert.Empty(t, storage.PageViews())

		// the requests must have been dropped, even after flushing
		pipe.Stop()
		assert.Empty(t, storage.Sessions())
		assert.Empty(t, storage.PageViews())

		// wait again to clean up background routines for asynchronous retries
		time.Sleep(time.Second * 130)
	})
}

func TestPipeConcurrency(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// create a pipeline with concurrent workers and max channel size
		storage := db.NewMock()
		pipe := NewPipe(PipeOptions{
			Storage:          storage,
			Worker:           5,
			WorkerBufferSize: 10,
			WorkerTimeout:    time.Second * 5,
		}).Use(&sessionStep{})

		// create sample requests concurrently
		var wg sync.WaitGroup

		for range 100 {
			wg.Go(func() {
				for range 100 {
					req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
					req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
					accepted, err := pipe.Process(&Request{Request: req})
					assert.NoError(t, err)
					assert.True(t, accepted)
					time.Sleep(time.Duration(rand.N(1000)) * time.Millisecond)
				}
			})
		}

		// a few requests must have been processed after a while
		time.Sleep(time.Duration(rand.N(10)+5) * time.Second)
		assert.NotZero(t, len(storage.Sessions()))
		assert.NotZero(t, len(storage.PageViews()))
		t.Log(len(storage.Sessions()), len(storage.PageViews()))

		// wait until all requests have been sent and stop the pipeline
		wg.Wait()
		pipe.Stop()
		assert.Len(t, storage.Sessions(), 10000)
		assert.Len(t, storage.PageViews(), 10000)
	})
}

func TestPipeOverrideTimeout(t *testing.T) {
	// reference time for comparison
	now := time.Now()

	// create a simple pipeline without sessions
	storage := db.NewMock()
	pipe := NewPipe(PipeOptions{
		Storage: storage,
	})

	// create two requests, one with the time set to 0, and one with the time set to an hour ago
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
	pastTime := time.Now().UTC().Add(-time.Hour)
	accepted, err := pipe.Process(&Request{Request: req, Time: pastTime})
	assert.NoError(t, err)
	assert.True(t, accepted)
	accepted, err = pipe.Process(&Request{Request: req})
	assert.NoError(t, err)
	assert.True(t, accepted)

	// one should have set now and one to the time that was passed for it
	pipe.Stop()
	pageViews := storage.PageViews()
	assert.Len(t, pageViews, 2)
	assert.Equal(t, pastTime, pageViews[0].Time)
	assert.True(t, pageViews[1].Time.After(now))
}

func TestPipeNoRequest(t *testing.T) {
	storage := db.NewMock()
	pipe := NewPipe(PipeOptions{
		Storage: storage,
	})
	accepted, err := pipe.Process(&Request{})
	assert.NoError(t, err)
	assert.False(t, accepted)
	pipe.Stop()
	assert.Empty(t, storage.PageViews())
}

func TestPipePrefetch(t *testing.T) {
	// list of pre-fetch headers
	header := []struct{ key, value string }{
		{"X-Moz", "prefetch"},
		{"X-Purpose", "prefetch"},
		{"X-Purpose", "preview"},
		{"Purpose", "prefetch"},
		{"Purpose", "preview"},
	}

	// create a simple pipeline without sessions
	storage := db.NewMock()
	pipe := NewPipe(PipeOptions{
		Storage: storage,
	})

	// create one request per pre-fetch header
	for _, h := range header {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		req.Header.Set(h.key, h.value)
		accepted, err := pipe.Process(&Request{Request: req})
		assert.NoError(t, err)
		assert.False(t, accepted)
	}

	// no requests must have been stored
	pipe.Stop()
	assert.Empty(t, storage.PageViews())
}

type sessionStep struct{}

func (s *sessionStep) Step(request *Request) (bool, error) {
	request.Session = new(model.Session)
	return false, nil
}

type storageWithError struct {
	db.Mock
	errorOnSave error
	m           sync.RWMutex
}

func newStorageWithError(err error) *storageWithError {
	return &storageWithError{
		Mock:        *db.NewMock(),
		errorOnSave: err,
	}
}

func (client *storageWithError) SavePageViews(_ context.Context, pageViews []model.PageView) error {
	client.m.RLock()
	defer client.m.RUnlock()

	if client.errorOnSave != nil {
		return client.errorOnSave
	}

	return client.Mock.SavePageViews(context.Background(), pageViews)
}

func (client *storageWithError) SaveSessions(_ context.Context, sessions []model.Session) error {
	client.m.RLock()
	defer client.m.RUnlock()

	if client.errorOnSave != nil {
		return client.errorOnSave
	}

	return client.Mock.SaveSessions(context.Background(), sessions)
}

func (client *storageWithError) setErrorOnSave(err error) {
	client.m.Lock()
	defer client.m.Unlock()
	client.errorOnSave = err
}
