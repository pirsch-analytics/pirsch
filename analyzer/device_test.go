package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_Platform(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: time.Now(), Path: "/"},
		{VisitorID: 1, Time: time.Now(), Path: "/foo"},
		{VisitorID: 1, Time: time.Now(), Path: "/bar"},
		{VisitorID: 2, Time: time.Now(), Path: "/"},
		{VisitorID: 3, Time: time.Now(), Path: "/"},
		{VisitorID: 4, Time: time.Now(), Path: "/"},
		{VisitorID: 5, Time: time.Now(), Path: "/"},
		{VisitorID: 6, Time: time.Now(), Path: "/"},
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
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	platform, err := analyzer.Device.Platform(&Filter{From: util.PastDay(5), To: util.Today()})
	assert.NoError(t, err)
	assert.Equal(t, 3, platform.PlatformDesktop)
	assert.Equal(t, 2, platform.PlatformMobile)
	assert.Equal(t, 1, platform.PlatformUnknown)
	assert.InDelta(t, 0.5, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0.3333, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0.1666, platform.RelativePlatformUnknown, 0.01)
	platform, err = analyzer.Device.Platform(&Filter{Path: []string{"/foo"}})
	assert.NoError(t, err)
	assert.Equal(t, 1, platform.PlatformDesktop)
	assert.Equal(t, 0, platform.PlatformMobile)
	assert.Equal(t, 0, platform.PlatformUnknown)
	assert.InDelta(t, 1, platform.RelativePlatformDesktop, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformMobile, 0.01)
	assert.InDelta(t, 0, platform.RelativePlatformUnknown, 0.01)
	_, err = analyzer.Device.Platform(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.Platform(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_Browser(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserEdge},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserEdge},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserChrome},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserFirefox},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserFirefox},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserSafari},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserChrome},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserChrome},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Device.Browser(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, pirsch.BrowserChrome, visitors[0].Browser)
	assert.Equal(t, pirsch.BrowserFirefox, visitors[1].Browser)
	assert.Equal(t, pirsch.BrowserSafari, visitors[2].Browser)
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
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldBrowser,
			Input: "Firefox",
		},
	}})
	assert.NoError(t, err)
}

func TestAnalyzer_BrowserVersion(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserEdge, BrowserVersion: "85.0"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserEdge, BrowserVersion: "85.0"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserChrome, BrowserVersion: "85.1"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserFirefox, BrowserVersion: "89.0.0"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserFirefox, BrowserVersion: "89.0.1"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserSafari, BrowserVersion: "14.1.2"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserChrome, BrowserVersion: "87.2"},
			{Sign: 1, VisitorID: 7, Time: time.Now(), Start: time.Now(), Browser: pirsch.BrowserChrome, BrowserVersion: "86.0"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Device.BrowserVersion(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, pirsch.BrowserChrome, visitors[0].Browser)
	assert.Equal(t, pirsch.BrowserChrome, visitors[1].Browser)
	assert.Equal(t, pirsch.BrowserChrome, visitors[2].Browser)
	assert.Equal(t, pirsch.BrowserFirefox, visitors[3].Browser)
	assert.Equal(t, pirsch.BrowserFirefox, visitors[4].Browser)
	assert.Equal(t, pirsch.BrowserSafari, visitors[5].Browser)
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
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldBrowserVersion,
			Input: "100.0",
		},
	}})
	assert.NoError(t, err)
}

func TestAnalyzer_OS(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pirsch.OSLinux},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pirsch.OSLinux},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pirsch.OSWindows},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), OS: pirsch.OSMac},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), OS: pirsch.OSMac},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), OS: pirsch.OSAndroid},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), OS: pirsch.OSWindows},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), OS: pirsch.OSWindows},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Device.OS(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, pirsch.OSWindows, visitors[0].OS)
	assert.Equal(t, pirsch.OSMac, visitors[1].OS)
	assert.Equal(t, pirsch.OSAndroid, visitors[2].OS)
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
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldOS,
			Input: "Windows",
		},
	}})
	assert.NoError(t, err)
}

func TestAnalyzer_OSVersion(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pirsch.OSLinux, OSVersion: "1"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pirsch.OSLinux, OSVersion: "1"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), OS: pirsch.OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), OS: pirsch.OSWindows, OSVersion: "10"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), OS: pirsch.OSMac, OSVersion: "14.0.0"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), OS: pirsch.OSMac, OSVersion: "13.1.0"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), OS: pirsch.OSLinux},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), OS: pirsch.OSWindows, OSVersion: "9"},
			{Sign: 1, VisitorID: 7, Time: time.Now(), Start: time.Now(), OS: pirsch.OSWindows, OSVersion: "8"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Device.OSVersion(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 6)
	assert.Equal(t, pirsch.OSWindows, visitors[0].OS)
	assert.Equal(t, pirsch.OSLinux, visitors[1].OS)
	assert.Equal(t, pirsch.OSMac, visitors[2].OS)
	assert.Equal(t, pirsch.OSMac, visitors[3].OS)
	assert.Equal(t, pirsch.OSWindows, visitors[4].OS)
	assert.Equal(t, pirsch.OSWindows, visitors[5].OS)
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
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldOSVersion,
			Input: "10.0",
		},
	}})
	assert.NoError(t, err)
}

func TestAnalyzer_ScreenClass(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), ScreenClass: "S", ScreenWidth: 415, ScreenHeight: 600},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), ScreenClass: "S", ScreenWidth: 415, ScreenHeight: 600},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), ScreenClass: "XXL", ScreenWidth: 3840, ScreenHeight: 2080},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), ScreenClass: "XL", ScreenWidth: 2560, ScreenHeight: 1440},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), ScreenClass: "XL", ScreenWidth: 2560, ScreenHeight: 1440},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), ScreenClass: "L", ScreenWidth: 1980, ScreenHeight: 1080},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), ScreenClass: "XXL", ScreenWidth: 3840, ScreenHeight: 2080},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), ScreenClass: "XXL", ScreenWidth: 3840, ScreenHeight: 2080},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
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
	visitors, err = analyzer.Device.ScreenClass(&Filter{ScreenWidth: []string{"2560"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "XL", visitors[0].ScreenClass)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
	visitors, err = analyzer.Device.ScreenClass(&Filter{ScreenHeight: []string{"1080"}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "L", visitors[0].ScreenClass)
	assert.Equal(t, 1, visitors[0].Visitors)
	assert.InDelta(t, 0.1666, visitors[0].RelativeVisitors, 0.01)
	_, err = analyzer.Device.ScreenClass(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Device.ScreenClass(getMaxFilter("event"))
	assert.NoError(t, err)
}
