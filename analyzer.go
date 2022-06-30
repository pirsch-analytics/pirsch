package pirsch

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrNoPeriodOrDay is returned in case no period or day was specified to calculate the growth rate.
	ErrNoPeriodOrDay = errors.New("no period or day specified")
)

// AnalyzerConfig is the optional configuration for the Analyzer.
type AnalyzerConfig struct {
	// IsBotThreshold see HitOptions.IsBotThreshold.
	IsBotThreshold uint8

	// DisableBotFilter disables IsBotThreshold (otherwise these would be set to the default value).
	DisableBotFilter bool
}

func (config *AnalyzerConfig) validate() {
	if config.DisableBotFilter {
		config.IsBotThreshold = 0
	} else if config.IsBotThreshold == 0 {
		config.IsBotThreshold = defaultIsBotThreshold
	}
}

// Analyzer provides an interface to analyze statistics.
type Analyzer struct {
	store    Store
	minIsBot uint8
}

// NewAnalyzer returns a new Analyzer for given Store.
func NewAnalyzer(store Store, config *AnalyzerConfig) *Analyzer {
	if config == nil {
		config = new(AnalyzerConfig)
	}

	config.validate()
	return &Analyzer{
		store,
		config.IsBotThreshold,
	}
}

// ActiveVisitors returns the active visitors per path and (optional) page title and the total number of active visitors for given duration.
// Use time.Minute*5 for example to get the active visitors for the past 5 minutes.
func (analyzer *Analyzer) ActiveVisitors(filter *Filter, duration time.Duration) ([]ActiveVisitorStats, int, error) {
	filter = analyzer.getFilter(filter)
	filter.From = time.Now().UTC().Add(-duration)
	filter.IncludeTime = true
	title := ""

	if filter.IncludeTitle {
		title = ",title"
	}

	filterArgs, filterQuery := filter.query(false)
	innerFilterArgs, innerFilterQuery := filter.queryTime(true)
	args := make([]any, 0, len(innerFilterArgs)+len(filterArgs))
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT path %s,
		uniq(visitor_id) visitors
		FROM page_view v `, title))

	if analyzer.minIsBot > 0 || filter.EntryPath != "" || filter.ExitPath != "" {
		args = append(args, innerFilterArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id, entry_path, exit_path
			HAVING sum(sign) > 0
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerFilterQuery))
	}

	args = append(args, filterArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
		GROUP BY path %s
		ORDER BY visitors DESC, path
		%s`, filterQuery, title, filter.withLimit()))
	stats, err := analyzer.store.SelectActiveVisitorStats(title != "", query.String(), args...)

	if err != nil {
		return nil, 0, err
	}

	query.Reset()
	query.WriteString(`SELECT uniq(visitor_id) visitors
		FROM page_view v `)

	if analyzer.minIsBot > 0 || filter.EntryPath != "" || filter.ExitPath != "" {
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id, entry_path, exit_path
			HAVING sum(sign) > 0
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerFilterQuery))
	}

	query.WriteString(fmt.Sprintf(`WHERE %s`, filterQuery))
	count, err := analyzer.store.Count(query.String(), args...)

	if err != nil {
		return nil, 0, err
	}

	return stats, count, nil
}

