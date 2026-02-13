package analyzer

import (
	"fmt"
	"testing"
	"time"

	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestAnalyzer_Platform(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: time.Now(), Path: "/", Desktop: true},
		{VisitorID: 1, Time: time.Now(), Path: "/foo", Desktop: true},
		{VisitorID: 1, Time: time.Now(), Path: "/bar", Desktop: true},
		{VisitorID: 2, Time: time.Now(), Path: "/", Mobile: true},
		{VisitorID: 3, Time: time.Now(), Path: "/", Mobile: true},
		{VisitorID: 4, Time: time.Now(), Path: "/"},
		{VisitorID: 5, Time: time.Now(), Path: "/", Desktop: true},
		{VisitorID: 6, Time: time.Now(), Path: "/", Desktop: true},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Desktop: true},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Desktop: true},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Desktop: true},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Mobile: true},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Mobile: true},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now()},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), Desktop: true},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), Desktop: true},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	platform, err := analyzer.Device.Platform(&Filter{
		From: util.PastDay(5),
		To:   util.Today(),
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, platform.PlatformDesktop)
	assert.Equal(t, 2, platform.PlatformMobile)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0.5, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0.1666, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		Path: []string{"/foo"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, platform.PlatformDesktop)
	assert.Equal(t, 0, platform.PlatformMobile)
	assert.Equal(t, 0, platform.PlatformUnknown)
	assert.InDelta(t, 0.1666, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		Platform: pkg.PlatformDesktop,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, platform.PlatformDesktop)
	assert.Equal(t, 0, platform.PlatformMobile)
	assert.Equal(t, 0, platform.PlatformUnknown)
	assert.InDelta(t, 0.5, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		Platform: pkg.PlatformMobile,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, platform.PlatformDesktop)
	assert.Equal(t, 2, platform.PlatformMobile)
	assert.Equal(t, 0, platform.PlatformUnknown)
	assert.InDelta(t, 0, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		Platform: pkg.PlatformUnknown,
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, platform.PlatformDesktop)
	assert.Equal(t, 0, platform.PlatformMobile)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0.1666, platform.RelativePlatformUnknown, 0.01)
	_, err = analyzer.Device.Platform(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.Platform(getMaxFilter("event"))
	assert.NoError(t, err)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_device" (date, category, visitors) VALUES
		('%s', 'Desktop', 2), ('%s', 'mobile', 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	platform, err = analyzer.Device.Platform(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, platform.PlatformDesktop)
	assert.Equal(t, 3, platform.PlatformMobile)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0.5555, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0.1111, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Platform:      pkg.PlatformDesktop,
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, platform.PlatformDesktop)
	assert.InDelta(t, 0.5555, platform.RelativePlatformDesktop, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Platform:      pkg.PlatformMobile,
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, platform.PlatformMobile)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Platform:      pkg.PlatformUnknown,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0.1111, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{
		From:                 util.PastDay(1),
		To:                   util.Today(),
		ImportedUntil:        util.Today(),
		Period:               pkg.PeriodMonth,
		Limit:                10,
		Sample:               10_000_000,
		MaxTimeOnPageSeconds: 3600,
		Path:                 []string{"/"},
	})
	assert.NoError(t, err)
	assert.Equal(t, 5, platform.PlatformDesktop)
	assert.Equal(t, 3, platform.PlatformMobile)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0.5555, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0.1111, platform.RelativePlatformUnknown, 0.01)
	maxFilter := getMaxFilter("event")
	maxFilter.ImportedUntil = util.Today()
	_, err = analyzer.Device.Platform(maxFilter)
	assert.NoError(t, err)
	_, err = analyzer.Device.Platform(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		PathPattern:   []string{"(?i)^/.*$"},
		Sample:        10_000_000,
	})
	assert.NoError(t, err)
}

func TestAnalyzer_Browser(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserEdge},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserEdge},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserFirefox},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserFirefox},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserSafari},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Device.Browser(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, pkg.BrowserChrome, visitors[0].Browser)
	assert.Equal(t, pkg.BrowserFirefox, visitors[1].Browser)
	assert.Equal(t, pkg.BrowserSafari, visitors[2].Browser)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Device.Browser(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.Browser(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Device.Browser(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldBrowser,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldBrowser,
			Input: "Firefox",
		},
	}})
	assert.NoError(t, err)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_browser" (date, browser, visitors) VALUES
		('%s', 'Chrome', 2), ('%s', 'Firefox', 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Device.Browser(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, pkg.BrowserChrome, visitors[0].Browser)
	assert.Equal(t, pkg.BrowserFirefox, visitors[1].Browser)
	assert.Equal(t, pkg.BrowserSafari, visitors[2].Browser)
	assert.Equal(t, 5, visitors[0].Visitors)
	assert.Equal(t, 3, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5555, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1111, visitors[2].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.Browser(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Browser:       []string{pkg.BrowserFirefox},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, pkg.BrowserFirefox, visitors[0].Browser)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.33, visitors[0].RelativeVisitors, 0.01)
}

