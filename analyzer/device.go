package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
)

// Device aggregates device statistics.
type Device struct {
	analyzer *Analyzer
	store    db.Store
}

// Platform returns the visitor count grouped by platform.
func (device *Device) Platform(filter *Filter) (*model.PlatformStats, error) {
	// TODO
	return &model.PlatformStats{}, nil

	/*filter = device.analyzer.getFilter(filter)
	table := filter.table()
	var args []any
	query := ""

	if table == "session" {
		filterArgs, filterQuery := filter.query(true)
		query = `SELECT uniqIf(visitor_id, desktop = 1) platform_desktop,
			uniqIf(visitor_id, mobile = 1) platform_mobile,
			uniq(visitor_id)-platform_desktop-platform_mobile platform_unknown,
			"platform_desktop" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_desktop,
			"platform_mobile" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_mobile,
			"platform_unknown" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_unknown
			FROM session s `

		if len(filter.Path) != 0 || len(filter.PathPattern) != 0 {
			entryPath, exitPath, eventName := filter.EntryPath, filter.ExitPath, filter.EventName
			filter.EntryPath, filter.ExitPath, filter.EventName = nil, nil, nil
			innerFilterArgs, innerFilterQuery := filter.query(false)
			filter.EntryPath, filter.ExitPath, filter.EventName = entryPath, exitPath, eventName
			args = append(args, innerFilterArgs...)
			query += fmt.Sprintf(`INNER JOIN (
				SELECT visitor_id,
				session_id,
				path
				FROM page_view
				WHERE %s
			) v
			ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerFilterQuery)
		}

		args = append(args, filterArgs...)
		query += fmt.Sprintf(`WHERE %s HAVING sum(sign) > 0`, filterQuery)
	} else {
		var innerArgs []any
		innerQuery := ""

		if device.analyzer.minIsBot > 0 || len(filter.EntryPath) != 0 || len(filter.ExitPath) != 0 {
			fields := make([]Field, 0, 2)

			if len(filter.EntryPath) != 0 {
				fields = append(fields, FieldEntryPath)
			}

			if len(filter.ExitPath) != 0 {
				fields = append(fields, FieldExitPath)
			}

			innerArgs, innerQuery = filter.joinSessions(table, fields)
			filter.EntryPath, filter.ExitPath = nil, nil
		}

		filterArgs, filterQuery := filter.query(false)
		args = make([]any, 0, len(filterArgs)*3+len(innerArgs)*3)
		args = append(args, innerArgs...)
		args = append(args, filterArgs...)
		args = append(args, innerArgs...)
		args = append(args, filterArgs...)
		args = append(args, innerArgs...)
		args = append(args, filterArgs...)
		query = fmt.Sprintf(`SELECT toInt64OrDefault((
				SELECT uniq(visitor_id)
				FROM event v
				%s
				WHERE %s
				AND desktop = 1
				AND mobile = 0
			)) platform_desktop,
			toInt64OrDefault((
				SELECT uniq(visitor_id)
				FROM event v
				%s
				WHERE %s
				AND desktop = 0
				AND mobile = 1
			)) platform_mobile,
			toInt64OrDefault((
				SELECT uniq(visitor_id)
				FROM event v
				%s
				WHERE %s
				AND desktop = 0
				AND mobile = 0
			)) platform_unknown,
			"platform_desktop" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_desktop,
			"platform_mobile" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_mobile,
			"platform_unknown" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_unknown `,
			innerQuery, filterQuery, innerQuery, filterQuery, innerQuery, filterQuery)
	}

	stats, err := device.store.GetPlatformStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil*/
}

// Browser returns the visitor count grouped by browser.
func (device *Device) Browser(filter *Filter) ([]model.BrowserStats, error) {
	args, query := device.analyzer.selectByAttribute(filter, FieldBrowser)
	return device.store.SelectBrowserStats(query, args...)
}

// OS returns the visitor count grouped by operating system.
func (device *Device) OS(filter *Filter) ([]model.OSStats, error) {
	args, query := device.analyzer.selectByAttribute(filter, FieldOS)
	return device.store.SelectOSStats(query, args...)
}

// OSVersion returns the visitor count grouped by operating systems and version.
func (device *Device) OSVersion(filter *Filter) ([]model.OSVersionStats, error) {
	args, query := device.analyzer.getFilter(filter).buildQuery([]Field{
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
	})
	stats, err := device.store.SelectOSVersionStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// BrowserVersion returns the visitor count grouped by browser and version.
func (device *Device) BrowserVersion(filter *Filter) ([]model.BrowserVersionStats, error) {
	args, query := device.analyzer.getFilter(filter).buildQuery([]Field{
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
	})
	stats, err := device.store.SelectBrowserVersionStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ScreenClass returns the visitor count grouped by screen class.
func (device *Device) ScreenClass(filter *Filter) ([]model.ScreenClassStats, error) {
	args, query := device.analyzer.selectByAttribute(filter, FieldScreenClass)
	return device.store.SelectScreenClassStats(query, args...)
}
