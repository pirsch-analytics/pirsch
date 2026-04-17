package ua

import (
	"net/http"
	"strings"
	"unicode"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/util"
)

const (
	uaSystemLeftDelimiter     = '('
	uaSystemRightDelimiter    = ')'
	uaSystemDelimiter         = ";"
	uaProductVersionDelimiter = '/'
	uaVersionDelimiter        = '.'
)

// UserAgent parses the User-Agent header to extract useful information.
type UserAgent struct{}

// NewUserAgent creates a new UserAgent.
func NewUserAgent() *UserAgent {
	return &UserAgent{}
}

// Step implements ingest.PipeStep to process a step.
// It sets the User-Agent parameters for the request.
func (ua *UserAgent) Step(request *ingest.Request) (bool, error) {
	i := ua.parse(request.Request)
	request.Browser = util.Shorten(i.browser, 20)
	request.BrowserVersion = util.Shorten(i.browserVersion, 20)
	request.OS = util.Shorten(i.os, 20)
	request.OSVersion = util.Shorten(i.osVersion, 20)
	request.Desktop = i.isDesktop()
	request.Mobile = i.isMobile()
	return false, nil
}

func (ua *UserAgent) parse(r *http.Request) info {
	system, products, systemFromCH, productFromCH := ua.parseSystemAndProduct(r)
	userAgent := info{}

	if systemFromCH {
		userAgent.os, userAgent.osVersion = ua.mapOS(system)
	} else {
		userAgent.os, userAgent.osVersion = ua.getOS(system)
	}

	if productFromCH {
		userAgent.browser, userAgent.browserVersion = products[0], products[1]
	} else {
		userAgent.browser, userAgent.browserVersion = ua.getBrowser(products, system, userAgent.os)
	}

	userAgent.mobile = ua.getMobile(r)
	return userAgent
}

func (ua *UserAgent) getOS(system []string) (string, string) {
	os := ""
	version := ""

	for _, sys := range system {
		if strings.HasPrefix(sys, "Windows Phone") || strings.HasPrefix(sys, "Windows Mobile") {
			os = pkg.OSWindowsMobile
			version = ua.getWindowsMobileVersion(sys)
			break
		} else if strings.HasPrefix(sys, "Windows") {
			os = pkg.OSWindows
			version = ua.getWindowsVersion(sys)
			// leave a chance to detect IE...
		} else if strings.HasPrefix(sys, "Intel Mac OS X") || strings.HasPrefix(sys, "PPC Mac OS X") || strings.HasPrefix(sys, "Mac OS X Version") {
			os = pkg.OSMac
			version = ua.getMacVersion(sys)
			break
		} else if strings.HasPrefix(sys, "Linux") {
			os = pkg.OSLinux
			// this might be Android...
		} else if strings.HasPrefix(sys, "Android") {
			if prefix := ua.findPrefix(system, "Windows Mobile"); prefix != "" {
				os = pkg.OSWindowsMobile
				version = ua.getProductVersion(prefix, 1)
				break
			}

			os = pkg.OSAndroid
			version = ua.getAndroidVersion(sys)
			break
		} else if strings.HasPrefix(sys, "CPU iPhone OS") ||
			strings.HasPrefix(sys, "CPU OS") ||
			strings.HasPrefix(sys, "iPad") ||
			strings.HasPrefix(sys, "iPhone") {
			os = pkg.OSiOS
			version = ua.getIOSVersion(system)
			break
		} else if strings.HasPrefix(sys, "CrOS") {
			os = pkg.OSChrome
			version = ua.getChromeOSVersion(sys)
			break
		}
	}

	return os, version
}

func (ua *UserAgent) mapOS(system []string) (string, string) {
	if len(system) != 2 {
		return "", ""
	}

	os, found := osMapping[system[0]]

	if !found {
		return "", ""
	}

	if os == pkg.OSWindows {
		return os, windowsVersions[ua.getOSVersion(system[1], 1)]
	}

	return os, ua.getOSVersion(system[1], 1)
}

