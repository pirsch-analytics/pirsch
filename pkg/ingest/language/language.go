package language

import (
	"strings"

	"github.com/emvi/iso-639-1"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

// Language extracts and sets the language.
type Language struct{}

// NewLanguage returns a new Language.
func NewLanguage() *Language {
	return &Language{}
}

// Step implements ingest.PipeStep to process a step.
// It sets the screen class for the request.
func (l *Language) Step(request *ingest.Request) (bool, error) {
	lang := request.Request.Header.Get("Accept-Language")

	if lang != "" {
		left, _, _ := strings.Cut(lang, ";")
		left, _, _ = strings.Cut(left, ",")
		left, _, _ = strings.Cut(left, "-")
		code := strings.ToLower(strings.TrimSpace(left))

		if iso6391.ValidCode(code) {
			request.Language = code
		}
	}

	return false, nil
}
