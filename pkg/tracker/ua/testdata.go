package ua

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg"
)

type testUserAgent struct {
	ua             string
	browser        string
	browserVersion string
	os             string
	osVersion      string
}

// https://www.useragents.me/

var userAgentsEdge = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.74 Safari/537.36 Edg/79.0.309.43",
		browser:        pkg.BrowserEdge,
		browserVersion: "79.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 Edg/84.0.522.61",
		browser:        pkg.BrowserEdge,
		browserVersion: "84.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 Edg/84.0.522.44",
		browser:        pkg.BrowserEdge,
		browserVersion: "84.0",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; HD1913) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36 EdgA/45.6.2.5042",
		browser:        pkg.BrowserEdge,
		browserVersion: "45.6",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36 EdgA/45.6.2.5042",
		browser:        pkg.BrowserEdge,
		browserVersion: "45.6",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; Pixel 3 XL) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36 EdgA/45.6.2.5042",
		browser:        pkg.BrowserEdge,
		browserVersion: "45.6",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; ONEPLUS A6003) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36 EdgA/45.6.2.5042",
		browser:        pkg.BrowserEdge,
		browserVersion: "45.6",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 13_6_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0 EdgiOS/45.7.3 Mobile/15E148 Safari/605.1.15",
		browser:        pkg.BrowserEdge,
		browserVersion: "45.7",
		os:             pkg.OSiOS,
		osVersion:      "13.6",
	},
	{
		ua:             "Mozilla/5.0 (Windows Mobile 10; Android 10.0; Microsoft; Lumia 950XL) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Mobile Safari/537.36 Edge/40.15254.603",
		browser:        pkg.BrowserEdge,
		browserVersion: "40.15254",
		os:             pkg.OSWindowsMobile,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64; Xbox; Xbox One) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 Edge/44.18363.8131",
		browser:        pkg.BrowserEdge,
		browserVersion: "44.18363",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.70",
		browser:        pkg.BrowserEdge,
		browserVersion: "109.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
}

var userAgentsOpera = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 OPR/70.0.3728.119",
		browser:        pkg.BrowserOpera,
		browserVersion: "70.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 OPR/70.0.3728.119",
		browser:        pkg.BrowserOpera,
		browserVersion: "70.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 OPR/70.0.3728.119",
		browser:        pkg.BrowserOpera,
		browserVersion: "70.0",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 OPR/70.0.3728.119",
		browser:        pkg.BrowserOpera,
		browserVersion: "70.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; VOG-L29) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36 OPR/59.1.2926.54067",
		browser:        pkg.BrowserOpera,
		browserVersion: "59.1",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; SM-G970F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36 OPR/59.1.2926.54067",
		browser:        pkg.BrowserOpera,
		browserVersion: "59.1",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; SM-N975F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36 OPR/59.1.2926.54067",
		browser:        pkg.BrowserOpera,
		browserVersion: "59.1",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.115 Safari/537.36 OPR/27.0.1689.76",
		browser:        pkg.BrowserOpera,
		browserVersion: "27.0",
		os:             pkg.OSWindows,
		osVersion:      "8",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 OPR/94.0.0.0",
		browser:        pkg.BrowserOpera,
		browserVersion: "94.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
}

