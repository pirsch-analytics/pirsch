package ua

var (
	// filterProductPrefix is a list of product prefixes to ignore, as they provide no value in identifying a browser.
	filterProductPrefix = []string{
		"Mozilla/",
		"AppleWebKit/",
		"(KHTML,", // we split by space, so this was "(KHTML like Gecko)"
		"like",
		"Gecko", // just Gecko, to ignore it in IE user agent, as a product version string and in "(KHTML like Gecko)"
		"Mobile/",
		"QtWebEngine/",
	}

	// windowsVersions maps a Windows user agent versions to the product versions.
	// https://en.wikipedia.org/wiki/List_of_Microsoft_Windows_versions
	windowsVersions = map[string]string{
		"5.0":  "2000",
		"5.1":  "XP",
		"5.2":  "XP",
		"6.0":  "Vista",
		"6.1":  "7",
		"6.2":  "8",
		"6.3":  "8",
		"10.0": "10",
		"13.0": "11", // this is unreliable, as Chrome and Firefox decided to freeze the User-Agent header
		"CE":   "CE",
	}

	// safariVersions maps a Safari user agent versions to the product versions.
	// https://en.wikipedia.org/wiki/Safari_version_history
	safariVersions = map[string]string{
		"533.16": "5.0",
		"533.17": "5.0",
		"533.18": "5.0",
		"533.19": "5.0",
		"533.20": "5.0",
		"533.21": "5.0",
		"533.22": "5.0",
		"534.48": "5.1",
		"534.51": "5.1",
		"534.52": "5.1",
		"534.53": "5.1",
		"534.54": "5.1",
		"534.55": "5.1",
		"534.56": "5.1",
		"534.57": "5.1",
		"534.58": "5.1",
		"534.59": "5.1",
		"536.25": "6.0",
		"536.26": "6.0",
		"536.28": "6.0",
		"536.29": "6.0",
		"536.30": "6.0",
		"537.36": "6.0", // unlisted
		"537.43": "6.1",
		"537.85": "6.2",
		"537.71": "7.0",
		"537.73": "7.0",
		"537.75": "7.0",
		"537.76": "7.0",
		"537.77": "7.0",
		"537.78": "7.0",
		"600.3":  "7.1",
		"600.8":  "7.1",
		"538.35": "8.0",
		"600.6":  "8.0",
		"600.7":  "8.0",
		"601.1":  "9.0",
		"601.2":  "9.0",
		"601.3":  "9.0",
		"601.4":  "9.0",
		"601.5":  "9.1",
		"601.6":  "9.1",
		"601.7":  "9.1",
		"602.1":  "10.0",
		"602.2":  "10.0",
		"602.3":  "10.0",
		"602.4":  "10.0",
		"603.1":  "10.1",
		"603.2":  "10.1",
		"603.3":  "10.1",
		"604.2":  "11.0",
		"605.1":  "11.1",
		"606.1":  "12.0",
		"607.1":  "12.1",
		"608.2":  "13.0",
		"610.2":  "14.0",
		"610.3":  "14.0",
		"610.4":  "14.0",
		"611.1":  "14.1",
		"611.2":  "14.1",
		"611.3":  "14.1",
		"612.1":  "15.0",
		"612.2":  "15.1",
		"612.3":  "15.2",
		"612.4":  "15.3",
		"613.1":  "15.4",
		"613.2":  "15.5",
	}
)
