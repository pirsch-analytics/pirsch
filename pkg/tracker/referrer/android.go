package referrer

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	androidAppPrefix       = "android-app://"
	googlePlayStoreURL     = "https://play.google.com/store/apps/details?id=%s"
	androidAppCacheMaxSize = 10_000
	androidAppCacheMaxAge  = time.Hour * 24 * 7
)

var (
	androidAppCache = newAndroid()
)

type androidApp struct {
	name string
	icon string
}

type android struct {
	cache      map[string]androidApp
	maxSize    int
	maxAge     time.Duration
	nextUpdate time.Time
	m          sync.RWMutex
}

func newAndroid() *android {
	return &android{
		cache:      make(map[string]androidApp),
		maxSize:    androidAppCacheMaxSize,
		maxAge:     androidAppCacheMaxAge,
		nextUpdate: time.Now().UTC().Add(androidAppCacheMaxAge),
	}
}

func (cache *android) get(referrer string) (string, string) {
	packageName := referrer[len(androidAppPrefix):]

	if strings.HasSuffix(packageName, "/") {
		packageName = packageName[:len(packageName)-1]
	}

	cache.m.RLock()
	app, found := cache.cache[packageName]
	cache.m.RUnlock()

	if found {
		return app.name, app.icon
	}

	resp, err := http.Get(fmt.Sprintf(googlePlayStoreURL, packageName))

	if err != nil || resp.StatusCode != http.StatusOK {
		cache.updateApp(packageName, "", "")
		return "", ""
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	doc, err := html.Parse(resp.Body)

	if err != nil {
		cache.updateApp(packageName, "", "")
		return "", ""
	}

	titleNode := cache.findName(doc)

	if titleNode == nil {
		cache.updateApp(packageName, "", "")
		return "", ""
	}

	appName := cache.findTextNode(titleNode)

	if appName == nil {
		cache.updateApp(packageName, "", "")
		return "", ""
	}

	icon := ""
	iconNode := cache.findIcon(doc)

	if iconNode != nil {
		icon = cache.getHTMLAttribute(iconNode, "src")
	}

	cache.updateApp(packageName, appName.Data, icon)
	return appName.Data, icon
}

func (cache *android) updateApp(packageName, name, icon string) {
	cache.m.Lock()
	defer cache.m.Unlock()
	now := time.Now().UTC()

	if len(cache.cache) > cache.maxSize || now.After(cache.nextUpdate) {
		cache.cache = make(map[string]androidApp)
		cache.nextUpdate = now.Add(cache.maxAge)
	}

	cache.cache[packageName] = androidApp{name, icon}
}

func (cache *android) findName(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "h1" {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n := cache.findName(c); n != nil {
			return n
		}
	}

	return nil
}

func (cache *android) findIcon(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "img" && cache.hasHTMLAttribute(node, "itemprop", "image") {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n := cache.findIcon(c); n != nil {
			return n
		}
	}

	return nil
}

func (cache *android) findTextNode(node *html.Node) *html.Node {
	if node.Type == html.TextNode {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if n := cache.findTextNode(c); n != nil {
			return n
		}
	}

	return nil
}

func (cache *android) hasHTMLAttribute(node *html.Node, key, value string) bool {
	for _, attr := range node.Attr {
		if attr.Key == key && attr.Val == value {
			return true
		}
	}

	return false
}

func (cache *android) getHTMLAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}
