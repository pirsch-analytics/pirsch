package pkg

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

	// BrowserArc represents the Arc browser.
	BrowserArc = "Arc"

	// BrowserDuckDuckGo represents the DuckDuckGo browser.
	BrowserDuckDuckGo = "DuckDuckGo"

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

	// OSChrome represents the Chrome operating system.
	OSChrome = "Chrome OS"
)

const (
	// TableSessions is the sessions table name.
	TableSessions = "session_v7"

	// TablePageViews is the page views table name.
	TablePageViews = "page_view_v7"

	// TableEvents is the events table name.
	TableEvents = "event_v7"

	// TableImportedBrowser is an imported statistics table.
	TableImportedBrowser = "imported_browser"

	// TableImportedCity is an imported statistics table.
	TableImportedCity = "imported_city"

	// TableImportedCountry is an imported statistics table.
	TableImportedCountry = "imported_country"

	// TableImportedDevice is an imported statistics table.
	TableImportedDevice = "imported_device"

	// TableImportedEntryPage is an imported statistics table.
	TableImportedEntryPage = "imported_entry_page"

	// TableImportedExitPage is an imported statistics table.
	TableImportedExitPage = "imported_exit_page"

	// TableImportedLanguage is an imported statistics table.
	TableImportedLanguage = "imported_language"

	// TableImportedOS is an imported statistics table.
	TableImportedOS = "imported_os"

	// TableImportedPage is an imported statistics table.
	TableImportedPage = "imported_page"

	// TableImportedReferrer is an imported statistics table.
	TableImportedReferrer = "imported_referrer"

	// TableImportedRegion is an imported statistics table.
	TableImportedRegion = "imported_region"

	// TableImportedUTMCampaign is an imported statistics table.
	TableImportedUTMCampaign = "imported_utm_campaign"

	// TableImportedUTMMedium is an imported statistics table.
	TableImportedUTMMedium = "imported_utm_medium"

	// TableImportedUTMSource is an imported statistics table.
	TableImportedUTMSource = "imported_utm_source"

	// TableImportedVisitors is an imported statistics table.
	TableImportedVisitors = "imported_visitors"
)

const (
	// PlatformUnknown is the platform for an unknown device.
	PlatformUnknown = int8(iota)

	// PlatformDesktop is the platform for a desktop device.
	PlatformDesktop

	// PlatformMobile is the platform for a mobile device.
	PlatformMobile
)

var (
	// ImportedTables is a list of all imported statistics tables.
	ImportedTables = []string{
		TableImportedBrowser,
		TableImportedCity,
		TableImportedCountry,
		TableImportedDevice,
		TableImportedEntryPage,
		TableImportedExitPage,
		TableImportedLanguage,
		TableImportedOS,
		TableImportedPage,
		TableImportedReferrer,
		TableImportedRegion,
		TableImportedUTMCampaign,
		TableImportedUTMMedium,
		TableImportedUTMSource,
		TableImportedVisitors,
	}
)
