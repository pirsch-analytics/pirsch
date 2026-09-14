package ua

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
)

type info struct {
	browser         string
	browserVersion  string
	browserRevision string
	os              string
	osVersion       string
	mobile          *bool
}

func (ua *info) platform() int8 {
	if ua.mobile != nil {
		if *ua.mobile {
			return pkg.PlatformMobile
		}
	}

	if ua.os == pkg.OSWindows || ua.os == pkg.OSMac || ua.os == pkg.OSLinux {
		return pkg.PlatformDesktop
	}

	return pkg.PlatformUnknown
}
