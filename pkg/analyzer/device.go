package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
)

// Device aggregates device statistics.
type Device struct {
	analyzer *Analyzer
	store    db.Store
}

// Platform returns the visitor count grouped by platform.
func (device *Device) Platform(filter *Filter) (*model.PlatformStats, error) {
	filter = device.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldPlatformDesktop,
		FieldPlatformMobile,
		FieldPlatformUnknown,
		FieldRelativePlatformDesktop,
		FieldRelativePlatformMobile,
		FieldRelativePlatformUnknown,
	}, nil, nil, []Field{
		FieldPlatformDesktop,
		FieldPlatformMobile,
		FieldPlatformUnknown,
	}, "imported_device")
	stats, err := device.store.GetPlatformStats(filter.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Browser returns the visitor count grouped by browser.
func (device *Device) Browser(filter *Filter) ([]model.BrowserStats, error) {
	ctx, q, args := device.analyzer.selectByAttribute(filter, "imported_browser", FieldBrowser)
	return device.store.SelectBrowserStats(ctx, q, args...)
}

// OS returns the visitor count grouped by operating system.
func (device *Device) OS(filter *Filter) ([]model.OSStats, error) {
	ctx, q, args := device.analyzer.selectByAttribute(filter, "imported_os", FieldOS)
	return device.store.SelectOSStats(ctx, q, args...)
}

// OSVersion returns the visitor count grouped by operating systems and version.
func (device *Device) OSVersion(filter *Filter) ([]model.OSVersionStats, error) {
	filter = device.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldOS,
		FieldOSVersion,
		FieldVisitors,
		FieldRelativeVisitors,
	}, []Field{
		FieldOS,
		FieldOSVersion,
	}, []Field{
		FieldVisitors,
		FieldOS,
		FieldOSVersion,
	}, nil, "")
	stats, err := device.store.SelectOSVersionStats(filter.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// BrowserVersion returns the visitor count grouped by browser and version.
func (device *Device) BrowserVersion(filter *Filter) ([]model.BrowserVersionStats, error) {
	filter = device.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldBrowser,
		FieldBrowserVersion,
		FieldVisitors,
		FieldRelativeVisitors,
	}, []Field{
		FieldBrowser,
		FieldBrowserVersion,
	}, []Field{
		FieldVisitors,
		FieldBrowser,
		FieldBrowserVersion,
	}, nil, "")
	stats, err := device.store.SelectBrowserVersionStats(filter.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ScreenClass returns the visitor count grouped by screen class.
func (device *Device) ScreenClass(filter *Filter) ([]model.ScreenClassStats, error) {
	ctx, q, args := device.analyzer.selectByAttribute(filter, "", FieldScreenClass)
	return device.store.SelectScreenClassStats(ctx, q, args...)
}