func (ua *UserAgent) getBrowser(products []string, system []string, os string) (string, string) {
	browser := ""
	version := ""

	// special case for IE
	if v := ua.isIE(system); v != "" {
		return pkg.BrowserIE, v
	}

	productChrome := ""
	productSafari := ""

	for _, product := range products {
		if strings.HasPrefix(product, "Chrome/") {
			productChrome = product
		} else if strings.HasPrefix(product, "Safari/") {
			productSafari = product
		} else if strings.HasPrefix(product, "DuckDuckGo/") {
			return pkg.BrowserDuckDuckGo, ua.getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Arc/") || strings.HasPrefix(product, "ArcMobile2/") {
			return pkg.BrowserArc, ua.getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "CriOS/") {
			return pkg.BrowserChrome, ua.getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Edg/") || strings.HasPrefix(product, "Edge/") || strings.HasPrefix(product, "EdgA/") || strings.HasPrefix(product, "EdgiOS/") {
			return pkg.BrowserEdge, ua.getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Chromium/") {
			return pkg.BrowserChrome, ua.getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Firefox/") || strings.HasPrefix(product, "FxiOS/") {
			return pkg.BrowserFirefox, ua.getProductVersion(product, 1)
		} else if strings.HasPrefix(product, "Opera/") || strings.HasPrefix(product, "OPR/") {
			return pkg.BrowserOpera, ua.getProductVersion(product, 1)
		}
	}

	// If we get to that point, it's going to get ugly and inaccurate, because Safari and Chrome send almost identical user agents most of the time.
	// But anything coming from Mac or iOS is most likely Safari, I guess...
	if (os == pkg.OSMac || os == pkg.OSiOS) && productSafari != "" && productChrome == "" {
		browser = pkg.BrowserSafari
		version = ua.getSafariVersion(products, productSafari)
	} else if productChrome != "" {
		browser = pkg.BrowserChrome
		version = ua.getProductVersion(productChrome, 1)
	}

	return browser, version
}

func (ua *UserAgent) getMobile(r *http.Request) *bool {
	mobile := r.Header.Get("Sec-CH-UA-Mobile")

	if mobile != "" && (mobile == "?0" || mobile == "?1") {
		b := mobile == "?1"
		return &b
	}

	return nil
}

// older Safari versions send their version number inside the Version/ product string instead of the Safari/ part
func (ua *UserAgent) getSafariVersion(products []string, productSafari string) string {
	productVersion := ua.findPrefix(products, "Version/")

	if productVersion != "" {
		return ua.getProductVersion(productVersion, 1)
	}

	return safariVersions[ua.getProductVersion(productSafari, 1)]
}

// isIE checks if the user agent is IE from the system information part and returns the version number,
// or else an empty string is returned.
// The version number is part of "MSIE <version>" or "rv:<version>".
func (ua *UserAgent) isIE(system []string) string {
	for _, sys := range system {
		if strings.HasPrefix(sys, "Trident/") {
			if msie := ua.findPrefix(system, "MSIE "); msie != "" {
				return msie[5:]
			}

			if rv := ua.findPrefix(system, "rv:"); rv != "" {
				return rv[3:]
			}

			return ""
		} else if strings.HasPrefix(sys, "MSIE ") {
			return sys[5:]
		}
	}

	return ""
}

func (ua *UserAgent) getWindowsMobileVersion(system string) string {
	parts := strings.Split(system, " ")

	if len(parts) > 2 {
		return ua.getOSVersion(parts[2], 1)
	}

	return ""
}

func (ua *UserAgent) getWindowsVersion(system string) string {
	if i := strings.LastIndexByte(system, ' '); i > -1 {
		return windowsVersions[ua.getOSVersion(system[i+1:], 1)]
	}

	return "11" // assume Windows 11
}

func (ua *UserAgent) getMacVersion(system string) string {
	if len(system) > 14 {
		return ua.getOSVersion(strings.ReplaceAll(system[15:], "_", "."), 1)
	}

	return ""
}

func (ua *UserAgent) getAndroidVersion(system string) string {
	if len(system) > 7 {
		return ua.getOSVersion(system[8:], 1)
	}

	return ""
}

func (ua *UserAgent) getIOSVersion(system []string) string {
	for _, sys := range system {
		// CPU iPhone OS <version> like ...
		// CPU OS <version> like ...
		// iOS <version>
		parts := strings.Split(sys, " ")

		if len(parts) > 3 {
			if parts[2] == "OS" {
				return ua.getOSVersion(strings.Replace(parts[3], "_", ".", -1), 1)
			}

			return ua.getOSVersion(strings.Replace(parts[2], "_", ".", -1), 1)
		} else if len(parts) == 2 && parts[0] == "iOS" {
			return ua.getOSVersion(strings.Replace(parts[1], "_", ".", -1), 1)
		}
	}

	return ""
}

func (ua *UserAgent) getChromeOSVersion(system string) string {
	if i := strings.LastIndexByte(system, ' '); i > -1 {
		return ua.getOSVersion(system[i+1:], 1)
	}

	return ""
}

