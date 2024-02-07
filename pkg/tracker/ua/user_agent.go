package ua

import (
	"github.com/emvi/null"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"time"
)

// UserAgent contains information extracted from the User-Agent header.
// The creation time and User-Agent are stored in the database to find bots.
type UserAgent struct {
	// Time is the creation date for the database record.
	Time time.Time

	// UserAgent is the raw User-Agent string.
	UserAgent string

	// Browser is the browser name.
	Browser string

	// BrowserVersion is the browser (non-technical) version number.
	BrowserVersion string

	// OS is the operating system.
	OS string

	// OSVersion is the operating system version number.
	OSVersion string

	// Mobile indicated whether this is a mobile device from client hint headers.
	// It'll be set to null if the header is not present or empty.
	Mobile null.Bool
}

// IsDesktop returns true if the user agent is a desktop device.
func (ua *UserAgent) IsDesktop() bool {
	if ua.Mobile.Valid {
		return !ua.Mobile.Bool
	}

	return ua.OS == pkg.OSWindows || ua.OS == pkg.OSMac || ua.OS == pkg.OSLinux
}

// IsMobile returns true if the user agent is a mobile device.
func (ua *UserAgent) IsMobile() bool {
	if ua.Mobile.Valid {
		return ua.Mobile.Bool
	}

	return ua.OS == pkg.OSAndroid || ua.OS == pkg.OSiOS || ua.OS == pkg.OSWindowsMobile
}
