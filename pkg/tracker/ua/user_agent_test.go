package ua

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestParseSimple(t *testing.T) {
	// just a simple test to check Parse returns something for a clean User-Agent
	uaString := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0"
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", uaString)
	req.Header.Set("Sec-CH-UA-Mobile", "?1")
	ua := Parse(req)
	assert.InDelta(t, time.Now().UTC().UnixMilli(), ua.Time.UnixMilli(), 10)
	assert.Equal(t, uaString, ua.UserAgent)
	assert.Equal(t, pkg.OSMac, ua.OS)
	assert.Equal(t, "10.15", ua.OSVersion)
	assert.Equal(t, pkg.BrowserFirefox, ua.Browser)
	assert.Equal(t, "79.0", ua.BrowserVersion)
	assert.True(t, ua.Mobile.Bool)
}

func TestGetBrowser(t *testing.T) {
	for _, ua := range userAgentsAll {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", ua.ua)
		system, products, _, _ := parse(req)
		browser, version := getBrowser(products, system, ua.os)
		assert.Equal(t, ua.browser, browser)
		assert.Equal(t, ua.browserVersion, version)
	}
}

func TestGetBrowserChromeSafari(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	system, products, _, _ := parse(req)
	browser, version := getBrowser(products, system, pkg.OSMac)
	assert.Equal(t, pkg.BrowserChrome, browser)
	assert.Equal(t, "87.0", version)
	req.Header.Set("User-Agent", "AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.1 Safari/605.1.15")
	system, products, _, _ = parse(req)
	browser, version = getBrowser(products, system, pkg.OSMac)
	assert.Equal(t, pkg.BrowserSafari, browser)
	assert.Equal(t, "14.0", version)
}

func TestGetOS(t *testing.T) {
	for _, ua := range userAgentsAll {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", ua.ua)
		system, _, _, _ := parse(req)
		os, version := getOS(system)
		assert.Equal(t, ua.os, os)
		assert.Equal(t, ua.osVersion, version)
	}
}

func TestGetProductVersion(t *testing.T) {
	input := []struct {
		product string
		n       int
	}{
		{"", 0},
		{"", 1},
		{"", 42},
		{"   ", 0},
		{"Edg", 0},
		{"Edg/", 0},
		{"Edg/   ", 0},
		{"Safari/537.36", 0},
		{"Edg/79.0.309.43", 1},
		{"Chrome/79.0.3945.74", 2},
		{"Chrome/79.0.3945.74", 10},
	}
	expected := []string{
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		"537",
		"79.0",
		"79.0.3945",
		"79.0.3945.74",
	}

	for i, in := range input {
		assert.Equal(t, expected[i], getProductVersion(in.product, in.n))
	}
}

func TestGetOSVersion(t *testing.T) {
	input := []struct {
		version string
		n       int
	}{
		{"", 0},
		{"", 1},
		{"", 42},
		{"   ", 0},
		{"10.0", 0},
		{"10.0", 1},
		{"10.15.6", 2},
		{"10.15.6", 42},
	}
	expected := []string{
		"",
		"",
		"",
		"",
		"10",
		"10.0",
		"10.15.6",
		"10.15.6",
	}

	for i, in := range input {
		assert.Equal(t, expected[i], getOSVersion(in.version, in.n))
	}
}

func TestParseClientHints(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/116.0")
	req.Header.Set("Sec-CH-UA-Platform", pkg.OSChrome)
	req.Header.Set("Sec-CH-UA-Platform-Version", "6.4.10")
	ua := Parse(req)
	assert.Equal(t, pkg.OSChrome, ua.OS)
	assert.Equal(t, "6.4", ua.OSVersion)
	req.Header.Set("Sec-CH-UA-Platform", "Unknown")
	ua = Parse(req)
	assert.Equal(t, pkg.OSLinux, ua.OS)
	assert.Empty(t, ua.OSVersion)
	req.Header.Set("Sec-CH-UA-Platform", "Does not exist")
	ua = Parse(req)
	assert.Empty(t, ua.OS)
	assert.Empty(t, ua.OSVersion)
	req.Header.Set("Sec-CH-UA-Platform", "Windows")
	req.Header.Set("Sec-CH-UA-Platform-Version", "13.0.0")
	ua = Parse(req)
	assert.Equal(t, pkg.OSWindows, ua.OS)
	assert.Equal(t, "11", ua.OSVersion)
}

func TestParse(t *testing.T) {
	input := []string{
		// empty
		"",
		"  ",
		"'  '",
		` "   "`,

		// clean and simple
		"(system)",
		"version",
		"(system) version",

		// whitespace
		"   (system)   ",
		"   version    ",
		"   (   system   )   version   ",
		"   (  ;  system    ;  )   version   ",

		// multiple system entries and versions
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	}
	expected := [][][]string{
		{{}, {}},
		{{}, {}},
		{{}, {}},
		{{}, {}},
		{{"system"}, {}},
		{{}, {"version"}},
		{{"system"}, {"version"}},
		{{"system"}, {}},
		{{}, {"version"}},
		{{"system"}, {"version"}},
		{{"system"}, {"version"}},
		{{"Macintosh", "Intel Mac OS X 10_10_5"}, {"Chrome/63.0.3239.132", "Safari/537.36"}},
	}

	for i, in := range input {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", in)
		system, products, systemFromCH, productFromCH := parse(req)
		assert.ElementsMatch(t, expected[i][0], system)
		assert.ElementsMatch(t, expected[i][1], products)
		assert.False(t, systemFromCH)
		assert.False(t, productFromCH)
	}
}
