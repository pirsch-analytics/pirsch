package ua

import (
	"github.com/emvi/null"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"net/http"
	"strings"
	"time"
	"unicode"
)

const (
	// used to parse the User-Agent header
	uaSystemLeftDelimiter     = '('
	uaSystemRightDelimiter    = ')'
	uaSystemDelimiter         = ";"
	uaProductVersionDelimiter = '/'
	uaVersionDelimiter        = '.'
)

// Parse parses the User-Agent header for given request and returns the extracted information.
// This supports major browsers and operating systems.
func Parse(r *http.Request) model.UserAgent {
	system, products, systemFromCH, productFromCH := parse(r)
	userAgent := model.UserAgent{
		Time:      time.Now().UTC(),
		UserAgent: r.UserAgent(),
	}

	if systemFromCH {
		userAgent.OS, userAgent.OSVersion = mapOS(system)
	} else {
		userAgent.OS, userAgent.OSVersion = getOS(system)
	}

	if productFromCH {
		userAgent.Browser, userAgent.BrowserVersion = products[0], products[1]
	} else {
		userAgent.Browser, userAgent.BrowserVersion = getBrowser(products, system, userAgent.OS)
	}

	userAgent.Mobile = getMobile(r)
	return userAgent
}

func getOS(system []string) (string, string) {
	os := ""
	version := ""

	for _, sys := range system {
		if strings.HasPrefix(sys, "Windows Phone") || strings.HasPrefix(sys, "Windows Mobile") {
			os = pkg.OSWindowsMobile
			version = getWindowsMobileVersion(sys)
			break
		} else if strings.HasPrefix(sys, "Windows") {
			os = pkg.OSWindows
			version = getWindowsVersion(sys)
			// leave a chance to detect IE...
		} else if strings.HasPrefix(sys, "Intel Mac OS X") || strings.HasPrefix(sys, "PPC Mac OS X") {
			os = pkg.OSMac
			version = getMacVersion(sys)
			break
		} else if strings.HasPrefix(sys, "Linux") {
			os = pkg.OSLinux
			// this might be Android...
		} else if strings.HasPrefix(sys, "Android") {
			if prefix := findPrefix(system, "Windows Mobile"); prefix != "" {
				os = pkg.OSWindowsMobile
				version = getProductVersion(prefix, 1)
				break
			}

			os = pkg.OSAndroid
			version = getAndroidVersion(sys)
			break
		} else if strings.HasPrefix(sys, "CPU iPhone OS") || strings.HasPrefix(sys, "CPU OS") {
			os = pkg.OSiOS
			version = getiOSVersion(sys)
			break
		} else if strings.HasPrefix(sys, "CrOS") {
			os = pkg.OSChrome
			version = getChromeOSVersion(sys)
			break
		}
	}

	return os, version
}

func mapOS(system []string) (string, string) {
	if len(system) != 2 {
		return "", ""
	}

	os, found := osMapping[system[0]]

	if !found {
		return "", ""
	}

	if os == pkg.OSWindows {
		return os, windowsVersions[getOSVersion(system[1], 1)]
	}

	return os, getOSVersion(system[1], 1)
}

func getBrowser(products []string, system []string, os string) (string, string) {
	browser := ""
	version := ""

	// special case for IE
	if v := isIE(system); v != "" {
		return pkg.BrowserIE, v
	}

	productChrome := ""
	productSafari := ""

	for _, product := range products {
		if strings.HasPrefix(product, "Chrome/") {
			productChrome = product
		} else if strings.HasPrefix(product, "Safari/") {
			productSafari = product
		} else if strings.HasPrefix(product, "CriOS/") {
			return pkg.BrowserChrome, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Edg/") || strings.HasPrefix(product, "Edge/") || strings.HasPrefix(product, "EdgA/") || strings.HasPrefix(product, "EdgiOS/") {
			return pkg.BrowserEdge, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Chromium/") {
			return pkg.BrowserChrome, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Firefox/") || strings.HasPrefix(product, "FxiOS/") {
			return pkg.BrowserFirefox, getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Opera/") || strings.HasPrefix(product, "OPR/") {
			return pkg.BrowserOpera, getProductVersion(product, 1)
		}
	}

	// When we made it to this point, it's gone get ugly and inaccurate, as Safari and Chrome send almost identical
	// user agents most of the time. But anything coming from Mac or iOS is most likely Safari, I guess...
	if (os == pkg.OSMac || os == pkg.OSiOS) && productSafari != "" && productChrome == "" {
		browser = pkg.BrowserSafari
		version = getSafariVersion(products, productSafari)
	} else if productChrome != "" {
		browser = pkg.BrowserChrome
		version = getProductVersion(productChrome, 1)
	}

	return browser, version
}

