package pirsch

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHitFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7,fr;q=0.6,nb;q=0.5,la;q=0.4")
	req.Header.Set("User-Agent", "user-agent")
	req.Header.Set("Referer", "ref")
	hit := hitFromRequest(req, "salt", &HitOptions{
		TenantID: sql.NullInt64{Int64: 42, Valid: true},
	})

	if hit.TenantID.Int64 != 42 ||
		!hit.TenantID.Valid ||
		hit.Fingerprint == "" ||
		hit.Path != "/test/path" ||
		hit.URL != "/test/path?query=param&foo=bar#anchor" ||
		hit.Language != "de-de" ||
		hit.UserAgent != "user-agent" ||
		hit.Ref != "ref" ||
		hit.Time.IsZero() {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}

func TestHitFromRequestPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path?query=param&foo=bar#anchor", nil)
	hit := hitFromRequest(req, "salt", &HitOptions{
		Path: "/new/custom/path",
	})

	if hit.Path != "/new/custom/path" ||
		hit.URL != "/new/custom/path?query=param&foo=bar#anchor" {
		t.Fatalf("Hit not as expected: %v", hit)
	}
}
