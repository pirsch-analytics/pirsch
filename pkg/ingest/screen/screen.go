package screen

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

type Screen struct {
	classes []Class
}

// NewScreen returns a new Screen for the given list of size classifications.
func NewScreen(classes []Class) *Screen {
	// copy and sort classes so that the original stays unchanged
	c := make([]Class, len(classes))
	copy(c, classes)
	slices.SortFunc(c, func(a, b Class) int {
		if a.MinWidth < b.MinWidth {
			return 1
		} else if a.MinWidth > b.MinWidth {
			return -1
		}

		return 0
	})
	return &Screen{
		classes: c,
	}
}

// Step implements ingest.PipeStep to process a step.
// It sets the screen class for the request.
func (s *Screen) Step(request *ingest.Request) (bool, error) {
	if request.ScreenWidth == 0 {
		request.ScreenWidth = s.fromHeader(request.Request, "Sec-CH-Width")

		if request.ScreenWidth == 0 {
			request.ScreenWidth = s.fromHeader(request.Request, "Sec-CH-Viewport-Width")
		}

		if request.ScreenWidth == 0 {
			request.ScreenWidth = s.fromHeader(request.Request, "Width")
		}

		if request.ScreenWidth == 0 {
			request.ScreenWidth = s.fromHeader(request.Request, "Viewport-Width")
		}
	}

	for _, class := range s.classes {
		if request.ScreenWidth >= class.MinWidth {
			request.ScreenClass = class.Class
			break
		}
	}

	return false, nil
}

func (s *Screen) fromHeader(r *http.Request, header string) uint16 {
	h := r.Header.Get(header)

	if h != "" {
		w, err := strconv.Atoi(h)

		if err == nil && w > 0 {
			return uint16(w)
		}
	}

	return 0
}
