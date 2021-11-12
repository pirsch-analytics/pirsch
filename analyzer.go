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

type growthStats struct {
	Visitors   int
	Views      int
	Sessions   int
	Bounces    int
	BounceRate float64 `db:"bounce_rate"`
}

type totalVisitorSessionStats struct {
	Path     string
	Visitors int
	Views    int
	Sessions int
}

type avgTimeSpentStats struct {
	Path                    string
	AverageTimeSpentSeconds int `db:"average_time_spent_seconds"`
}

// Analyzer provides an interface to analyze statistics.
type Analyzer struct {
	store Store
}

// NewAnalyzer returns a new Analyzer for given Store.
func NewAnalyzer(store Store) *Analyzer {
	return &Analyzer{
		store,
	}
}

// ActiveVisitors returns the active visitors per path and (optional) page title and the total number of active visitors for given duration.
// Use time.Minute*5 for example to get the active visitors for the past 5 minutes.
func (analyzer *Analyzer) ActiveVisitors(filter *Filter, duration time.Duration) ([]ActiveVisitorStats, int, error) {
	filter = analyzer.getFilter(filter)
	filter.Start = time.Now().In(filter.Timezone).Add(-duration)
	title := ""

	if filter.IncludeTitle {
		title = ",title"
	}

	filterArgs, filterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	args := make([]interface{}, 0, len(innerFilterArgs)+len(filterArgs))
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT path %s,
		uniq(visitor_id) visitors
		FROM page_view v `, title))

	if filter.EntryPath != "" || filter.ExitPath != "" {
		args = append(args, innerFilterArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerFilterQuery))
	}

	args = append(args, filterArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
		GROUP BY path %s
		ORDER BY visitors DESC, path
		%s`, filterQuery, title, filter.withLimit()))
	var stats []ActiveVisitorStats

	if err := analyzer.store.Select(&stats, query.String(), args...); err != nil {
		return nil, 0, err
	}

	query.Reset()
	query.WriteString(`SELECT uniq(visitor_id) visitors
		FROM page_view v `)

	if filter.EntryPath != "" || filter.ExitPath != "" {
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
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
	args, query := buildQuery(analyzer.getFilter(filter), []field{
		fieldVisitors,
		fieldSessions,
		fieldViews,
		fieldBounces,
		fieldBounceRate,
	}, nil, nil)
	stats := new(TotalVisitorStats)

	if err := analyzer.store.Get(stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Visitors returns the visitor count, session count, bounce rate, and views grouped by day.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]VisitorStats, error) {
	args, query := buildQuery(analyzer.getFilter(filter), []field{
		fieldDay,
		fieldVisitors,
		fieldSessions,
		fieldViews,
		fieldBounces,
		fieldBounceRate,
	}, []field{
		fieldDay,
	}, []field{
		fieldDay,
		fieldVisitors,
	})
	var stats []VisitorStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Growth returns the growth rate for visitor count, session count, bounces, views, and average session duration or average time on page (if path is set).
// The growth rate is relative to the previous time range or day.
// The period or day for the filter must be set, else an error is returned.
func (analyzer *Analyzer) Growth(filter *Filter) (*Growth, error) {
	filter = analyzer.getFilter(filter)

	if filter.Day.IsZero() && (filter.From.IsZero() || filter.To.IsZero()) {
		return nil, ErrNoPeriodOrDay
	}

	fields := []field{
		fieldVisitors,
		fieldSessions,
		fieldViews,
		fieldBounces,
		fieldBounceRate,
	}
	args, query := buildQuery(filter, fields, nil, nil)
	current := new(growthStats)

	if err := analyzer.store.Get(current, query, args...); err != nil {
		return nil, err
	}

	var currentTimeSpent int
	var err error

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

	if filter.Day.IsZero() {
		days := filter.To.Sub(filter.From)
		filter.To = filter.From.Add(-time.Hour * 24)
		filter.From = filter.To.Add(-days)
	} else {
		filter.Day = filter.Day.Add(-time.Hour * 24)
	}

	args, query = buildQuery(filter, fields, nil, nil)
	previous := new(growthStats)

	if err := analyzer.store.Get(previous, query, args...); err != nil {
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
		VisitorsGrowth:  analyzer.calculateGrowth(current.Visitors, previous.Visitors),
		ViewsGrowth:     analyzer.calculateGrowth(current.Views, previous.Views),
		SessionsGrowth:  analyzer.calculateGrowth(current.Sessions, previous.Sessions),
		BouncesGrowth:   analyzer.calculateGrowthFloat64(current.BounceRate, previous.BounceRate),
		TimeSpentGrowth: analyzer.calculateGrowth(currentTimeSpent, previousTimeSpent),
	}, nil
}

// VisitorHours returns the visitor count grouped by time of day.
func (analyzer *Analyzer) VisitorHours(filter *Filter) ([]VisitorHourStats, error) {
	args, query := buildQuery(analyzer.getFilter(filter), []field{
		fieldHour,
		fieldVisitors,
	}, []field{
		fieldHour,
	}, []field{
		fieldHour,
	})
	var stats []VisitorHourStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Pages returns the visitor count, session count, bounce rate, views, and average time on page grouped by path and (optional) page title.
func (analyzer *Analyzer) Pages(filter *Filter) ([]PageStats, error) {
	filter = analyzer.getFilter(filter)
	fields := []field{
		fieldPath,
		fieldVisitors,
		fieldSessions,
		fieldRelativeVisitors,
		fieldViews,
		fieldRelativeViews,
		fieldBounces,
		fieldBounceRate,
	}
	groupBy := []field{
		fieldPath,
	}
	orderBy := []field{
		fieldVisitors,
		fieldPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, fieldTitle)
		groupBy = append(groupBy, fieldTitle)
		orderBy = append(orderBy, fieldTitle)
	}

	if filter.table() == "event" {
		fields = append(fields, fieldEventTimeSpent)
	}

	args, query := buildQuery(filter, fields, groupBy, orderBy)
	var stats []PageStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	if filter.IncludeTimeOnPage && filter.table() == "session" {
		paths := make([]string, 0, len(stats))

		for i := range stats {
			paths = append(paths, stats[i].Path)
		}

		top, err := analyzer.avgTimeOnPage(filter, paths)

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

	if filter.table() == "event" {
		return []EntryStats{}, nil
	}

	fields := []field{
		fieldEntryPath,
		fieldEntries,
	}
	groupBy := []field{
		fieldEntryPath,
	}
	orderBy := []field{
		fieldEntries,
		fieldEntryPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, fieldEntryTitle)
		groupBy = append(groupBy, fieldEntryTitle)
		orderBy = append(orderBy, fieldEntryTitle)
	}

	args, query := buildQuery(filter, fields, groupBy, orderBy)
	var stats []EntryStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(stats))

	for i := range stats {
		paths = append(paths, stats[i].Path)
	}

	total, err := analyzer.totalVisitorsSessions(filter, paths)

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

	if filter.IncludeTimeOnPage {
		top, err := analyzer.avgTimeOnPage(filter, paths)

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

	if filter.table() == "event" {
		return []ExitStats{}, nil
	}

	fields := []field{
		fieldExitPath,
		fieldExits,
	}
	groupBy := []field{
		fieldExitPath,
	}
	orderBy := []field{
		fieldExits,
		fieldExitPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, fieldExitTitle)
		groupBy = append(groupBy, fieldExitTitle)
		orderBy = append(orderBy, fieldExitTitle)
	}

	args, query := buildQuery(filter, fields, groupBy, orderBy)
	var stats []ExitStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(stats))

	for i := range stats {
		paths = append(paths, stats[i].Path)
	}

	total, err := analyzer.totalVisitorsSessions(filter, paths)

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

	return stats, nil
}