var userAgentsFirefox = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux i686; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (Linux x86_64; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 10_15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/28.0 Mobile/15E148 Safari/605.1.15",
		browser:        pkg.BrowserFirefox,
		browserVersion: "28.0",
		os:             pkg.OSiOS,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (iPad; CPU OS 10_15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/28.0 Mobile/15E148 Safari/605.1.15",
		browser:        pkg.BrowserFirefox,
		browserVersion: "28.0",
		os:             pkg.OSiOS,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (iPod touch; CPU iPhone OS 10_15_6 like Mac OS X) AppleWebKit/604.5.6 (KHTML, like Gecko) FxiOS/28.0 Mobile/15E148 Safari/605.1.15",
		browser:        pkg.BrowserFirefox,
		browserVersion: "28.0",
		os:             pkg.OSiOS,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (Android 10; Mobile; rv:68.0) Gecko/68.0 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Android 10; Mobile; LG-M255; rv:79.0) Gecko/79.0 Firefox/79.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "79.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "109.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/109.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "109.0",
		os:             pkg.OSLinux,
		osVersion:      "",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0",
		browser:        pkg.BrowserFirefox,
		browserVersion: "122.0",
		os:             pkg.OSLinux,
		osVersion:      "",
	},
	{
		ua:             `"Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0"`,
		browser:        pkg.BrowserFirefox,
		browserVersion: "122.0",
		os:             pkg.OSLinux,
		osVersion:      "",
	},
	{
		ua:             `\"Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0\"`,
		browser:        pkg.BrowserFirefox,
		browserVersion: "122.0",
		os:             pkg.OSLinux,
		osVersion:      "",
	},
}

var userAgentsChrome = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSLinux,
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; SM-A102U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; SM-G960U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; SM-N960U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; LM-Q720) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; LM-X420) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 10; LM-Q710(FGN)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "84.0",
		os:             pkg.OSAndroid,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3 like Mac OS X) AppleWebKit/602.1.50 (KHTML, like Gecko) CriOS/56.0.2924.75 Mobile/14E5239e Safari/602.1",
		browser:        pkg.BrowserChrome,
		browserVersion: "56.0",
		os:             pkg.OSiOS,
		osVersion:      "10.3",
	},
	{ // this can be optimized, but it's a fairly old Android version
		ua:        "Mozilla/5.0 (Linux; U; Android 4.1.1; en-gb; Build/KLP) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Safari/534.30",
		os:        pkg.OSAndroid,
		osVersion: "4.1",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 4.4; Nexus 5 Build/_BuildID_) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "30.0",
		os:             pkg.OSAndroid,
		osVersion:      "4.4",
	},
	{
		ua:             "Mozilla/5.0 (Linux; Android 5.1.1; Nexus 5 Build/LMY48B; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/43.0.2357.65 Mobile Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "43.0",
		os:             pkg.OSAndroid,
		osVersion:      "5.1",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "109.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "109.0",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "109.0",
		os:             pkg.OSLinux,
		osVersion:      "",
	},
	{
		ua:             "Mozilla/5.0 (X11; CrOS x86_64 15359.58.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.5615.134 Safari/537.36",
		browser:        pkg.BrowserChrome,
		browserVersion: "112.0",
		os:             pkg.OSChrome,
		osVersion:      "15359.58",
	},
}

var userAgentsSafari = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3 like Mac OS X) AppleWebKit/603.1.23 (KHTML, like Gecko) Version/10.0 Mobile/14E5239e Safari/602.1",
		browser:        pkg.BrowserSafari,
		browserVersion: "10.0",
		os:             pkg.OSiOS,
		osVersion:      "10.3",
	},
	{
		ua:             "Mozilla/5.0 (iPad; CPU OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5355d Safari/8536.25",
		browser:        pkg.BrowserSafari,
		browserVersion: "6.0",
		os:             pkg.OSiOS,
		osVersion:      "6.0",
	},
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5376e Safari/8536.25",
		browser:        pkg.BrowserSafari,
		browserVersion: "6.0",
		os:             pkg.OSiOS,
		osVersion:      "6.0",
	},
	{
		ua:        "Mozilla/5.0 (Windows; U; Windows NT 5.2; en-US) AppleWebKit/533.17.8 (KHTML, like Gecko) Version/5.0.1 Safari/533.17.8",
		os:        pkg.OSWindows,
		osVersion: "XP",
	},
	{
		ua:      "Mozilla/5.0 (Macintosh; U; PPC Mac OS X; sv-se) AppleWebKit/419 (KHTML, like Gecko) Safari/419.3",
		browser: pkg.BrowserSafari,
		os:      pkg.OSMac,
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Safari/605.1.15",
		browser:        pkg.BrowserSafari,
		browserVersion: "15.0",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Safari/605.1.15",
		browser:        pkg.BrowserSafari,
		browserVersion: "14.0",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 15_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Mobile/15E148 Safari/604.1",
		browser:        pkg.BrowserSafari,
		browserVersion: "15.1",
		os:             pkg.OSiOS,
		osVersion:      "15.1",
	},
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 14_4_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1",
		browser:        pkg.BrowserSafari,
		browserVersion: "14.0",
		os:             pkg.OSiOS,
		osVersion:      "14.4",
	},
	{
		ua:             "Mozilla/5.0 (iPhone; CPU iPhone OS 15_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Mobile/15E148 Safari/604.1",
		browser:        pkg.BrowserSafari,
		browserVersion: "15.4",
		os:             pkg.OSiOS,
		osVersion:      "15.4",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.2 Safari/605.1.15",
		browser:        pkg.BrowserSafari,
		browserVersion: "16.2",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Safari/605.1.15",
		browser:        pkg.BrowserSafari,
		browserVersion: "16.3",
		os:             pkg.OSMac,
		osVersion:      "10.15",
	},
}

