package pirsch

import (
	"net/http/httptest"
	"testing"
)

func TestParseForwardedHeader(t *testing.T) {
	header := []string{
		"for=12.34.56.78;host=example.com;proto=https, for=23.45.67.89",
		"for=12.34.56.78, for=23.45.67.89;secret=egah2CGj55fSJFs, for=10.1.2.3",
		"for=192.0.2.60;proto=http;by=203.0.113.43",
		"proto=http;by=203.0.113.43;for=192.0.2.61",
		"   ",
		"",
	}
	expected := []string{
		"12.34.56.78",
		"12.34.56.78",
		"192.0.2.60",
		"192.0.2.61",
		"",
		"",
	}

	for i, head := range header {
		if ip := parseForwardedHeader(head); ip != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], ip)
		}
	}
}

func TestParseXForwardedForHeader(t *testing.T) {
	header := []string{
		"127.0.0.1",
		"127.0.0.1, 23.21.45.67",
		"127.0.0.1,23.21.45.67",
		"   ",
		"",
	}
	expected := []string{
		"127.0.0.1",
		"127.0.0.1",
		"127.0.0.1",
		"",
		"",
	}

	for i, head := range header {
		if ip := parseXForwardedForHeader(head); ip != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], ip)
		}
	}
}

func TestGetIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "123.456.789.012:29302"

	// no header, default
	if ip := getIP(r); ip != "123.456.789.012" {
		t.Fatalf("Expected '123.456.789.012', but was: %v", ip)
	}

	// X-Real-IP
	r.Header.Set("X-Real-IP", "103.0.53.43")

	if ip := getIP(r); ip != "103.0.53.43" {
		t.Fatalf("Expected '103.0.53.43', but was: %v", ip)
	}

	// Forwarded
	r.Header.Set("Forwarded", "for=192.0.2.60;proto=http;by=203.0.113.43")

	if ip := getIP(r); ip != "192.0.2.60" {
		t.Fatalf("Expected '192.0.2.60', but was: %v", ip)
	}

	// X-Forwarded-For
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 23.21.45.67")

	if ip := getIP(r); ip != "127.0.0.1" {
		t.Fatalf("Expected '127.0.0.1', but was: %v", ip)
	}

	// CF-Connecting-IP
	r.Header.Set("CF-Connecting-IP", "127.0.0.1, 23.21.45.67")

	if ip := getIP(r); ip != "127.0.0.1" {
		t.Fatalf("Expected '127.0.0.1', but was: %v", ip)
	}
}