// TotalVisitors returns the total visitor count, session count, bounce rate, and views.
func (analyzer *Analyzer) TotalVisitors(filter *Filter) (*TotalVisitorStats, error) {
	args, query := analyzer.getFilter(filter).buildQuery([]Field{
		FieldVisitors,
		FieldSessions,
		FieldViews,
		FieldBounces,
		FieldBounceRate,
	}, nil, nil)
	stats, err := analyzer.store.GetTotalVisitorStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Visitors returns the visitor count, session count, bounce rate, and views grouped by day, week, month, or year.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]VisitorStats, error) {
	filter = analyzer.getFilter(filter)
	args, query := filter.buildQuery([]Field{
		FieldDay,
		FieldVisitors,
		FieldSessions,
		FieldViews,
		FieldBounces,
		FieldBounceRate,
	}, []Field{
		FieldDay,
	}, []Field{
		FieldDay,
		FieldVisitors,
	})
	stats, err := analyzer.store.SelectVisitorStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Growth returns the growth rate for visitor count, session count, bounces, views, and average session duration or average time on page (if path is set).
// The growth rate is relative to the previous time range or day.
// The period or day for the filter must be set, else an error is returned.
func (analyzer *Analyzer) Growth(filter *Filter) (*Growth, error) {
	filter = analyzer.getFilter(filter)

	if filter.From.IsZero() || filter.To.IsZero() {
		return nil, ErrNoPeriodOrDay
	}

	// get current statistics
	fields := []Field{
		FieldVisitors,
		FieldSessions,
		FieldViews,
		FieldBounces,
		FieldBounceRate,
	}
	args, query := filter.buildQuery(fields, nil, nil)
	current, err := analyzer.store.GetGrowthStats(query, args...)

	if err != nil {
		return nil, err
	}

	var currentTimeSpent int

	if filter.EventName != "" {
		currentTimeSpent, err = analyzer.totalEventDuration(filter)
	} else if filter.Path == "" {
		currentTimeSpent, err = analyzer.totalSessionDuration(filter)
	} else {
		currentTimeSpent, err = analyzer.totalTimeOnPage(filter)
	}

	if err != nil {
		return nil, err
	}

	// get previous statistics
	if filter.From.Equal(filter.To) {
		if filter.To.Equal(Today()) {
			filter.From = filter.From.Add(-time.Hour * 24 * 7)
			filter.To = time.Now().UTC().Add(-time.Hour * 24 * 7)
			filter.IncludeTime = true
		} else {
			filter.From = filter.From.Add(-time.Hour * 24 * 7)
			filter.To = filter.To.Add(-time.Hour * 24 * 7)
		}
	} else {
		days := filter.To.Sub(filter.From)

		if days >= time.Hour*24 {
			filter.To = filter.From.Add(-time.Hour * 24)
			filter.From = filter.To.Add(-days)
		} else {
			filter.From = filter.From.Add(-time.Hour * 24)
			filter.To = filter.To.Add(-time.Hour * 24)
		}
	}

	args, query = filter.buildQuery(fields, nil, nil)
	previous, err := analyzer.store.GetGrowthStats(query, args...)

	if err != nil {
		return nil, err
	}

	var previousTimeSpent int

	if filter.EventName != "" {
		previousTimeSpent, err = analyzer.totalEventDuration(filter)
	} else if filter.Path == "" {
		previousTimeSpent, err = analyzer.totalSessionDuration(filter)
	} else {
		previousTimeSpent, err = analyzer.totalTimeOnPage(filter)
	}

	if err != nil {
		return nil, err
	}

	return &Growth{
		VisitorsGrowth:  calculateGrowth(current.Visitors, previous.Visitors),
		ViewsGrowth:     calculateGrowth(current.Views, previous.Views),
		SessionsGrowth:  calculateGrowth(current.Sessions, previous.Sessions),
		BouncesGrowth:   calculateGrowth(current.BounceRate, previous.BounceRate),
		TimeSpentGrowth: calculateGrowth(currentTimeSpent, previousTimeSpent),
	}, nil
}

// VisitorHours returns the visitor count grouped by time of day.
func (analyzer *Analyzer) VisitorHours(filter *Filter) ([]VisitorHourStats, error) {
	args, query := analyzer.getFilter(filter).buildQuery([]Field{
		FieldHour,
		FieldVisitors,
		FieldSessions,
		FieldViews,
		FieldBounces,
		FieldBounceRate,
	}, []Field{
		FieldHour,
	}, []Field{
		FieldHour,
		FieldVisitors,
	})
	stats, err := analyzer.store.SelectVisitorHourStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Pages returns the visitor count, session count, bounce rate, views, and average time on page grouped by path and (optional) page title.
func (analyzer *Analyzer) Pages(filter *Filter) ([]PageStats, error) {
	filter = analyzer.getFilter(filter)
	fields := []Field{
		FieldPath,
		FieldVisitors,
		FieldSessions,
		FieldRelativeVisitors,
		FieldViews,
		FieldRelativeViews,
		FieldBounces,
		FieldBounceRate,
	}
	groupBy := []Field{
		FieldPath,
	}
	orderBy := []Field{
		FieldVisitors,
		FieldPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldTitle)
		groupBy = append(groupBy, FieldTitle)
		orderBy = append(orderBy, FieldTitle)
	}

	if filter.table() == "event" {
		fields = append(fields, FieldEventTimeSpent)
	}

	args, query := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := analyzer.store.SelectPageStats(filter.IncludeTitle, filter.table() == "event", query, args...)

	if err != nil {
		return nil, err
	}

	if filter.IncludeTimeOnPage && filter.table() == "session" {
		paths := make(map[string]struct{})

		for i := range stats {
			paths[stats[i].Path] = struct{}{}
		}

		pathList := make([]string, 0, len(paths))

		for path := range paths {
			pathList = append(pathList, path)
		}

		top, err := analyzer.avgTimeOnPage(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range top {
				if stats[i].Path == top[j].Path {
					stats[i].AverageTimeSpentSeconds = top[j].AverageTimeSpentSeconds
					break
				}
			}
		}
	}

	return stats, nil
}

// EntryPages returns the visitor count and time on page grouped by path and (optional) page title for the first page visited.
func (analyzer *Analyzer) EntryPages(filter *Filter) ([]EntryStats, error) {
	filter = analyzer.getFilter(filter)

	fields := []Field{
		FieldEntryPath,
		FieldEntries,
	}
	groupBy := []Field{
		FieldEntryPath,
	}
	orderBy := []Field{
		FieldEntries,
		FieldEntryPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldEntryTitle)
		groupBy = append(groupBy, FieldEntryTitle)
		orderBy = append(orderBy, FieldEntryTitle)
	}

	args, query := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := analyzer.store.SelectEntryStats(filter.IncludeTitle, query, args...)

	if err != nil {
		return nil, err
	}

	paths := make(map[string]struct{})

	for i := range stats {
		paths[stats[i].Path] = struct{}{}
	}

	pathList := make([]string, 0, len(paths))

	for path := range paths {
		pathList = append(pathList, path)
	}

	if filter.table() != "event" {
		total, err := analyzer.totalVisitorsSessions(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range total {
				if stats[i].Path == total[j].Path {
					stats[i].Visitors = total[j].Visitors
					stats[i].Sessions = total[j].Sessions
					stats[i].EntryRate = float64(stats[i].Entries) / float64(total[j].Sessions)
					break
				}
			}
		}
	}

	if filter.IncludeTimeOnPage && filter.table() != "event" {
		top, err := analyzer.avgTimeOnPage(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range top {
				if stats[i].Path == top[j].Path {
					stats[i].AverageTimeSpentSeconds = top[j].AverageTimeSpentSeconds
					break
				}
			}
		}
	}

	return stats, nil
}