func TestAnalyzer_BrowserVersion(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserEdge, BrowserVersion: "85.0"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserEdge, BrowserVersion: "85.0"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserFirefox, BrowserVersion: "89.0.0"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserFirefox, BrowserVersion: "89.0.1"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserSafari, BrowserVersion: "14.1.2"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome, BrowserVersion: "87.2"},
			{Sign: 1, VisitorID: 7, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome, BrowserVersion: "86.0"},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Device.BrowserVersion(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, pkg.BrowserChrome, visitors[0].Browser)
	assert.Equal(t, pkg.BrowserChrome, visitors[1].Browser)
	assert.Equal(t, pkg.BrowserChrome, visitors[2].Browser)
	assert.Equal(t, pkg.BrowserFirefox, visitors[3].Browser)
	assert.Equal(t, pkg.BrowserFirefox, visitors[4].Browser)
	assert.Equal(t, pkg.BrowserSafari, visitors[5].Browser)
	assert.Equal(t, "85.1", visitors[0].BrowserVersion)
	assert.Equal(t, "86.0", visitors[1].BrowserVersion)
	assert.Equal(t, "87.2", visitors[2].BrowserVersion)
	assert.Equal(t, "89.0.0", visitors[3].BrowserVersion)
	assert.Equal(t, "89.0.1", visitors[4].BrowserVersion)
	assert.Equal(t, "14.1.2", visitors[5].BrowserVersion)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.InDelta(t, 0.2857, visitors[0].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[1].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[2].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[3].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[4].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[5].RelativeVisitors, 0.001)
	_, err = analyzer.Device.BrowserVersion(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.BrowserVersion(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Device.BrowserVersion(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldBrowserVersion,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldBrowserVersion,
			Input: "100.0",
		},
	}})
	assert.NoError(t, err)
}

func TestAnalyzer_BrowserVersionSearchSort(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserFirefox, BrowserVersion: "89.0.0"},
			{Sign: 1, VisitorID: 7, Time: time.Now(), Start: time.Now(), Browser: pkg.BrowserChrome, BrowserVersion: "85.1"},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Device.BrowserVersion(&Filter{Sort: []Sort{
		{
			Field:     FieldBrowserVersion,
			Direction: pkg.DirectionASC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "85.1", visitors[0].BrowserVersion)
	assert.Equal(t, "89.0.0", visitors[1].BrowserVersion)
	visitors, err = analyzer.Device.BrowserVersion(&Filter{Sort: []Sort{
		{
			Field:     FieldBrowserVersion,
			Direction: pkg.DirectionDESC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "89.0.0", visitors[0].BrowserVersion)
	assert.Equal(t, "85.1", visitors[1].BrowserVersion)
	visitors, err = analyzer.Device.BrowserVersion(&Filter{Search: []Search{
		{
			Field: FieldBrowserVersion,
			Input: "89",
		},
	}, Sort: []Sort{
		{
			Field:     FieldBrowserVersion,
			Direction: pkg.DirectionDESC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "89.0.0", visitors[0].BrowserVersion)
}

func TestAnalyzer_OS(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pkg.OSLinux},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pkg.OSLinux},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), OS: pkg.OSMac},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), OS: pkg.OSMac},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), OS: pkg.OSAndroid},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Device.OS(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, pkg.OSWindows, visitors[0].OS)
	assert.Equal(t, pkg.OSMac, visitors[1].OS)
	assert.Equal(t, pkg.OSAndroid, visitors[2].OS)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Device.OS(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.OS(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Device.OS(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldOS,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldOS,
			Input: "Windows",
		},
	}})
	assert.NoError(t, err)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_os" (date, os, visitors) VALUES
		('%s', 'Windows', 2), ('%s', 'Mac', 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Device.OS(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, pkg.OSWindows, visitors[0].OS)
	assert.Equal(t, pkg.OSMac, visitors[1].OS)
	assert.Equal(t, pkg.OSAndroid, visitors[2].OS)
	assert.Equal(t, 5, visitors[0].Visitors)
	assert.Equal(t, 3, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5555, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1111, visitors[2].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.OS(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		OS:            []string{pkg.OSMac},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, pkg.OSMac, visitors[0].OS)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.33, visitors[0].RelativeVisitors, 0.01)
}

