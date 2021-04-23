package ua

import (
	"strings"
	"unicode"
)

const (
	// BrowserChrome represents the Chrome and Chromium browser.
	BrowserChrome = "Chrome"

	// BrowserFirefox represents the Firefox browser.
	BrowserFirefox = "Firefox"

	// BrowserSafari  represents the Safari browser.
	BrowserSafari = "Safari"

	// BrowserOpera represents the Opera browser.
	BrowserOpera = "Opera"

	// BrowserEdge represents the Edge browser.
	BrowserEdge = "Edge"

	// BrowserIE represents the Internet Explorer browser.
	BrowserIE = "IE"

	// OSWindows represents the Windows operating system.
	OSWindows = "Windows"

	// OSMac represents the Mac operating system.
	OSMac = "Mac"

	// OSLinux represents a Linux distribution.
	OSLinux = "Linux"

	// OSAndroid represents the Android operating system.
	OSAndroid = "Android"

	// OSiOS represents the iOS operating system.
	OSiOS = "iOS"

	// OSWindowsMobile represents the Windows Mobile operating system.
	OSWindowsMobile = "Windows Mobile"

	// used to parse the User-Agent header
	uaSystemLeftDelimiter     = '('
	uaSystemRightDelimiter    = ')'
	uaSystemDelimiter         = ";"
	uaProductVersionDelimiter = '/'
	uaVersionDelimiter        = '.'
)

// UserAgent contains information extracted from the User-Agent header.
type UserAgent struct {
	// Browser is the browser name.
	Browser string

	// BrowserVersion is the browser (non technical) version number.
	BrowserVersion string

	// OS is the operating system.
	OS string

	// OSVersion is the operating system version number.
	OSVersion string
}

// IsDesktop returns true if the user agent is a desktop device.
func (ua *UserAgent) IsDesktop() bool {
	return ua.OS == OSWindows || ua.OS == OSMac || ua.OS == OSLinux
}

// IsMobile returns true if the user agent is a mobile device.
func (ua *UserAgent) IsMobile() bool {
	return ua.OS == OSAndroid || ua.OS == OSiOS || ua.OS == OSWindowsMobile
}

// ParseUserAgent parses given User-Agent header and returns the extracted information.
// This just supports major browsers and operating systems, we don't care about browsers and OSes that have no market share,
// unless you prove us wrong.
func ParseUserAgent(ua string) UserAgent {
	system, products := parseUserAgent(ua)
	userAgent := UserAgent{}
	userAgent.OS, userAgent.OSVersion = getOS(system)
	userAgent.Browser, userAgent.BrowserVersion = getBrowser(products, system, userAgent.OS)
	return userAgent
}

func getOS(system []string) (string, string) {
	os := ""
	version := ""

	for _, sys := range system {
		if strings.HasPrefix(sys, "Windows Phone") || strings.HasPrefix(sys, "Windows Mobile") {
			os = OSWindowsMobile
			version = getWindowsMobileVersion(sys)
			break
		} else if strings.HasPrefix(sys, "Windows") {
			os = OSWindows
			version = getWindowsVersion(sys)
			// leave a chance to detect IE...
		} else if strings.HasPrefix(sys, "Intel Mac OS X") || strings.HasPrefix(sys, "PPC Mac OS X") {
			os = OSMac
			version = getMacVersion(sys)
			break
		} else if strings.HasPrefix(sys, "Linux") {
			os = OSLinux
			// this might be Android...
		} else if strings.HasPrefix(sys, "Android") {
			if prefix := findPrefix(system, "Windows Mobile"); prefix != "" {
				os = OSWindowsMobile
				version = getProductVersion(prefix, 1)
				break
			}

			os = OSAndroid
			version = getAndroidVersion(sys)
			break
		} else if strings.HasPrefix(sys, "CPU iPhone OS") || strings.HasPrefix(sys, "CPU OS") {
			os = OSiOS
			version = getiOSVersion(sys)
			break
		}
	}

	return os, version
}

func getBrowser(products []string, system []string, os string) (string, string) {
	browser := ""
	version := ""

	// special case for IE
	if v := isIE(system); v != "" {
		return BrowserIE, v
	}

	productChrome := ""
	productSafari := ""

	for _, product := range products {
		if strings.HasPrefix(product, "Chrome/") {
			productChrome = product
		} else if strings.HasPrefix(product, "Safari/") {
			productSafari = product
		} else if strings.HasPrefix(product, "CriOS/") {
			return BrowserChrome, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Edg/") || strings.HasPrefix(product, "Edge/") || strings.HasPrefix(product, "EdgA/") || strings.HasPrefix(product, "EdgiOS/") {
			return BrowserEdge, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Chromium/") {
			return BrowserChrome, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Firefox/") || strings.HasPrefix(product, "FxiOS/") {
			return BrowserFirefox, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Opera/") || strings.HasPrefix(product, "OPR/") {
			return BrowserOpera, getProductVersion(product, 1)
		}
	}

	// When we made it to this point, it's gone get ugly and inaccurate, as Safari and Chrome send almost identical
	// user agents most of the time. But anything coming from Mac or iOS is most likely Safari, I guess...
	if (os == OSMac || os == OSiOS) && productSafari != "" && productChrome == "" {
		browser = BrowserSafari
		version = getSafariVersion(products, productSafari)
	} else if productChrome != "" {
		browser = BrowserChrome
		version = getProductVersion(productChrome, 1)
	}

	return browser, version
}

