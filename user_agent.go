package pirsch

import (
	"strings"
	"unicode"
)

const (
	// Browser
	BrowserChrome  = "Chrome"
	BrowserFirefox = "Firefox"
	BrowserSafari  = "Safari"
	BrowserOpera   = "Opera"
	BrowserEdge    = "Edge"
	BrowserIE      = "IE"

	// Operating System
	OSWindows       = "Windows"
	OSMac           = "Mac"
	OSLinux         = "Linux"
	OSAndroid       = "Android"
	OSiOS           = "iOS"
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
		} else if strings.HasPrefix(sys, "Intel Mac OS X") {
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

	for _, product := range products {
		if strings.HasPrefix(product, "Chrome/") {
			// Chrome can be pretty much anything, so check for other browsers before deciding that this is Chrome
			if prefix := findPrefix(products, "Edg/", "Edge/", "EdgA/"); prefix != "" {
				browser = BrowserEdge
				version = getProductVersion(prefix, 1)
				break
			}

			if prefix := findPrefix(products, "Opera/", "OPR/"); prefix != "" {
				browser = BrowserOpera
				version = getProductVersion(prefix, 1)
				break
			}

			browser = BrowserChrome
			version = getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Chromium/") {
			browser = BrowserChrome
			version = getProductVersion(product, 1)
			break
		} else if strings.HasPrefix(product, "Firefox/") {
			browser = BrowserFirefox
			version = getProductVersion(product, 1)
			break
		} else if strings.HasPrefix(product, "Opera/") || strings.HasPrefix(product, "OPR/") {
			browser = BrowserOpera
			version = getProductVersion(product, 1)
			break
		} else if (os == OSMac || os == OSiOS) && (strings.HasPrefix(product, "Safari/") || product == "Mobile/15E148") {
			if prefix := findPrefix(products, "FxiOS/"); prefix != "" {
				browser = BrowserFirefox
				version = getProductVersion(prefix, 2)
				break
			}

			if prefix := findPrefix(products, "EdgiOS/"); prefix != "" {
				browser = BrowserEdge
				version = getProductVersion(prefix, 1)
				break
			}

			// TODO this might falsely identify Chrome?
			browser = BrowserSafari
			version = safariVersions[getProductVersion(product, 1)]
			break
		}
	}

	return browser, version
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