// ExitPages returns the visitor count and time on page grouped by path and (optional) page title for the last page visited.
func (analyzer *Analyzer) ExitPages(filter *Filter) ([]ExitStats, error) {
	filter = analyzer.getFilter(filter)

	fields := []Field{
		FieldExitPath,
		FieldExits,
	}
	groupBy := []Field{
		FieldExitPath,
	}
	orderBy := []Field{
		FieldExits,
		FieldExitPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldExitTitle)
		groupBy = append(groupBy, FieldExitTitle)
		orderBy = append(orderBy, FieldExitTitle)
	}

	args, query := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := analyzer.store.SelectExitStats(filter.IncludeTitle, query, args...)

	if err != nil {
		return nil, err
	}

	if filter.table() != "event" {
		paths := make(map[string]struct{})

		for i := range stats {
			paths[stats[i].Path] = struct{}{}
		}

		pathList := make([]string, 0, len(paths))

		for path := range paths {
			pathList = append(pathList, path)
		}

		total, err := analyzer.totalVisitorsSessions(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range total {
				if stats[i].Path == total[j].Path {
					stats[i].Visitors = total[j].Visitors
					stats[i].Sessions = total[j].Sessions
					stats[i].ExitRate = float64(stats[i].Exits) / float64(total[j].Sessions)
					break
				}
			}
		}
	}

	return stats, nil
}

