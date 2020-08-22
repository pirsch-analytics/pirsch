package pirsch

type testUserAgent struct {
	ua             string
	browser        string
	browserVersion string
	os             string
	osVersion      string
}

var userAgentsEdge = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.74 Safari/537.36 Edg/79.0.309.43",
		browser:        BrowserEdge,
		browserVersion: "79.0",
		os:             OSWindows,
		osVersion:      "10",
	},
}

var userAgentsOpera = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 OPR/70.0.3728.119",
		browser:        BrowserOpera,
		browserVersion: "70.0",
		os:             OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 OPR/70.0.3728.119",
		browser:        BrowserOpera,
		browserVersion: "70.0",
		os:             OSMac,
		osVersion:      "10.15.6",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 OPR/70.0.3728.119",
		browser:        BrowserOpera,
		browserVersion: "70.0",
		os:             OSLinux,
	},
}

var userAgentsFirefox = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        BrowserFirefox,
		browserVersion: "79.0",
		os:             OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        BrowserFirefox,
		browserVersion: "79.0",
		os:             OSMac,
		osVersion:      "10.15",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux i686; rv:79.0) Gecko/20100101 Firefox/79.0",
		browser:        BrowserFirefox,
		browserVersion: "79.0",
		os:             OSLinux,
	},
}

var userAgentsChrome = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        BrowserChrome,
		browserVersion: "84.0",
		os:             OSWindows,
		osVersion:      "10",
	},
	{
		ua:             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        BrowserChrome,
		browserVersion: "84.0",
		os:             OSLinux,
	},
}

var userAgentsSafari = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36",
		browser:        BrowserSafari,
		browserVersion: "6.0",
		os:             OSMac,
		osVersion:      "10.15.6",
	},
}

var userAgentsIE = []testUserAgent{
	{
		ua:             "Mozilla/5.0 (Windows NT 10.0; Trident/7.0; rv:11.0) like Gecko",
		browser:        BrowserIE,
		browserVersion: "11.0",
		os:             OSWindows,
		osVersion:      "10",
	},
}

var userAgentsAll = mergeUserAgentLists(userAgentsEdge,
	userAgentsOpera,
	userAgentsFirefox,
	userAgentsChrome,
	userAgentsSafari,
	userAgentsIE)

func mergeUserAgentLists(ua ...[]testUserAgent) []testUserAgent {
	list := make([]testUserAgent, 0)

	for _, l := range ua {
		list = append(list, l...)
	}

	return list
}
