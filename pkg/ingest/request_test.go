package ingest

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/some/path", nil)
	r := Request{
		Request:   req,
		EventName: " test ",
	}
	r.validate()
	assert.InDelta(t, time.Now().UTC().UnixMilli(), r.Time.UnixMilli(), 100)
	assert.Equal(t, "/some/path", r.Path)
	assert.Equal(t, "example.com", r.Hostname)
	assert.Equal(t, "test", r.EventName)
}

func TestRequestOverrideTime(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/some/path", nil)
	overrideTime := time.Now().UTC().Add(-time.Minute)
	r := Request{
		Request: req,
		Time:    overrideTime,
	}
	r.validate()
	assert.Equal(t, overrideTime, r.Time)
}

func TestRequestOverridePath(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/some/path", nil)
	r := Request{
		Request: req,
		Path:    "/override/path",
	}
	r.validate()
	assert.Equal(t, "/override/path", r.Path)
}

func TestRequestDefaultPath(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	r := Request{
		Request: req,
	}
	r.validate()
	assert.Equal(t, "/", r.Path)
}
