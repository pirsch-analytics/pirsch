package pirsch

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFingerprint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test")
	req.RemoteAddr = "127.0.0.1:80"
	hash := md5.New()

	if _, err := io.WriteString(hash, "test127.0.0.1"+time.Now().UTC().Format("20060102")+"salt"); err != nil {
		t.Fatal(err)
	}

	fp := hex.EncodeToString(hash.Sum(nil))

	if out := Fingerprint(req, "salt"); out != fp {
		t.Fatalf("Fingerprint '%v' not as expected '%v'", out, fp)
	}
}
