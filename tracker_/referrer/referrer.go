package referrer

import (
	"fmt"
	"golang.org/x/net/html"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

const (
	androidAppPrefix   = "android-app://"
	googlePlayStoreURL = "https://play.google.com/store/apps/details?id=%s"
)

// QueryParams is a list of query parameters to set the referrer.
var QueryParams = []string{
	"ref",
	"referer",
	"referrer",
	"source",
	"utm_source",
}

// Ignore returns whether a referrer should be ignored or not.
func Ignore(r *http.Request) bool {
	referrer := getFromHeaderOrQuery(r)

	if referrer == "" {
		return false
	}

	u, err := url.ParseRequestURI(referrer)

	if err == nil {
		referrer = u.Hostname()
	}

	referrer = stripSubdomain(referrer)
	_, found := blacklist[referrer]
	return found
}

// Get returns the referrer for given request.
func Get(r *http.Request, ref string, domainBlacklist []string, ignoreSubdomain bool) (string, string, string) {
	referrer := ""

	if ref != "" {
		referrer = ref
	} else {
		referrer = getFromHeaderOrQuery(r)
	}

	if referrer == "" {
		return "", "", ""
	}

	if strings.HasPrefix(strings.ToLower(referrer), androidAppPrefix) {
		name, icon := getAndroidAppName(referrer)
		return referrer, name, icon
	}

	u, err := url.ParseRequestURI(referrer)

	if err != nil {
		if isIP(referrer) {
			return "", "", ""
		}

		// accept non-url referrers (from utm_source for example)
		if !containsString(domainBlacklist, referrer) {
			return "", strings.TrimSpace(referrer), ""
		}

		return "", "", ""
	}

	if u.Path == "/" {
		u.Path = ""
	}

	hostname := strings.ToLower(u.Hostname())

	if isIP(hostname) {
		return "", "", ""
	}

	name := groups[hostname+u.Path]

	if name == "" {
		// try again without path
		name = groups[hostname]
	}

	if ignoreSubdomain {
		hostname = stripSubdomain(hostname)
	}

	if containsString(domainBlacklist, hostname) {
		return "", "", ""
	}

	if name == "" {
		name = hostname
	}

	// remove query parameters and anchor
	u.Host = strings.ToLower(u.Host)
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), name, ""
}

func getFromHeaderOrQuery(r *http.Request) string {
	referrer := r.Header.Get("Referer")

	if referrer == "" {
		for _, param := range QueryParams {
			referrer = r.URL.Query().Get(param)

			if referrer != "" {
				return referrer
			}
		}
	}

	return referrer
}

func isIP(referrer string) bool {
	referrer = strings.Trim(referrer, "/")

	if strings.Contains(referrer, ":") {
		var err error
		referrer, _, err = net.SplitHostPort(referrer)

		if err != nil {
			return false
		}
	}

	_, err := netip.ParseAddr(referrer)
	return err == nil
}

func stripSubdomain(hostname string) string {
	if hostname == "" {
		return ""
	}

	runes := []rune(hostname)
	index := len(runes) - 1
	dots := 0

	for i := index; i > 0; i-- {
		if runes[i] == '.' {
			dots++

			if dots == 2 {
				index++
				break
			}
		}

		index--
	}

	return hostname[index:]
}

func getAndroidAppName(referrer string) (string, string) {
	packageName := referrer[len(androidAppPrefix):]
	resp, err := http.Get(fmt.Sprintf(googlePlayStoreURL, packageName))

	if err != nil || resp.StatusCode != http.StatusOK {
		return "", ""
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	doc, err := html.Parse(resp.Body)

	if err != nil {
		return "", ""
	}

	titleNode := findAndroidAppName(doc)

	if titleNode == nil {
		return "", ""
	}

	appName := findTextNode(titleNode)

	if appName == nil {
		return "", ""
	}

	icon := ""
	iconNode := findAndroidAppIcon(doc)

	if iconNode != nil {
		icon = getHTMLAttribute(iconNode, "src")
	}

	return appName.Data, icon
}

func findAndroidAppName(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "h1" {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n := findAndroidAppName(c); n != nil {
			return n
		}
	}

	return nil
}

func findAndroidAppIcon(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "img" && hasHTMLAttribute(node, "itemprop", "image") {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n := findAndroidAppIcon(c); n != nil {
			return n
		}
	}

	return nil
}

func findTextNode(node *html.Node) *html.Node {
	if node.Type == html.TextNode {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n := findTextNode(c); n != nil {
			return n
		}
	}

	return nil
}

func hasHTMLAttribute(node *html.Node, key, value string) bool {
	for _, attr := range node.Attr {
		if attr.Key == key && attr.Val == value {
			return true
		}
	}

	return false
}

func getHTMLAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}

func containsString(list []string, str string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}

	return false
}
