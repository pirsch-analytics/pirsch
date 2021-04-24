package pirsch

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLanguage(t *testing.T) {
	input := []string{
		"",
		"  \t ",
		"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5",
		"en-us, en",
		"en-gb, en",
		"invalid",
	}
	expected := []string{
		"",
		"",
		"fr",
		"en",
		"en",
		"",
	}

	for i, in := range input {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", in)

		if lang := getLanguage(req); lang != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], lang)
		}
	}
}
