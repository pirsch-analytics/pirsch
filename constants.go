package pirsch

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
)

const (
	// PeriodDay groups the results by date.
	PeriodDay = Period(iota)

	// PeriodWeek groups the results by week.
	PeriodWeek

	// PeriodMonth groups the results by month.
	PeriodMonth

	// PeriodYear groups the result by year.
	PeriodYear

	// PlatformDesktop filters for everything on desktops.
	PlatformDesktop = "desktop"

	// PlatformMobile filters for everything on mobile devices.
	PlatformMobile = "mobile"

	// PlatformUnknown filters for everything where the platform is unspecified.
	PlatformUnknown = "unknown"

	// Unknown filters for an unknown (empty) value.
	// This is a synonym for "null".
	Unknown = "null"

	// DirectionASC sorts results in ascending order.
	DirectionASC = Direction("ASC")

	// DirectionDESC sorts results in descending order.
	DirectionDESC = Direction("DESC")
)

const (
	// GeoLite2Filename is the default filename of the GeoLite2 database.
	GeoLite2Filename = "GeoLite2-City.mmdb"
)

// Period is used to group results.
type Period int

// Direction is used to sort results.
type Direction string

// NullClient is a placeholder for no client (0).
var NullClient = int64(0)
