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
	assert.Equal(t, "2a8693dce9885db1cd837bd9f36acabb", Fingerprint(req, "salt"))
}
