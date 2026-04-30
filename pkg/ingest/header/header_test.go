package header

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	input := []map[string]string{
		{"User-Agent": ""},
		{"Accept-Language": ""},
		{"Accept-Encoding": ""},
		{"Connection": "close"},
		{"Connection": "keep-alive"},
		{"TE": "set"},
		{"Pragma": "set"},
		{"Pragma": "set", "Cache-Control": "set"},
		{"Sec-Fetch-Site": "empty", "Referer": ""},
		{"Sec-Fetch-Site": "empty", "Referer": "https://google.com"},
		{"Sec-Fetch-Site": "none", "Referer": "https://google.com"},
		{"Upgrade-Insecure-Requests": "1", "Sec-Fetch-Mode": "foo"},
		{"Upgrade-Insecure-Requests": "0", "Sec-Fetch-Mode": "cors"},
		{"Upgrade-Insecure-Requests": "1", "Sec-Fetch-Mode": "cors"},
		{"Sec-Fetch-Dest": "none", "Upgrade-Insecure-Requests": "1"},
		{"Sec-Fetch-Dest": "empty", "Upgrade-Insecure-Requests": "0"},
		{"Sec-Fetch-Dest": "empty", "Upgrade-Insecure-Requests": "1"},
		{"proto": "HTTP/1.1"},
		{"proto": "HTTP/1.1", "Sec-Fetch-Site": "something"},
		{"proto": "HTTP/1.1", "Sec-Fetch-Mode": "something"},
		{"proto": "HTTP/1.1", "Sec-Fetch-Dest": "something"},
	}
	results := []string{
		"ua-missing",
		"al-missing",
		"ae-missing",
		"http2-close",
		"http2-alive",
		"te",
		"pragma-cc",
		"",
		"",
		"",
		"sfs-referrer",
		"",
		"",
		"ui-cors",
		"",
		"",
		"sfd-ui",
		"",
		"http11-sf",
		"http11-sf",
		"http11-sf",
	}
	filter := NewHeader()

	for i, h := range input {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
		req.Proto = "HTTP/2.0"
		req.ProtoMajor = 2
		req.ProtoMinor = 0

		for k, v := range h {
			if k == "proto" {
				req.Proto = "HTTP/1.1"
				req.ProtoMajor = 1
				req.ProtoMinor = 1
			} else {
				req.Header.Set(k, v)
			}
		}

		r := &ingest.Request{
			Request: req,
		}
		ignore, err := filter.Step(r)
		assert.Nil(t, err)
		assert.Equalf(t, results[i] != "", ignore, fmt.Sprint(h))
		assert.Equal(t, results[i], r.BotReason)
	}
}
