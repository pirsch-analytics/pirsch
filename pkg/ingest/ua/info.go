package ua

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

type info struct {
	browser        string
	browserVersion string
	os             string
	osVersion      string
	mobile         *bool
}

func (ua *info) isDesktop() bool {
	if ua.mobile != nil {
		return !*ua.mobile
	}

	return ua.os == pkg.OSWindows || ua.os == pkg.OSMac || ua.os == pkg.OSLinux
}

func (ua *info) isMobile() bool {
	if ua.mobile != nil {
		return *ua.mobile
	}

	return ua.os == pkg.OSAndroid || ua.os == pkg.OSiOS || ua.os == pkg.OSWindowsMobile
}
