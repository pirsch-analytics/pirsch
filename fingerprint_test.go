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

func TestGetIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "123.456.789.123:98765"

	if ip := getIP(req); ip != "123.456.789.123" {
		t.Fatalf("IP not as expected: %v", ip)
	}
}

func TestGetIPXForwarededFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "123.456.789.123:98765"
	header := []string{
		"127.0.0.1",
		"127.0.0.1,23.21.45.67",
		"",
	}
	expected := []string{
		"127.0.0.1",
		"127.0.0.1",
		"123.456.789.123",
	}

	for i, head := range header {
		req.Header.Set("X-Forwarded-For", head)

		if ip := getIP(req); ip != expected[i] {
			t.Fatalf("Expected IP '%v', but was: %v", expected[i], ip)
		}
	}
}

func TestGetIPForwareded(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "123.456.789.123:98765"
	header := []string{
		`for="_gazonk"`,
		`For="[2001:db8:cafe::17]:4711"`,
		"for=192.0.2.60;proto=http;by=203.0.113.43",
		"for=192.0.2.43, for=198.51.100.17",
		"",
	}
	expected := []string{
		`for="_gazonk"`,
		`For="[2001:db8:cafe::17]:4711"`,
		"for=192.0.2.60",
		"for=192.0.2.43",
		"123.456.789.123",
	}

	for i, head := range header {
		req.Header.Set("Forwarded", head)

		if ip := getIP(req); ip != expected[i] {
			t.Fatalf("Expected IP '%v', but was: %v", expected[i], ip)
		}
	}
}
