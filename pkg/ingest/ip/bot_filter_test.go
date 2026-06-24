package ip

import (
	"net/http"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestBotFilter(t *testing.T) {
	list := NewList()
	list.Update([]string{
		"123.128.0.12",
	}, nil, nil, nil, nil, nil)
	f := NewBotFilter([]Filter{list})
	r, _ := http.NewRequest("GET", "/", nil)
	req := &ingest.Request{Request: r, IP: "123.128.0.12"}

	// ignore request
	cancel, err := f.Step(req)
	assert.True(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "123.128.0.12", req.IP)
	assert.Equal(t, "ip", req.BotReason)

	// do not ignore request
	r, _ = http.NewRequest("GET", "/", nil)
	req = &ingest.Request{
		Request:          r,
		IP:               "123.128.0.12",
		DisableBotFilter: true,
	}
	cancel, err = f.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "123.128.0.12", req.IP)
	assert.Empty(t, req.BotReason)

	// do not ignore request
	r, _ = http.NewRequest("GET", "/", nil)
	req = &ingest.Request{Request: r, IP: "123.128.0.13"}
	cancel, err = f.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "123.128.0.13", req.IP)
	assert.Empty(t, req.BotReason)
}