// PageConversions returns the visitor count, views, and conversion rate for conversion goals.
// This function is supposed to be used with the Filter.PathPattern, to list page conversions.
func (analyzer *Analyzer) PageConversions(filter *Filter) (*PageConversionsStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.PathPattern == "" {
		return nil, nil
	}

	args, query := filter.buildQuery([]Field{
		FieldVisitors,
		FieldViews,
		FieldCR,
	}, nil, []Field{
		FieldVisitors,
	})
	stats, err := analyzer.store.GetPageConversionsStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Events returns the visitor count, views, and conversion rate for custom events.
func (analyzer *Analyzer) Events(filter *Filter) ([]EventStats, error) {
	filter = analyzer.getFilter(filter)
	filter.eventFilter = true
	args, query := filter.buildQuery([]Field{
		FieldEventName,
		FieldVisitors,
		FieldViews,
		FieldCR,
		FieldEventTimeSpent,
		FieldEventMetaKeys,
	}, []Field{
		FieldEventName,
	}, []Field{
		FieldVisitors,
		FieldEventName,
	})
	stats, err := analyzer.store.SelectEventStats(false, query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// EventBreakdown returns the visitor count, views, and conversion rate for a custom event grouping them by a meta value for given key.
// The Filter.EventName and Filter.EventMetaKey must be set, or otherwise the result set will be empty.
func (analyzer *Analyzer) EventBreakdown(filter *Filter) ([]EventStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.EventName == "" || filter.EventMetaKey == "" {
		return []EventStats{}, nil
	}

	args, query := filter.buildQuery([]Field{
		FieldEventName,
		FieldVisitors,
		FieldViews,
		FieldCR,
		FieldEventTimeSpent,
		FieldEventMetaValues,
	}, []Field{
		FieldEventName,
		FieldEventMetaValues,
	}, []Field{
		FieldVisitors,
		FieldEventMetaValues,
	})
	stats, err := analyzer.store.SelectEventStats(true, query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// EventList returns events as a list. The metadata is grouped as key-value pairs.
func (analyzer *Analyzer) EventList(filter *Filter) ([]EventListStats, error) {
	filter = analyzer.getFilter(filter)
	filter.eventFilter = true
	args, query := filter.buildQuery([]Field{
		FieldEventName,
		FieldEventMeta,
		FieldVisitors,
		FieldCount,
	}, []Field{
		FieldEventName,
		FieldEventMeta,
	}, []Field{
		FieldCount,
		FieldEventName,
	})
	stats, err := analyzer.store.SelectEventListStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Referrer returns the visitor count and bounce rate grouped by referrer.
func (analyzer *Analyzer) Referrer(filter *Filter) ([]ReferrerStats, error) {
	filter = analyzer.getFilter(filter)
	fields := []Field{
		FieldReferrerName,
		FieldReferrerIcon,
		FieldVisitors,
		FieldSessions,
		FieldRelativeVisitors,
		FieldBounces,
		FieldBounceRate,
	}
	groupBy := []Field{
		FieldReferrerName,
	}
	orderBy := []Field{
		FieldVisitors,
		FieldReferrerName,
	}

	if filter.Referrer != "" || filter.ReferrerName != "" {
		fields = append(fields, FieldReferrer)
		groupBy = append(groupBy, FieldReferrer)
		orderBy = append(orderBy, FieldReferrer)
	} else {
		fields = append(fields, FieldAnyReferrer)
	}

	args, query := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := analyzer.store.SelectReferrerStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Platform returns the visitor count grouped by platform.
func (analyzer *Analyzer) Platform(filter *Filter) (*PlatformStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	var args []any
	query := ""

	if table == "session" {
		filterArgs, filterQuery := filter.query(true)
		query = `SELECT sum(desktop*sign) platform_desktop,
			sum(mobile*sign) platform_mobile,
			sum(sign)-platform_desktop-platform_mobile platform_unknown,
			"platform_desktop" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_desktop,
			"platform_mobile" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_mobile,
			"platform_unknown" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_unknown
			FROM session s `

		if filter.Path != "" || filter.PathPattern != "" {
			entryPath, exitPath, eventName := filter.EntryPath, filter.ExitPath, filter.EventName
			filter.EntryPath, filter.ExitPath, filter.EventName = "", "", ""
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

		if analyzer.minIsBot > 0 || filter.EntryPath != "" || filter.ExitPath != "" {
			fields := make([]Field, 0, 2)

			if filter.EntryPath != "" {
				fields = append(fields, FieldEntryPath)
			}

			if filter.ExitPath != "" {
				fields = append(fields, FieldExitPath)
			}

			innerArgs, innerQuery = filter.joinSessions(table, fields)
			filter.EntryPath, filter.ExitPath = "", ""
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

	stats, err := analyzer.store.GetPlatformStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Languages returns the visitor count grouped by language.
func (analyzer *Analyzer) Languages(filter *Filter) ([]LanguageStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldLanguage)
	return analyzer.store.SelectLanguageStats(query, args...)
}

// Countries returns the visitor count grouped by country.
func (analyzer *Analyzer) Countries(filter *Filter) ([]CountryStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldCountry)
	return analyzer.store.SelectCountryStats(query, args...)
}

// Cities returns the visitor count grouped by city.
func (analyzer *Analyzer) Cities(filter *Filter) ([]CityStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldCity, FieldCountry)
	return analyzer.store.SelectCityStats(query, args...)
}

// Browser returns the visitor count grouped by browser.
func (analyzer *Analyzer) Browser(filter *Filter) ([]BrowserStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldBrowser)
	return analyzer.store.SelectBrowserStats(query, args...)
}

// OS returns the visitor count grouped by operating system.
func (analyzer *Analyzer) OS(filter *Filter) ([]OSStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldOS)
	return analyzer.store.SelectOSStats(query, args...)
}

// ScreenClass returns the visitor count grouped by screen class.
func (analyzer *Analyzer) ScreenClass(filter *Filter) ([]ScreenClassStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldScreenClass)
	return analyzer.store.SelectScreenClassStats(query, args...)
}

// UTMSource returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMSource(filter *Filter) ([]UTMSourceStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldUTMSource)
	return analyzer.store.SelectUTMSourceStats(query, args...)
}

// UTMMedium returns the visitor count grouped by utm medium.
func (analyzer *Analyzer) UTMMedium(filter *Filter) ([]UTMMediumStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldUTMMedium)
	return analyzer.store.SelectUTMMediumStats(query, args...)
}

// UTMCampaign returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMCampaign(filter *Filter) ([]UTMCampaignStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldUTMCampaign)
	return analyzer.store.SelectUTMCampaignStats(query, args...)
}

// UTMContent returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMContent(filter *Filter) ([]UTMContentStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldUTMContent)
	return analyzer.store.SelectUTMContentStats(query, args...)
}

// UTMTerm returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMTerm(filter *Filter) ([]UTMTermStats, error) {
	args, query := analyzer.selectByAttribute(filter, FieldUTMTerm)
	return analyzer.store.SelectUTMTermStats(query, args...)
}

