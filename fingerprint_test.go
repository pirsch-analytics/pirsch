package pirsch

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFingerprint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test")
	req.RemoteAddr = "127.0.0.1:80"
	hash := md5.New()
	_, err := io.WriteString(hash, "test127.0.0.1salt")
	assert.NoError(t, err)
	fp := hex.EncodeToString(hash.Sum(nil))
	assert.Equal(t, fp, Fingerprint(req, "salt"))
}
