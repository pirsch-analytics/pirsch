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

	// CustomMetricTypeInteger transforms the metadata value of an event to an integer (64 bit).
	CustomMetricTypeInteger = CustomMetricType("toInt64OrZero")

	// CustomMetricTypeFloat transforms the metadata value of an event to a floating point value (64 bit).
	CustomMetricTypeFloat = CustomMetricType("toFloat64OrZero")
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
)

// Period is used to group results.
type Period int

// Direction is used to sort results.
type Direction string

// CustomMetricType is used to set the type of a custom metric event meta value.
type CustomMetricType string

// NullClient is a placeholder for no client (0).
var NullClient = int64(0)