// OSVersion returns the visitor count grouped by operating systems and version.
func (analyzer *Analyzer) OSVersion(filter *Filter) ([]OSVersionStats, error) {
	args, query := analyzer.getFilter(filter).buildQuery([]Field{
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
	stats, err := analyzer.store.SelectOSVersionStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// BrowserVersion returns the visitor count grouped by browser and version.
func (analyzer *Analyzer) BrowserVersion(filter *Filter) ([]BrowserVersionStats, error) {
	args, query := analyzer.getFilter(filter).buildQuery([]Field{
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
	stats, err := analyzer.store.SelectBrowserVersionStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgSessionDuration returns the average session duration grouped by day, week, month, or year.
func (analyzer *Analyzer) AvgSessionDuration(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []TimeSpentStats{}, nil
	}

	filterArgs, filterQuery := filter.query(true)
	innerFilterArgs, innerFilterQuery := filter.queryTime(false)
	withFillArgs, withFillQuery := filter.withFill()
	args := make([]any, 0, len(filterArgs)+len(innerFilterArgs)+len(withFillArgs))
	var query strings.Builder

	if filter.Period != PeriodDay {
		switch filter.Period {
		case PeriodWeek:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfWeek(day) week FROM (`)
		case PeriodMonth:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfMonth(day) month FROM (`)
		case PeriodYear:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfYear(day) year FROM (`)
		}
	}

	query.WriteString(fmt.Sprintf(`SELECT day, average_time_spent_seconds
		FROM (
			SELECT toDate(time, '%s') day,
			sum(duration_seconds*sign) duration,
			sum(sign) n,
			toUInt64(ifNotFinite(round(duration/n), 0)) average_time_spent_seconds
			FROM session s `, time.UTC.String()))

	if filter.Path != "" || filter.PathPattern != "" {
		args = append(args, innerFilterArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			path
			FROM page_view
			WHERE %s
		) v
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerFilterQuery))
	}

	args = append(args, filterArgs...)
	args = append(args, withFillArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
			AND duration_seconds != 0
			GROUP BY day
			HAVING sum(sign) > 0
			ORDER BY day
			%s
		)`, filterQuery, withFillQuery))

	if filter.Period != PeriodDay {
		switch filter.Period {
		case PeriodWeek:
			query.WriteString(`) GROUP BY week ORDER BY week ASC`)
		case PeriodMonth:
			query.WriteString(`) GROUP BY month ORDER BY month ASC`)
		case PeriodYear:
			query.WriteString(`) GROUP BY year ORDER BY year ASC`)
		}
	}

	stats, err := analyzer.store.SelectTimeSpentStats(filter.Period, query.String(), args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgTimeOnPage returns the average time on page grouped by day, week, month, or year.
func (analyzer *Analyzer) AvgTimeOnPage(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []TimeSpentStats{}, nil
	}

	timeArgs, timeQuery := filter.queryTime(false)
	fieldArgs, fieldQuery := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	withFillArgs, withFillQuery := filter.withFill()
	args := make([]any, 0, len(timeArgs)*2+len(fieldArgs)+len(withFillArgs))
	var query strings.Builder

	if filter.Period != PeriodDay {
		switch filter.Period {
		case PeriodWeek:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfWeek(day) week FROM (`)
		case PeriodMonth:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfMonth(day) month FROM (`)
		case PeriodYear:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfYear(day) year FROM (`)
		}
	}

	query.WriteString(fmt.Sprintf(`SELECT day,
		ifNull(toUInt64(avg(nullIf(time_on_page, 0))), 0) average_time_spent_seconds
		FROM (
			SELECT day,
			%s time_on_page
			FROM (
				SELECT session_id,
				toDate(time, '%s') day,
				duration_seconds
				%s
				FROM page_view v `, analyzer.timeOnPageQuery(filter), time.UTC.String(), fieldsQuery))

	if analyzer.minIsBot > 0 || filter.EntryPath != "" || filter.ExitPath != "" {
		innerTimeArgs, innerTimeQuery := filter.queryTime(false)
		args = append(args, innerTimeArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id, entry_path, exit_path
			HAVING sum(sign) > 0
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerTimeQuery))
	}

	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	args = append(args, withFillArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
				ORDER BY visitor_id, session_id, time
			)
			WHERE session_id = neighbor(session_id, 1, null)
			AND time_on_page > 0
			%s
		)
		GROUP BY day
		ORDER BY day
		%s`, timeQuery, fieldQuery, withFillQuery))

	if filter.Period != PeriodDay {
		switch filter.Period {
		case PeriodWeek:
			query.WriteString(`) GROUP BY week ORDER BY week ASC`)
		case PeriodMonth:
			query.WriteString(`) GROUP BY month ORDER BY month ASC`)
		case PeriodYear:
			query.WriteString(`) GROUP BY year ORDER BY year ASC`)
		}
	}

	stats, err := analyzer.store.SelectTimeSpentStats(filter.Period, query.String(), args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (analyzer *Analyzer) totalVisitorsSessions(filter *Filter, paths []string) ([]TotalVisitorSessionStats, error) {
	if len(paths) == 0 {
		return []TotalVisitorSessionStats{}, nil
	}

	filter = analyzer.getFilter(filter)
	filter.Path, filter.PathPattern, filter.EntryPath, filter.ExitPath = "", "", "", ""
	filterArgs, filterQuery := filter.query(analyzer.minIsBot > 0)
	pathQuery := strings.Repeat("?,", len(paths))

	for _, path := range paths {
		filterArgs = append(filterArgs, path)
	}

	var query string

	if analyzer.minIsBot > 0 {
		query = fmt.Sprintf(`SELECT path,
			uniq(visitor_id) visitors,
			uniq(visitor_id, session_id) sessions,
			count(1) views
			FROM page_view v
			INNER JOIN (
				SELECT visitor_id,
				session_id
				FROM session
				WHERE %s
				GROUP BY visitor_id, session_id
				HAVING sum(sign) > 0
			) s
			ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id
			WHERE path IN (%s)
			GROUP BY path
			ORDER BY visitors DESC, sessions DESC
			%s`, filterQuery, pathQuery[:len(pathQuery)-1], filter.withLimit())
	} else {
		query = fmt.Sprintf(`SELECT path,
			uniq(visitor_id) visitors,
			uniq(visitor_id, session_id) sessions,
			count(1) views
			FROM page_view
			WHERE %s
			AND path IN (%s)
			GROUP BY path
			ORDER BY visitors DESC, sessions DESC
			%s`, filterQuery, pathQuery[:len(pathQuery)-1], filter.withLimit())
	}

	stats, err := analyzer.store.SelectTotalVisitorSessionStats(query, filterArgs...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (analyzer *Analyzer) totalSessionDuration(filter *Filter) (int, error) {
	filterArgs, filterQuery := filter.query(true)
	innerFilterArgs, innerFilterQuery := filter.queryTime(false)
	args := make([]any, 0, len(innerFilterArgs)+len(filterArgs))
	var query strings.Builder
	query.WriteString(`SELECT sum(duration_seconds)
		FROM (
			SELECT sum(duration_seconds*sign) duration_seconds
			FROM session s `)

	if filter.Path != "" || filter.PathPattern != "" {
		args = append(args, innerFilterArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			path
			FROM page_view
			WHERE %s
		) v
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerFilterQuery))
	}

	args = append(args, filterArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
			GROUP BY visitor_id, session_id
			HAVING sum(sign) > 0
		)`, filterQuery))
	averageTimeSpentSeconds, err := analyzer.store.Count(query.String(), args...)

	if err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) totalEventDuration(filter *Filter) (int, error) {
	filterArgs, filterQuery := filter.query(false)
	var query string

	if analyzer.minIsBot > 0 {
		innerFilterArgs, innerFilterQuery := filter.queryTime(true)
		query = fmt.Sprintf(`SELECT sum(duration_seconds)
			FROM event e
			INNER JOIN (
				SELECT visitor_id,
				session_id
				FROM session
				WHERE %s
			) s
			ON s.visitor_id = e.visitor_id AND s.session_id = e.session_id
			WHERE %s`, innerFilterQuery, filterQuery)
		innerFilterArgs = append(innerFilterArgs, filterArgs...)
		filterArgs = innerFilterArgs
	} else {
		query = fmt.Sprintf(`SELECT sum(duration_seconds) FROM event WHERE %s`, filterQuery)
	}

	averageTimeSpentSeconds, err := analyzer.store.Count(query, filterArgs...)

	if err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) totalTimeOnPage(filter *Filter) (int, error) {
	timeArgs, timeQuery := filter.queryTime(false)
	fieldArgs, fieldQuery := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]any, 0, len(timeArgs)*2+len(fieldArgs))
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT sum(time_on_page) average_time_spent_seconds
		FROM (
			SELECT %s time_on_page
			FROM (
				SELECT session_id %s,
				sum(duration_seconds) duration_seconds
				FROM page_view v `, analyzer.timeOnPageQuery(filter), fieldsQuery))

	if analyzer.minIsBot > 0 || filter.EntryPath != "" || filter.ExitPath != "" {
		innerTimeArgs, innerTimeQuery := filter.queryTime(true)
		args = append(args, innerTimeArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id, entry_path, exit_path
			HAVING sum(sign) > 0
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerTimeQuery))
	}

	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
				GROUP BY visitor_id, session_id, time %s
				ORDER BY visitor_id, session_id, time
			)
			WHERE session_id = neighbor(session_id, 1, null)
			AND time_on_page > 0
			%s
		)`, timeQuery, fieldsQuery, fieldQuery))
	averageTimeSpentSeconds, err := analyzer.store.Count(query.String(), args...)

	if err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) avgTimeOnPage(filter *Filter, paths []string) ([]AvgTimeSpentStats, error) {
	if len(paths) == 0 {
		return []AvgTimeSpentStats{}, nil
	}

	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []AvgTimeSpentStats{}, nil
	}

	filter.Search, filter.Sort, filter.Offset, filter.Limit = nil, nil, 0, 0
	timeArgs, timeQuery := filter.queryTime(false)
	fieldArgs, fieldQuery := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]any, 0, len(timeArgs)*2+len(fieldArgs))
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT path,
		ifNull(toUInt64(avg(nullIf(time_on_page, 0))), 0) average_time_spent_seconds
		FROM (
			SELECT path,
			%s time_on_page
			FROM (
				SELECT session_id,
				path,
				duration_seconds
				%s
				FROM page_view v `, analyzer.timeOnPageQuery(filter), fieldsQuery))

	if analyzer.minIsBot > 0 || filter.EntryPath != "" || filter.ExitPath != "" {
		innerTimeArgs, innerTimeQuery := filter.queryTime(false)
		args = append(args, innerTimeArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id, entry_path, exit_path
			HAVING sum(sign) > 0
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerTimeQuery))
	}

	args = append(args, timeArgs...)
	pathQuery := strings.Repeat("?,", len(paths))

	for _, path := range paths {
		args = append(args, path)
	}

	args = append(args, fieldArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
				ORDER BY visitor_id, session_id, time
			)
			WHERE time_on_page > 0
			AND session_id = neighbor(session_id, 1, null)
			AND path IN (%s)
			%s
		)
		GROUP BY path`, timeQuery, pathQuery[:len(pathQuery)-1], fieldQuery))
	stats, err := analyzer.store.SelectAvgTimeSpentStats(query.String(), args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (analyzer *Analyzer) timeOnPageQuery(filter *Filter) string {
	timeOnPage := "neighbor(duration_seconds, 1, 0)"

	if filter.MaxTimeOnPageSeconds > 0 {
		timeOnPage = fmt.Sprintf("least(neighbor(duration_seconds, 1, 0), %d)", filter.MaxTimeOnPageSeconds)
	}

	return timeOnPage
}

func (analyzer *Analyzer) selectByAttribute(filter *Filter, attr ...Field) ([]any, string) {
	fields := make([]Field, 0, len(attr)+2)
	fields = append(fields, attr...)
	fields = append(fields, FieldVisitors, FieldRelativeVisitors)
	orderBy := make([]Field, 0, len(attr)+1)
	orderBy = append(orderBy, FieldVisitors)
	orderBy = append(orderBy, attr...)
	return analyzer.getFilter(filter).buildQuery(fields, attr, orderBy)
}

func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		filter = NewFilter(NullClient)
	}

	filter.validate()

	if analyzer.minIsBot > 0 {
		filter.minIsBot = analyzer.minIsBot
	}

	filterCopy := *filter
	return &filterCopy
}

func calculateGrowth[T int | float64](current, previous T) float64 {
	if current == 0 && previous == 0 {
		return 0
	} else if previous == 0 {
		return 1
	}

	c := float64(current)
	p := float64(previous)
	return (c - p) / p
}
