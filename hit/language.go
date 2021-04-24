package hit

import (
	iso6391 "github.com/emvi/iso-639-1"
	"net/http"
	"strings"
)

func getLanguage(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")

	if lang != "" {
		langs := strings.Split(lang, ";")
		parts := strings.Split(langs[0], ",")
		parts = strings.Split(parts[0], "-")
		code := strings.ToLower(strings.TrimSpace(parts[0]))

		if iso6391.ValidCode(code) {
			return code
		}
	}

	return ""
}