// PageConversions returns the visitor count, views, and conversion rate for conversion goals.
// This function is supposed to be used with the Filter.PathPattern, to list page conversions.
func (analyzer *Analyzer) PageConversions(filter *Filter) (*PageConversionsStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.PathPattern == "" {
		return nil, nil
	}

	args, query := buildQuery(filter, []field{
		fieldVisitors,
		fieldViews,
		fieldCR,
	}, nil, []field{
		fieldVisitors,
	})
	stats := new(PageConversionsStats)

	if err := analyzer.store.Get(stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Events returns the visitor count, views, and conversion rate for custom events.
func (analyzer *Analyzer) Events(filter *Filter) ([]EventStats, error) {
	filter = analyzer.getFilter(filter)
	filter.eventFilter = true
	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	query := fmt.Sprintf(`SELECT event_name,
		count(DISTINCT visitor_id) visitors,
		count(1) views,
		visitors / greatest((
			SELECT uniq(visitor_id)
			FROM session
			WHERE %s
		), 1) cr,
		ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0) average_duration_seconds,
		groupUniqArrayArray(event_meta_keys) meta_keys
		FROM event
		WHERE %s
		GROUP BY event_name
		ORDER BY visitors DESC, event_name
		%s`, innerFilterQuery, outerFilterQuery, filter.withLimit())
	var stats []EventStats

	if err := analyzer.store.Select(&stats, query, innerFilterArgs...); err != nil {
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

	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, filter.EventMetaKey)
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	innerFilterArgs = append(innerFilterArgs, filter.EventMetaKey)
	query := fmt.Sprintf(`SELECT event_name,
		count(DISTINCT visitor_id) visitors,
		count(1) views,
		visitors / greatest((
			SELECT uniq(visitor_id)
			FROM session
			WHERE %s
		), 1) cr,
		ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0) average_duration_seconds,
		event_meta_values[indexOf(event_meta_keys, ?)] meta_value
		FROM event
		WHERE %s
		AND has(event_meta_keys, ?)
		GROUP BY event_name, meta_value
		ORDER BY visitors DESC, meta_value
		%s`, innerFilterQuery, outerFilterQuery, filter.withLimit())
	var stats []EventStats

	if err := analyzer.store.Select(&stats, query, innerFilterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Referrer returns the visitor count and bounce rate grouped by referrer.
func (analyzer *Analyzer) Referrer(filter *Filter) ([]ReferrerStats, error) {
	filter = analyzer.getFilter(filter)
	fields := []field{
		fieldReferrerName,
		fieldReferrerIcon,
		fieldVisitors,
		fieldSessions,
		fieldRelativeVisitors,
		fieldBounces,
		fieldBounceRate,
	}
	groupBy := []field{
		fieldReferrerName,
	}
	orderBy := []field{
		fieldVisitors,
		fieldReferrerName,
	}

	if filter.Referrer != "" || filter.ReferrerName != "" {
		fields = append(fields, fieldReferrer)
		groupBy = append(groupBy, fieldReferrer)
		orderBy = append(orderBy, fieldReferrer)
	} else {
		fields = append(fields, fieldAnyReferrer)
	}

	args, query := buildQuery(filter, fields, groupBy, orderBy)
	var stats []ReferrerStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Platform returns the visitor count grouped by platform.
func (analyzer *Analyzer) Platform(filter *Filter) (*PlatformStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	filterArgs, filterQuery := filter.query()
	args := make([]interface{}, 0, len(filterArgs)*4)
	var query strings.Builder

	if table == "session" {
		query.WriteString(`SELECT sum(desktop*sign) platform_desktop,
			sum(mobile*sign) platform_mobile,
			sum(sign)-platform_desktop-platform_mobile platform_unknown, `)
	} else {
		args = append(args, filterArgs...)
		args = append(args, filterArgs...)
		args = append(args, filterArgs...)
		query.WriteString(fmt.Sprintf(`SELECT (
				SELECT uniq(visitor_id)
				FROM event
				WHERE %s
				AND desktop = 1
				AND mobile = 0
			) platform_desktop,
			(
				SELECT uniq(visitor_id)
				FROM event
				WHERE %s
				AND desktop = 0
				AND mobile = 1
			) platform_mobile,
			(
				SELECT uniq(visitor_id)
				FROM event
				WHERE %s
				AND desktop = 0
				AND mobile = 0
			) platform_unknown, `, filterQuery, filterQuery, filterQuery))
	}

	query.WriteString(`"platform_desktop" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_desktop,
		"platform_mobile" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_mobile,
		"platform_unknown" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_unknown `)

	if table == "session" {
		query.WriteString(`FROM session s `)

		if filter.Path != "" || filter.PathPattern != "" {
			entryPath, exitPath, eventName := filter.EntryPath, filter.ExitPath, filter.EventName
			filter.EntryPath, filter.ExitPath, filter.EventName = "", "", ""
			innerFilterArgs, innerFilterQuery := filter.query()
			filter.EntryPath, filter.ExitPath, filter.EventName = entryPath, exitPath, eventName
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
		query.WriteString(fmt.Sprintf(`WHERE %s`, filterQuery))
	}

	stats := new(PlatformStats)

	if err := analyzer.store.Get(stats, query.String(), args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Languages returns the visitor count grouped by language.
func (analyzer *Analyzer) Languages(filter *Filter) ([]LanguageStats, error) {
	var stats []LanguageStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldLanguage); err != nil {
		return nil, err
	}

	return stats, nil
}

// Countries returns the visitor count grouped by country.
func (analyzer *Analyzer) Countries(filter *Filter) ([]CountryStats, error) {
	var stats []CountryStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldCountry); err != nil {
		return nil, err
	}

	return stats, nil
}

// Cities returns the visitor count grouped by city.
func (analyzer *Analyzer) Cities(filter *Filter) ([]CityStats, error) {
	var stats []CityStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldCity); err != nil {
		return nil, err
	}

	return stats, nil
}

// Browser returns the visitor count grouped by browser.
func (analyzer *Analyzer) Browser(filter *Filter) ([]BrowserStats, error) {
	var stats []BrowserStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldBrowser); err != nil {
		return nil, err
	}

	return stats, nil
}

// OS returns the visitor count grouped by operating system.
func (analyzer *Analyzer) OS(filter *Filter) ([]OSStats, error) {
	var stats []OSStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldOS); err != nil {
		return nil, err
	}

	return stats, nil
}

// ScreenClass returns the visitor count grouped by screen class.
func (analyzer *Analyzer) ScreenClass(filter *Filter) ([]ScreenClassStats, error) {
	var stats []ScreenClassStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldScreenClass); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMSource returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMSource(filter *Filter) ([]UTMSourceStats, error) {
	var stats []UTMSourceStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldUTMSource); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMMedium returns the visitor count grouped by utm medium.
func (analyzer *Analyzer) UTMMedium(filter *Filter) ([]UTMMediumStats, error) {
	var stats []UTMMediumStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldUTMMedium); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMCampaign returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMCampaign(filter *Filter) ([]UTMCampaignStats, error) {
	var stats []UTMCampaignStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldUTMCampaign); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMContent returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMContent(filter *Filter) ([]UTMContentStats, error) {
	var stats []UTMContentStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldUTMContent); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMTerm returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMTerm(filter *Filter) ([]UTMTermStats, error) {
	var stats []UTMTermStats

	if err := analyzer.selectByAttribute(&stats, filter, fieldUTMTerm); err != nil {
		return nil, err
	}

	return stats, nil
}

// OSVersion returns the visitor count grouped by operating systems and version.
func (analyzer *Analyzer) OSVersion(filter *Filter) ([]OSVersionStats, error) {
	args, query := buildQuery(analyzer.getFilter(filter), []field{
		fieldOS,
		fieldOSVersion,
		fieldVisitors,
		fieldRelativeVisitors,
	}, []field{
		fieldOS,
		fieldOSVersion,
	}, []field{
		fieldVisitors,
		fieldOS,
		fieldOSVersion,
	})
	var stats []OSVersionStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// BrowserVersion returns the visitor count grouped by browser and version.
func (analyzer *Analyzer) BrowserVersion(filter *Filter) ([]BrowserVersionStats, error) {
	args, query := buildQuery(analyzer.getFilter(filter), []field{
		fieldBrowser,
		fieldBrowserVersion,
		fieldVisitors,
		fieldRelativeVisitors,
	}, []field{
		fieldBrowser,
		fieldBrowserVersion,
	}, []field{
		fieldVisitors,
		fieldBrowser,
		fieldBrowserVersion,
	})
	var stats []BrowserVersionStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgSessionDuration returns the average session duration grouped by day.
func (analyzer *Analyzer) AvgSessionDuration(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []TimeSpentStats{}, nil
	}

	filterArgs, filterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	withFillArgs, withFillQuery := filter.withFill()
	args := make([]interface{}, 0, len(filterArgs)+len(innerFilterArgs)+len(withFillArgs))
	var query strings.Builder
	query.WriteString(`SELECT toDate(time) day,
		ifNull(toUInt64(avg(nullIf(duration_seconds, 0)*sign)), 0) average_time_spent_seconds
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
	args = append(args, withFillArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
		AND duration_seconds != 0
		GROUP BY day
		ORDER BY day
		%s`, filterQuery, withFillQuery))
	var stats []TimeSpentStats

	if err := analyzer.store.Select(&stats, query.String(), args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgTimeOnPage returns the average time on page grouped by day.
func (analyzer *Analyzer) AvgTimeOnPage(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []TimeSpentStats{}, nil
	}

	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	withFillArgs, withFillQuery := filter.withFill()
	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs)+len(withFillArgs))
	var query strings.Builder
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
				FROM page_view v `, analyzer.timeOnPageQuery(filter), filter.Timezone.String(), fieldsQuery))

	if filter.EntryPath != "" || filter.ExitPath != "" {
		args = append(args, timeArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, timeQuery))
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
	var stats []TimeSpentStats

	if err := analyzer.store.Select(&stats, query.String(), args...); err != nil {
		return nil, err
	}

	return stats, nil
}

func (analyzer *Analyzer) totalVisitorsSessions(filter *Filter, paths []string) ([]totalVisitorSessionStats, error) {
	if len(paths) == 0 {
		return []totalVisitorSessionStats{}, nil
	}

	filter = analyzer.getFilter(filter)
	filter.Path, filter.EntryPath, filter.ExitPath = "", "", ""
	filterArgs, filterQuery := filter.query()
	pathQuery := strings.Repeat("?,", len(paths))

	for _, path := range paths {
		filterArgs = append(filterArgs, path)
	}

	query := fmt.Sprintf(`SELECT path,
		uniq(visitor_id) visitors,
		uniq(visitor_id, session_id) sessions,
		count(1) views
		FROM page_view
		WHERE %s
		AND path IN (%s)
		GROUP BY path
		ORDER BY visitors DESC, sessions DESC
		%s`, filterQuery, pathQuery[:len(pathQuery)-1], filter.withLimit())
	var stats []totalVisitorSessionStats

	if err := analyzer.store.Select(&stats, query, filterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

func (analyzer *Analyzer) totalSessionDuration(filter *Filter) (int, error) {
	filterArgs, filterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	args := make([]interface{}, 0, len(innerFilterArgs)+len(filterArgs))
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
		)`, filterQuery))
	var averageTimeSpentSeconds int

	if err := analyzer.store.Get(&averageTimeSpentSeconds, query.String(), args...); err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) totalEventDuration(filter *Filter) (int, error) {
	filterArgs, filterQuery := filter.query()
	query := fmt.Sprintf(`SELECT sum(duration_seconds) FROM event WHERE %s`, filterQuery)
	var averageTimeSpentSeconds int

	if err := analyzer.store.Get(&averageTimeSpentSeconds, query, filterArgs...); err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) totalTimeOnPage(filter *Filter) (int, error) {
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT sum(time_on_page) average_time_spent_seconds
		FROM (
			SELECT %s time_on_page
			FROM (
				SELECT session_id %s,
				sum(duration_seconds) duration_seconds
				FROM page_view v `, analyzer.timeOnPageQuery(filter), fieldsQuery))

	if filter.EntryPath != "" || filter.ExitPath != "" {
		args = append(args, timeArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, timeQuery))
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
	stats := new(struct {
		AverageTimeSpentSeconds int `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
	})

	if err := analyzer.store.Get(stats, query.String(), args...); err != nil {
		return 0, err
	}

	return stats.AverageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) avgTimeOnPage(filter *Filter, paths []string) ([]avgTimeSpentStats, error) {
	if len(paths) == 0 {
		return []avgTimeSpentStats{}, nil
	}

	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []avgTimeSpentStats{}, nil
	}

	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
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

	if filter.EntryPath != "" || filter.ExitPath != "" {
		args = append(args, timeArgs...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			entry_path,
			exit_path
			FROM session
			WHERE %s
		) s
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, timeQuery))
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
	var stats []avgTimeSpentStats

	if err := analyzer.store.Select(&stats, query.String(), args...); err != nil {
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

func (analyzer *Analyzer) selectByAttribute(results interface{}, filter *Filter, attr field) error {
	args, query := buildQuery(analyzer.getFilter(filter), []field{
		attr,
		fieldVisitors,
		fieldRelativeVisitors,
	}, []field{
		attr,
	}, []field{
		fieldVisitors,
		attr,
	})
	return analyzer.store.Select(results, query, args...)
}

func (analyzer *Analyzer) calculateGrowth(current, previous int) float64 {
	if current == 0 && previous == 0 {
		return 0
	} else if previous == 0 {
		return 1
	}

	c := float64(current)
	p := float64(previous)
	return (c - p) / p
}

func (analyzer *Analyzer) calculateGrowthFloat64(current, previous float64) float64 {
	if current == 0 && previous == 0 {
		return 0
	} else if previous == 0 {
		return 1
	}

	c := current
	p := previous
	return (c - p) / p
}

func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		filter = NewFilter(NullClient)
	}

	filter.validate()
	filterCopy := *filter
	return &filterCopy
}