var userAgentsIE = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Trident/7.0; rv:11.0) like Gecko",
		browser:        pkg.BrowserIE,
		browserVersion: "11.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0)",
		browser:        pkg.BrowserIE,
		browserVersion: "8.0",
		os:             pkg.OSWindows,
		osVersion:      "XP",
	},
	{
		ua:             "Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; WOW64; Trident/4.0;)",
		browser:        pkg.BrowserIE,
		browserVersion: "7.0",
		os:             pkg.OSWindows,
		osVersion:      "Vista",
	},
	{
		ua:             "Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0)",
		browser:        pkg.BrowserIE,
		browserVersion: "8.0",
		os:             pkg.OSWindows,
		osVersion:      "7",
	},
	{
		ua:             "Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1)",
		browser:        pkg.BrowserIE,
		browserVersion: "9.0",
		os:             pkg.OSWindows,
		osVersion:      "7",
	},
	{
		ua:             "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; WOW64; Trident/6.0)",
		browser:        pkg.BrowserIE,
		browserVersion: "10.0",
		os:             pkg.OSWindows,
		osVersion:      "7",
	},
	{
		ua:             "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2)",
		browser:        pkg.BrowserIE,
		browserVersion: "10.0",
		os:             pkg.OSWindows,
		osVersion:      "8",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko",
		browser:        pkg.BrowserIE,
		browserVersion: "11.0",
		os:             pkg.OSWindows,
		osVersion:      "7",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 6.2; Trident/7.0; rv:11.0) like Gecko",
		browser:        pkg.BrowserIE,
		browserVersion: "11.0",
		os:             pkg.OSWindows,
		osVersion:      "8",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
		browser:        pkg.BrowserIE,
		browserVersion: "11.0",
		os:             pkg.OSWindows,
		osVersion:      "8",
	},
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		browser:        pkg.BrowserIE,
		browserVersion: "11.0",
		os:             pkg.OSWindows,
		osVersion:      "10",
	},
}

var userAgentsArc = []testUserAgent{
	{
		ua:             "Arc/1.11.0 (Mac OS X Version 13.5.2 (Build 22G91))",
		browser:        pkg.BrowserArc,
		browserVersion: "1.11",
		os:             pkg.OSMac,
		osVersion:      "13.5",
	},
}

var userAgentsAll = mergeUserAgentLists(userAgentsEdge,
	userAgentsOpera,
	userAgentsFirefox,
	userAgentsChrome,
	userAgentsSafari,
	userAgentsIE,
	userAgentsArc)

func mergeUserAgentLists(ua ...[]testUserAgent) []testUserAgent {
	list := make([]testUserAgent, 0)

	for _, l := range ua {
		list = append(list, l...)
	}

	return list
}
