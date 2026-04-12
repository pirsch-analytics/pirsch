package ingest

import (
	"context"
	"errors"
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

func TestPipeTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// create a pipeline with 5 seconds timeout
		storage := db.NewMock()
		pipe := NewPipe(PipeOptions{
			Storage:       storage,
			WorkerTimeout: time.Second * 5,
		}).Use(func(request *Request) (bool, error) {
			request.session = new(model.Session)
			return false, nil
		})
		defer pipe.Stop()

		// create two sample requests
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		assert.NoError(t, pipe.Process(&Request{
			Request: req,
		}))
		assert.NoError(t, pipe.Process(&Request{
			Request: req,
		}))

		// nothing should have been stored immediately
		time.Sleep(time.Second * 3)
		synctest.Wait()
		assert.Empty(t, storage.GetSessions())
		assert.Empty(t, storage.GetPageViews())

		// wait for the clock and check the result without stopping the pipeline
		time.Sleep(time.Second * 3)
		synctest.Wait()
		assert.Len(t, storage.GetSessions(), 2)
		assert.Len(t, storage.GetPageViews(), 2)
	})
}

func TestPipeBufferLimit(t *testing.T) {
	// create a pipeline with one worker and a max channel size of 10 requests
	storage := db.NewMock()
	pipe := NewPipe(PipeOptions{
		Storage:          storage,
		Worker:           1,
		WorkerBufferSize: 10,
	}).Use(func(request *Request) (bool, error) {
		request.session = new(model.Session)
		return false, nil
	})

	// ingest requests
	for range 11 {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		assert.NoError(t, pipe.Process(&Request{
			Request: req,
		}))
	}

	// the buffer must have been flushed when it reached 10 requests
	assert.Len(t, storage.GetSessions(), 10)
	assert.Len(t, storage.GetPageViews(), 10)

	// flush the remaining request
	pipe.Stop()
	assert.Len(t, storage.GetSessions(), 11)
	assert.Len(t, storage.GetPageViews(), 11)
}

func TestPipeRetrySave(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// create a pipeline with failing storage
		storage := newStorageWithError(errors.New("error on save"))
		pipe := NewPipe(PipeOptions{
			Storage:       storage,
			Worker:        1,
			WorkerTimeout: time.Second * 5,
		}).Use(func(request *Request) (bool, error) {
			request.session = new(model.Session)
			return false, nil
		})
		defer pipe.Stop()

		// create a sample request
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		assert.NoError(t, pipe.Process(&Request{
			Request: req,
		}))

		// nothing must have been stored due to the storage error
		time.Sleep(time.Second * 6)
		synctest.Wait()
		assert.Empty(t, storage.GetSessions())
		assert.Empty(t, storage.GetPageViews())

		// nothing must have been stored due to the storage error
		time.Sleep(time.Second * 30)
		synctest.Wait()
		assert.Empty(t, storage.GetSessions())
		assert.Empty(t, storage.GetPageViews())

		// nothing must have been stored due to the storage error
		time.Sleep(time.Second * 50)
		synctest.Wait()
		assert.Empty(t, storage.GetSessions())
		assert.Empty(t, storage.GetPageViews())

		// the data must have been stored after a successful retry
		storage.setErrorOnSave(nil)
		time.Sleep(time.Second * 66) // backup time + jitter
		synctest.Wait()
		assert.Len(t, storage.GetSessions(), 1)
		assert.Len(t, storage.GetPageViews(), 1)
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
		}).Use(func(request *Request) (bool, error) {
			request.session = new(model.Session)
			return false, nil
		})

		// create a sample request
		req, _ := http.NewRequest(http.MethodGet, "https://example.com/", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		assert.NoError(t, pipe.Process(&Request{
			Request: req,
		}))

		// nothing must have been stored after exhausting the maximum number of retries
		time.Sleep(time.Second * 130)
		synctest.Wait()
		assert.Empty(t, storage.GetSessions())
		assert.Empty(t, storage.GetPageViews())

		// the requests must have been dropped even after flushing
		pipe.Stop()
		assert.Empty(t, storage.GetSessions())
		assert.Empty(t, storage.GetPageViews())
	})
}

// TODO test concurrency

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