func getMobile(r *http.Request) null.Bool {
	mobile := r.Header.Get("Sec-CH-UA-Mobile")

	if mobile != "" && (mobile == "?0" || mobile == "?1") {
		return null.NewBool(mobile == "?1", true)
	}

	return null.NewBool(false, false)
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
		return getOSVersion(parts[2], 1)
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
		return getOSVersion(strings.ReplaceAll(system[15:], "_", "."), 1)
	}

	return ""
}

func getAndroidVersion(system string) string {
	if len(system) > 7 {
		return getOSVersion(system[8:], 1)
	}

	return ""
}

func getiOSVersion(system string) string {
	parts := strings.Split(system, " ")

	// CPU iPhone OS <version> like ...
	// CPU OS <version> like ...
	if len(parts) > 3 {
		if parts[2] == "OS" {
			return getOSVersion(strings.Replace(parts[3], "_", ".", -1), 1)
		}

		return getOSVersion(strings.Replace(parts[2], "_", ".", -1), 1)
	}

	return ""
}

func getChromeOSVersion(system string) string {
	if i := strings.LastIndexByte(system, ' '); i > -1 {
		return getOSVersion(system[i+1:], 1)
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
func parse(r *http.Request) ([]string, []string, bool, bool) {
	ua := strings.Trim(r.UserAgent(), ` '"`)

	if ua == "" {
		return nil, nil, false, false
	}

	systemStart := strings.IndexRune(ua, uaSystemLeftDelimiter)
	systemEnd := strings.IndexRune(ua, uaSystemRightDelimiter)
	chPlatform := strings.Trim(r.Header.Get("Sec-CH-UA-Platform"), `"'`)
	platformFromCH := false
	var system []string

	if chPlatform != "" && strings.ToLower(chPlatform) != "unknown" {
		system = []string{
			chPlatform,
			strings.Trim(r.Header.Get("Sec-CH-UA-Platform-Version"), `"'`),
		}
		platformFromCH = true
	} else {
		system = parseSystem(ua, systemStart, systemEnd)
	}

	chProduct := r.Header.Get("Sec-CH-UA")
	productFromCH := false
	var products []string

	if chProduct != "" {
		products = parseProductsFromCH(chProduct)

		if len(products) == 0 {
			products = parseProducts(ua, systemStart, systemEnd)
		} else {
			productFromCH = true
		}
	} else {
		products = parseProducts(ua, systemStart, systemEnd)
	}

	return system, products, platformFromCH, productFromCH
}

func parseSystem(ua string, systemStart, systemEnd int) []string {
	var system []string

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

	return system
}

func parseProducts(ua string, systemStart, systemEnd int) []string {
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

	return products
}

func ignoreProductString(product string) bool {
	for _, prefix := range filterProductPrefix {
		if strings.HasPrefix(product, prefix) {
			return true
		}
	}

	return false
}

func parseProductsFromCH(header string) []string {
	productStrings := strings.Split(header, ",")
	genericProduct, genericVersion := "", ""

	for _, str := range productStrings {
		product, version, found := strings.Cut(str, ";")

		if found {
			product = strings.Trim(product, `"'`)
			version = strings.Trim(version, `"'`)

			if strings.Contains(product, "Google Chrome") {
				return []string{pkg.BrowserChrome, parseProductVersion(version)}
			} else if strings.Contains(product, "Microsoft Edge") {
				return []string{pkg.BrowserEdge, parseProductVersion(version)}
			} else if strings.Contains(product, "Opera") {
				return []string{pkg.BrowserOpera, parseProductVersion(version)}
			} else if !strings.Contains(product, "Brand") && !strings.Contains(product, "Chromium") {
				genericProduct = strings.Trim(product, `"' `)
				genericVersion = parseProductVersion(version)
			}
		}
	}

	if genericProduct != "" {
		return []string{genericProduct, genericVersion}
	}

	return nil
}

func parseProductVersion(version string) string {
	version = strings.ToLower(version)

	if strings.HasPrefix(version, `v="`) {
		return strings.Trim(version[3:], `"`)
	}

	return ""
}
