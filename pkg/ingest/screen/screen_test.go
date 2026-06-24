package screen

import (
	"math/rand"
	"net/http"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestScreen(t *testing.T) {
	classes := make([]Class, len(Classes))
	copy(classes, Classes)

	for i := range classes {
		j := rand.Intn(i + 1)
		classes[i], classes[j] = classes[j], classes[i]
	}

	s := NewScreen(classes)
	assert.Equal(t, s.classes[0].MinWidth, Classes[0].MinWidth)
	assert.Equal(t, s.classes[len(s.classes)-1].MinWidth, Classes[len(Classes)-1].MinWidth)
}

func TestScreenWidth(t *testing.T) {
	input := []uint16{
		42,
		1024,
		1025,
		1919,
		2559,
		3839,
		5119,
		5120,
		0,
	}
	expected := []string{
		"XS",
		"XL",
		"XL",
		"HD",
		"Full HD",
		"WQHD",
		"UHD 4K",
		"UHD 5K",
		"",
	}
	s := NewScreen(Classes)

	for i, in := range input {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		r := &ingest.Request{
			Request:     req,
			ScreenWidth: in,
		}
		cancel, err := s.Step(r)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.Equal(t, expected[i], r.ScreenClass)
	}
}

func TestScreenHeader(t *testing.T) {
	input := []struct {
		header string
		value  string
	}{
		{"Sec-CH-Width", "1"},
		{"Sec-CH-Viewport-Width", "2"},
		{"Width", "3"},
		{"Viewport-Width", "4"},
	}
	expected := []uint16{1, 2, 3, 4}
	s := NewScreen(Classes)

	for i, in := range input {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		req.Header.Set(in.header, in.value)
		assert.Equal(t, expected[i], s.fromHeader(req, in.header))
	}
}
