package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseUserAgent(t *testing.T) {
	// just a simple test to check ParseUserAgent returns something for a clean User-Agent
	ua := ParseUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0")
	assert.Equal(t, OSMac, ua.OS)
	assert.Equal(t, "10.15", ua.OSVersion)
	assert.Equal(t, BrowserFirefox, ua.Browser)
	assert.Equal(t, "79.0", ua.BrowserVersion)
}

func TestGetBrowser(t *testing.T) {
	for _, ua := range userAgentsAll {
		system, products := parseUserAgent(ua.ua)
		browser, version := getBrowser(products, system, ua.os)
		assert.Equal(t, ua.browser, browser)
		assert.Equal(t, ua.browserVersion, version)
	}
}

func TestGetBrowserChromeSafari(t *testing.T) {
	chrome := "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"
	system, products := parseUserAgent(chrome)
	browser, version := getBrowser(products, system, OSMac)
	assert.Equal(t, BrowserChrome, browser)
	assert.Equal(t, "87.0", version)
	safari := "AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.1 Safari/605.1.15"
	system, products = parseUserAgent(safari)
	browser, version = getBrowser(products, system, OSMac)
	assert.Equal(t, BrowserSafari, browser)
	assert.Equal(t, "14.0", version)
}

func TestGetOS(t *testing.T) {
	for _, ua := range userAgentsAll {
		system, _ := parseUserAgent(ua.ua)
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
		system, products := parseUserAgent(in)
		assert.ElementsMatch(t, expected[i][0], system)
		assert.ElementsMatch(t, expected[i][1], products)
	}
}
