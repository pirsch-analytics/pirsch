package tracker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFingerprint(t *testing.T) {
	SetFingerprintKeys(42, 99)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test")
	req.RemoteAddr = "127.0.0.1:80"
	assert.Equal(t, uint64(0x4f97de4b2cbf6e12), Fingerprint(req, "salt", nil, nil))
}
