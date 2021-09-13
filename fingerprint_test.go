package pirsch

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFingerprint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test")
	req.RemoteAddr = "127.0.0.1:80"
	assert.Equal(t, "6d1f15a605abe4edbe53f18a57ce7654", Fingerprint(req, "salt"))
}
