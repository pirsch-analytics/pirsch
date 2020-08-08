package pirsch

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHitFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "user-agent")
	req.Header.Set("Referer", "ref")
	hit := hitFromRequest(req, "salt")

	if hit.Fingerprint == "" ||
		hit.Path != "/test/path" ||
		hit.URL != "/test/path?query=param&foo=bar#anchor" ||
		hit.Language != "de-de" ||
		hit.UserAgent != "user-agent" ||
		hit.Ref != "ref" ||
		hit.Time.IsZero() {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}