func (ua *UserAgent) findPrefix(list []string, prefix ...string) string {
	for _, entry := range list {
		for _, pre := range prefix {
			if strings.HasPrefix(entry, pre) {
				return entry
			}
		}
	}

	return ""
}

func (ua *UserAgent) getProductVersion(version string, n int) string {
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

func (ua *UserAgent) getOSVersion(version string, n int) string {
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

func (ua *UserAgent) parseSystemAndProduct(r *http.Request) ([]string, []string, bool, bool) {
	userAgent := strings.Trim(r.UserAgent(), ` '"`)

	if userAgent == "" {
		return nil, nil, false, false
	}

	systemStart := strings.IndexRune(userAgent, uaSystemLeftDelimiter)
	systemEnd := strings.IndexRune(userAgent, uaSystemRightDelimiter)
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
		system = ua.parseSystem(userAgent, systemStart, systemEnd)
	}

	chProduct := r.Header.Get("Sec-CH-UA")
	productFromCH := false
	var products []string

	if chProduct != "" {
		products = ua.parseProductsFromCH(chProduct)

		if len(products) == 0 {
			products = ua.parseProducts(userAgent, systemStart, systemEnd)
		} else {
			productFromCH = true
		}
	} else {
		products = ua.parseProducts(userAgent, systemStart, systemEnd)
	}

	return system, products, platformFromCH, productFromCH
}

func (ua *UserAgent) parseSystem(userAgent string, systemStart, systemEnd int) []string {
	var system []string

	if systemStart > -1 && systemEnd > -1 && systemStart < systemEnd {
		systemParts := strings.Split(userAgent[systemStart+1:systemEnd], uaSystemDelimiter)
		system = make([]string, 0, len(systemParts))

		for i := range systemParts {
			systemParts[i] = strings.TrimSpace(systemParts[i])

			if !ua.ignoreSystemString(systemParts[i]) {
				system = append(system, systemParts[i])
			}
		}
	}

	return system
}

func (ua *UserAgent) parseProducts(userAgent string, systemStart, systemEnd int) []string {
	var productStrings []string

	if systemStart > -1 && systemEnd > -1 {
		productStrings = strings.Fields(userAgent[:systemStart] + " " + userAgent[systemEnd+1:])
	} else {
		productStrings = strings.Fields(userAgent)
	}

	products := make([]string, 0, len(productStrings))

	for _, str := range productStrings {
		if !ua.ignoreProductString(str) {
			products = append(products, str)
		}
	}

	return products
}

func (ua *UserAgent) ignoreSystemString(system string) bool {
	return system == "" || strings.ContainsAny(system, "\"'`=")
}

func (ua *UserAgent) ignoreProductString(product string) bool {
	for _, prefix := range filterProductPrefix {
		if strings.HasPrefix(product, prefix) {
			return true
		}
	}

	return strings.ContainsAny(product, "\"'`=")
}

func (ua *UserAgent) parseProductsFromCH(header string) []string {
	productStrings := strings.Split(header, ",")

	if ua.isHeadlessCH(productStrings) {
		return []string{"Headless", ""}
	}

	genericProduct, genericVersion := "", ""

	for _, str := range productStrings {
		product, version, found := strings.Cut(str, ";")

		if found {
			product = strings.Trim(product, `"'`)
			version = strings.Trim(version, `"'`)

			if strings.Contains(product, "Google Chrome") {
				return []string{pkg.BrowserChrome, ua.parseProductVersion(version)}
			} else if strings.Contains(product, "Microsoft Edge") {
				return []string{pkg.BrowserEdge, ua.parseProductVersion(version)}
			} else if strings.Contains(product, "Opera") {
				return []string{pkg.BrowserOpera, ua.parseProductVersion(version)}
			} else if strings.Contains(product, "Arc") {
				return []string{pkg.BrowserArc, ua.parseProductVersion(version)}
			} else if !strings.Contains(product, "Not") && !strings.Contains(product, "Brand") && !strings.Contains(product, "Chromium") {
				genericProduct = strings.Trim(product, `"' `)
				genericVersion = ua.parseProductVersion(version)
			}
		}
	}

	if genericProduct != "" {
		return []string{genericProduct, genericVersion}
	}

	return nil
}

func (ua *UserAgent) isHeadlessCH(productStrings []string) bool {
	for _, str := range productStrings {
		if strings.Contains(strings.ToLower(str), "headless") {
			return true
		}
	}

	return false
}

func (ua *UserAgent) parseProductVersion(version string) string {
	version = strings.ToLower(version)

	if strings.HasPrefix(version, `v="`) {
		return strings.Trim(version[3:], `"`)
	}

	return ""
}