func TestAnalyzer_OSVersion(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pkg.OSLinux, OSVersion: "1"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pkg.OSLinux, OSVersion: "1"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), OS: pkg.OSMac, OSVersion: "14.0.0"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), OS: pkg.OSMac, OSVersion: "13.1.0"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), OS: pkg.OSLinux},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows, OSVersion: "9"},
			{Sign: 1, VisitorID: 7, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows, OSVersion: "8"},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Device.OSVersion(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, pkg.OSWindows, visitors[0].OS)
	assert.Equal(t, pkg.OSLinux, visitors[1].OS)
	assert.Equal(t, pkg.OSMac, visitors[2].OS)
	assert.Equal(t, pkg.OSMac, visitors[3].OS)
	assert.Equal(t, pkg.OSWindows, visitors[4].OS)
	assert.Equal(t, pkg.OSWindows, visitors[5].OS)
	assert.Equal(t, "10", visitors[0].OSVersion)
	assert.Empty(t, visitors[1].OSVersion)
	assert.Equal(t, "13.1.0", visitors[2].OSVersion)
	assert.Equal(t, "14.0.0", visitors[3].OSVersion)
	assert.Equal(t, "8", visitors[4].OSVersion)
	assert.Equal(t, "9", visitors[5].OSVersion)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.Equal(t, 1, visitors[4].Visitors)
	assert.Equal(t, 1, visitors[5].Visitors)
	assert.InDelta(t, 0.2857, visitors[0].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[1].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[2].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[3].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[4].RelativeVisitors, 0.001)
	assert.InDelta(t, 0.1428, visitors[5].RelativeVisitors, 0.001)
	_, err = analyzer.Device.OSVersion(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.OSVersion(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Device.OSVersion(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldOSVersion,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldOSVersion,
			Input: "10.0",
		},
	}})
	assert.NoError(t, err)
}

func TestAnalyzer_OSVersionSearchSort(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), OS: pkg.OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), OS: pkg.OSMac, OSVersion: "14.0.0"},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Device.OSVersion(&Filter{Sort: []Sort{
		{
			Field:     FieldOSVersion,
			Direction: pkg.DirectionASC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "10", visitors[0].OSVersion)
	assert.Equal(t, "14.0.0", visitors[1].OSVersion)
	visitors, err = analyzer.Device.OSVersion(&Filter{Sort: []Sort{
		{
			Field:     FieldOSVersion,
			Direction: pkg.DirectionDESC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "14.0.0", visitors[0].OSVersion)
	assert.Equal(t, "10", visitors[1].OSVersion)
	visitors, err = analyzer.Device.OSVersion(&Filter{Search: []Search{
		{
			Field: FieldOSVersion,
			Input: "14",
		},
	}, Sort: []Sort{
		{
			Field:     FieldOSVersion,
			Direction: pkg.DirectionDESC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "14.0.0", visitors[0].OSVersion)
}

func TestAnalyzer_ScreenClass(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), ScreenClass: "S"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), ScreenClass: "S"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), ScreenClass: "XXL"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), ScreenClass: "XL"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), ScreenClass: "XL"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), ScreenClass: "L"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), ScreenClass: "XXL"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), ScreenClass: "XXL"},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: time.Now(), Path: "/", ScreenClass: "XXL", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: time.Now(), Path: "/foo", ScreenClass: "XXL", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: time.Now(), Path: "/bar", ScreenClass: "XXL", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 2, Time: time.Now(), Path: "/", ScreenClass: "XL"},
		{VisitorID: 3, Time: time.Now(), Path: "/", ScreenClass: "XL"},
		{VisitorID: 4, Time: time.Now(), Path: "/", ScreenClass: "L", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 5, Time: time.Now(), Path: "/", ScreenClass: "XXL"},
		{VisitorID: 6, Time: time.Now(), Path: "/", ScreenClass: "XXL", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Device.ScreenClass(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "XXL", visitors[0].ScreenClass)
	assert.Equal(t, "XL", visitors[1].ScreenClass)
	assert.Equal(t, "L", visitors[2].ScreenClass)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.ScreenClass(&Filter{
		ScreenClass: []string{"XL"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "XL", visitors[0].ScreenClass)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.ScreenClass(&Filter{
		ScreenClass: []string{"L"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "L", visitors[0].ScreenClass)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.InDelta(t, 0.1666, visitors[0].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.ScreenClass(&Filter{
		Tag: []string{"author"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "XXL", visitors[0].ScreenClass)
	assert.Equal(t, "L", visitors[1].ScreenClass)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[1].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.ScreenClass(&Filter{
		Tag: []string{"!author"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "XL", visitors[0].ScreenClass)
	assert.Equal(t, "XXL", visitors[1].ScreenClass)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[1].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.ScreenClass(&Filter{
		Tag: []string{"author"},
		Tags: map[string]string{
			"author": "Alice",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "XXL", visitors[0].ScreenClass)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.InDelta(t, 0.1666, visitors[0].RelativeVisitors, 0.01)
	_, err = analyzer.Device.ScreenClass(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.ScreenClass(getMaxFilter("event"))
	assert.NoError(t, err)
}