// older Safari versions send their version number inside the Version/ product string instead of the Safari/ part
func getSafariVersion(products []string, productSafari string) string {
	productVersion := findPrefix(products, "Version/")

	if productVersion != "" {
		return getProductVersion(productVersion, 1)
	}

	return safariVersions[getProductVersion(productSafari, 1)]
}

// isIE checks if the user agent is IE from the system information part and returns the version number,
// or else an empty string is returned.
// The version number is part of "MSIE <version>" or "rv:<version>".
func isIE(system []string) string {
	for _, sys := range system {
		if strings.HasPrefix(sys, "Trident/") {
			if msie := findPrefix(system, "MSIE "); msie != "" {
				return msie[5:]
			}

			if rv := findPrefix(system, "rv:"); rv != "" {
				return rv[3:]
			}

			return ""
		} else if strings.HasPrefix(sys, "MSIE ") {
			return sys[5:]
		}
	}

	return ""
}

func getWindowsMobileVersion(system string) string {
	parts := strings.Split(system, " ")

	if len(parts) > 2 {
		return getOSVersion(parts[2], 2)
	}

	return ""
}

func getWindowsVersion(system string) string {
	if i := strings.LastIndexByte(system, ' '); i > -1 {
		return windowsVersions[getOSVersion(system[i+1:], 1)]
	}

	return ""
}

func getMacVersion(system string) string {
	if len(system) > 14 {
		return getOSVersion(strings.ReplaceAll(system[15:], "_", "."), 2)
	}

	return ""
}

func getAndroidVersion(system string) string {
	if len(system) > 7 {
		return getOSVersion(system[8:], 2)
	}

	return ""
}

func getiOSVersion(system string) string {
	parts := strings.Split(system, " ")

	// CPU iPhone OS <version> like ...
	// CPU OS <version> like ...
	if len(parts) > 3 {
		if parts[2] == "OS" {
			return getOSVersion(strings.Replace(parts[3], "_", ".", -1), 2)
		}

		return getOSVersion(strings.Replace(parts[2], "_", ".", -1), 2)
	}

	return ""
}

// returns the first prefix it finds in the prefix list, or else an empty string is returned
func findPrefix(list []string, prefix ...string) string {
	for _, entry := range list {
		for _, pre := range prefix {
			if strings.HasPrefix(entry, pre) {
				return entry
			}
		}
	}

	return ""
}

func getProductVersion(version string, n int) string {
	out := make([]rune, 0, len(version))
	read := false
	dots := 0

	for _, r := range []rune(version) {
		if r == uaProductVersionDelimiter {
			read = true
		} else if read && unicode.IsNumber(r) || r == uaVersionDelimiter {
			if r == uaVersionDelimiter {
				dots++

				if dots > n {
					break
				}
			}

			out = append(out, r)
		}
	}

	return string(out)
}

func getOSVersion(version string, n int) string {
	out := make([]rune, 0, len(version))
	dots := 0

	for _, r := range []rune(version) {
		if unicode.IsNumber(r) || r == uaVersionDelimiter {
			if r == uaVersionDelimiter {
				dots++

				if dots > n {
					break
				}
			}

			out = append(out, r)
		}
	}

	return string(out)
}

// parses, filters and returns the system and product strings
func parseUserAgent(ua string) ([]string, []string) {
	// remove leading spaces, single and double quotes
	ua = strings.Trim(ua, ` '"`)

	if ua == "" {
		return nil, nil
	}

	// extract client system data
	var system []string
	systemStart := strings.IndexRune(ua, uaSystemLeftDelimiter)
	systemEnd := strings.IndexRune(ua, uaSystemRightDelimiter)

	if systemStart > -1 && systemEnd > -1 && systemStart < systemEnd {
		systemParts := strings.Split(ua[systemStart+1:systemEnd], uaSystemDelimiter)
		system = make([]string, 0, len(systemParts))

		for i := range systemParts {
			systemParts[i] = strings.TrimSpace(systemParts[i])

			if systemParts[i] != "" {
				system = append(system, systemParts[i])
			}
		}
	}

	// parse products and filter for meaningless strings
	var productStrings []string

	if systemStart > -1 && systemEnd > -1 {
		productStrings = strings.Fields(ua[:systemStart] + " " + ua[systemEnd+1:])
	} else {
		productStrings = strings.Fields(ua)
	}

	products := make([]string, 0, len(productStrings))

	for _, str := range productStrings {
		if !ignoreProductString(str) {
			products = append(products, str)
		}
	}

	return system, products
}

func ignoreProductString(product string) bool {
	for _, prefix := range filterProductPrefix {
		if strings.HasPrefix(product, prefix) {
			return true
		}
	}

	return false
}
