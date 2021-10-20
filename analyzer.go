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
	BounceRate float64 `db:"bounce_rate"`
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
	title := filter.groupByTitle()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	timeArgs = append(timeArgs, fieldArgs...)
	fieldsQuery := filter.fields()

	if filter.Path == "" {
		if fieldsQuery == "" {
			fieldsQuery = "path"
		} else {
			fieldsQuery += ",path"
		}
	}

	if filter.IncludeTitle {
		fieldsQuery += ",title"
	}

	query := fmt.Sprintf(`SELECT path %s,
		sum(visitors) visitors
		FROM (
			SELECT path %s,
			count(DISTINCT visitor_id) visitors
			FROM (
				SELECT %s,
				visitor_id,
				argMax(path, time) exit_path
				FROM session
				WHERE %s
				GROUP BY visitor_id, session_id, %s
			)
			%s
			GROUP BY path %s
		)
		GROUP BY path %s
		ORDER BY visitors DESC %s, path ASC
		%s`, title, title, fieldsQuery, timeQuery, fieldsQuery, fieldQuery, title, title, title, filter.withLimit())
	var stats []ActiveVisitorStats

	if err := analyzer.store.Select(&stats, query, timeArgs...); err != nil {
		return nil, 0, err
	}

	query = fmt.Sprintf(`SELECT count(DISTINCT visitor_id) visitors
		FROM (
			SELECT %s,
			visitor_id,
			argMax(path, time) exit_path
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id, %s
		)
		%s`, fieldsQuery, timeQuery, fieldsQuery, fieldQuery)
	count, err := analyzer.store.Count(query, timeArgs...)

	if err != nil {
		return nil, 0, err
	}

	return stats, count, nil
}

// Visitors returns the visitor count, session count, bounce rate, views, and average session duration grouped by day.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]VisitorStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	filterArgs, filterQuery := filter.query()
	withFillArgs, withFillQuery := filter.withFill()
	filterArgs = append(filterArgs, withFillArgs...)
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT toDate(time, '%s') day,
		count(DISTINCT visitor_id) visitors,
		count(DISTINCT visitor_id, session_id) sessions,
		count(1) views `, filter.Timezone.String()))

	if table == "session" {
		query.WriteString(`, sum(sign) bounces, bounces / IF(sessions = 0, 1, sessions) bounce_rate `)
	}

	query.WriteString(fmt.Sprintf(`FROM %s
		WHERE %s
		GROUP BY day
		ORDER BY day ASC %s, visitors DESC`, table, filterQuery, withFillQuery))
	var stats []VisitorStats

	if err := analyzer.store.Select(&stats, query.String(), filterArgs...); err != nil {
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

	table := filter.table()
	filterArgs, filterQuery := filter.query()
	var query strings.Builder
	query.WriteString(`SELECT count(DISTINCT visitor_id) visitors,
		count(DISTINCT(visitor_id, session_id)) sessions,
		count(1) views `)

	if table == "session" {
		query.WriteString(`, sum(sign) / IF(sessions = 0, 1, sessions) bounce_rate `)
	}

	query.WriteString(fmt.Sprintf(`FROM %s WHERE %s`, table, filterQuery))
	current := new(growthStats)

	if err := analyzer.store.Get(current, query.String(), filterArgs...); err != nil {
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

	filterArgs, _ = filter.query()
	previous := new(growthStats)

	if err := analyzer.store.Get(previous, query.String(), filterArgs...); err != nil {
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
	// we cannot read from materialized views here, as they only represent the last point in time
	filter = analyzer.getFilter(filter)
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	timeArgs = append(timeArgs, fieldArgs...)
	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT hour,
		sum(visitors) visitors
		FROM (
			SELECT toHour(time, '%s') hour %s,
			count(DISTINCT visitor_id) visitors `, filter.Timezone.String(), fieldsQuery))

	if table == "session" {
		query.WriteString(`,argMax(path, time) exit_path `)
	}

	query.WriteString(fmt.Sprintf(`FROM %s
			WHERE %s
			GROUP BY visitor_id, session_id, hour %s
		)
		%s
		GROUP BY hour
		ORDER BY hour WITH FILL FROM 0 TO 24`, table, timeQuery, fieldsQuery, fieldQuery))
	var stats []VisitorHourStats

	if err := analyzer.store.Select(&stats, query.String(), timeArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Pages returns the visitor count, session count, bounce rate, views, and average time on page grouped by path and (optional) page title.
func (analyzer *Analyzer) Pages(filter *Filter) ([]PageStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	title := filter.groupByTitle()
	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, innerFilterArgs...)
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT path %s,
		count(DISTINCT visitor_id) visitors,
		count(DISTINCT visitor_id, session_id) sessions,
		visitors / greatest((
			SELECT count(DISTINCT visitor_id)
			FROM %s
			WHERE %s
		), 1) relative_visitors,
		count(1) views, `, title, table, innerFilterQuery))

	if table == "sessions" {
		query.WriteString(fmt.Sprintf(`views / greatest((
				SELECT sum(page_views*sign) views
				FROM session
				WHERE %s
				GROUP BY visitor_id, session_id
			), 1) relative_views,
			countIf(is_bounce) bounces,
			bounces / IF(sessions = 0, 1, sessions) bounce_rate, `, innerFilterQuery))
	} else {
		query.WriteString(fmt.Sprintf(`views / greatest((
				SELECT count(1) views
				FROM event
				WHERE %s
			), 1) relative_views, `, innerFilterQuery))
	}

	query.WriteString(fmt.Sprintf(`ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0) average_time_spent_seconds
		FROM %s
		WHERE %s
		GROUP BY path %s
		ORDER BY visitors DESC %s, path
		%s`, table, outerFilterQuery, title, title, filter.withLimit()))
	var stats []PageStats

	if err := analyzer.store.Select(&stats, query.String(), innerFilterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// EntryPages returns the visitor count and time on page grouped by path and (optional) page title for the first page visited.
func (analyzer *Analyzer) EntryPages(filter *Filter) ([]EntryStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []EntryStats{}, nil
	}

	title := filter.groupByTitle()
	joinTitle := ""

	if title != "" {
		joinTitle = "AND sessions.title = v.title"
	}

	outerFilterArgs, outerFilterQuery := filter.query()

	if filter.ExitPath != "" {
		filter.Path = filter.ExitPath
		filter.ExitPath = ""
	}

	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	query := fmt.Sprintf(`SELECT entry_path %s,
		count(DISTINCT visitor_id, session_id) entries,
		any(v.visitors) visitors,
		any(v.sessions) sessions,
		entries/sessions entry_rate,
		ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0) average_time_spent_seconds
		FROM sessions
		INNER JOIN (
			SELECT path %s,
			count(DISTINCT visitor_id) visitors,
			count(DISTINCT visitor_id, session_id) sessions
			FROM sessions
			WHERE %s
			GROUP BY path %s
		) AS v
		ON entry_path = v.path %s
		WHERE %s
		GROUP BY entry_path %s
		ORDER BY entries DESC %s, entry_path
		%s`, title, title, innerFilterQuery, title, joinTitle, outerFilterQuery, title, title, filter.withLimit())
	var stats []EntryStats

	if err := analyzer.store.Select(&stats, query, innerFilterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// ExitPages returns the visitor count and time on page grouped by path and (optional) page title for the last page visited.
func (analyzer *Analyzer) ExitPages(filter *Filter) ([]ExitStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []ExitStats{}, nil
	}

	title := filter.groupByTitle()
	joinTitle := ""

	if title != "" {
		joinTitle = "AND sessions.title = v.title"
	}

	outerFilterArgs, outerFilterQuery := filter.query()

	if filter.ExitPath != "" {
		filter.Path = filter.ExitPath
		filter.ExitPath = ""
	}

	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	query := fmt.Sprintf(`SELECT exit_path %s,
		count(DISTINCT visitor_id, session_id) exits,
		any(v.visitors) visitors,
		any(v.sessions) sessions,
		exits/sessions exit_rate
		FROM sessions
		INNER JOIN (
			SELECT path %s,
			count(DISTINCT visitor_id) visitors,
			count(DISTINCT visitor_id, session_id) sessions
			FROM sessions
			WHERE %s
			GROUP BY path %s
		) AS v
		ON exit_path = v.path %s
		WHERE %s
		GROUP BY exit_path %s
		ORDER BY exits DESC %s, exit_path
		%s`, title, title, innerFilterQuery, title, joinTitle, outerFilterQuery, title, title, filter.withLimit())
	var stats []ExitStats

	if err := analyzer.store.Select(&stats, query, innerFilterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// PageConversions returns the visitor count, views, and conversion rate for conversion goals.
// This function is supposed to be used with the Filter.PathPattern, to list page conversions.
func (analyzer *Analyzer) PageConversions(filter *Filter) (*PageConversionsStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	var query strings.Builder
	query.WriteString(`SELECT count(DISTINCT visitor_id) visitors, `)

	if table == "sessions" {
		query.WriteString(`sum(page_views) views, `)
	} else {
		query.WriteString(`count(1) views, `)
	}

	query.WriteString(fmt.Sprintf(`visitors / greatest((
			SELECT count(DISTINCT visitor_id)
			FROM sessions
			WHERE %s
		), 1) cr
		FROM %s
		WHERE %s
		ORDER BY visitors DESC
		%s`, innerFilterQuery, table, outerFilterQuery, filter.withLimit()))
	stats := new(PageConversionsStats)

	if err := analyzer.store.Get(stats, query.String(), innerFilterArgs...); err != nil {
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
			SELECT count(DISTINCT visitor_id)
			FROM sessions
			WHERE %s
		), 1) cr,
		ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0) average_duration_seconds,
		groupUniqArrayArray(event_meta_keys) meta_keys
		FROM events
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
			SELECT count(DISTINCT visitor_id)
			FROM sessions
			WHERE %s
		), 1) cr,
		ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0) average_duration_seconds,
		event_meta_values[indexOf(event_meta_keys, ?)] meta_value
		FROM events
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
	table := filter.table()
	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	ref := ""
	groupSortRef := ""

	if filter.Referrer != "" || filter.ReferrerName != "" {
		ref = "referrer ref,"
		groupSortRef = ",ref"
	} else {
		ref = "any(referrer) ref,"
	}

	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT %s referrer_name,
		any(referrer_icon) referrer_icon,
		count(DISTINCT visitor_id) visitors,
		count(DISTINCT(visitor_id, session_id)) sessions,
		visitors / greatest((
			SELECT count(DISTINCT visitor_id)
			FROM sessions
			WHERE %s
		), 1) relative_visitors `, ref, innerFilterQuery))

	if table == "sessions" {
		query.WriteString(`, countIf(is_bounce) bounces, bounces / IF(sessions = 0, 1, sessions) bounce_rate `)
	}

	query.WriteString(fmt.Sprintf(`FROM %s
		WHERE %s
		GROUP BY referrer_name %s
		ORDER BY visitors DESC, referrer_name %s
		%s`, table, outerFilterQuery, groupSortRef, groupSortRef, filter.withLimit()))
	var stats []ReferrerStats

	if err := analyzer.store.Select(&stats, query.String(), innerFilterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Platform returns the visitor count grouped by platform.
func (analyzer *Analyzer) Platform(filter *Filter) (*PlatformStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	filterArgs, filterQuery := filter.query()
	args := make([]interface{}, 0, len(filterArgs)*3)
	args = append(args, filterArgs...)
	args = append(args, filterArgs...)
	args = append(args, filterArgs...)
	query := fmt.Sprintf(`SELECT (
			SELECT count(DISTINCT visitor_id)
			FROM %s
			WHERE %s
			AND desktop = 1
			AND mobile = 0
		) AS "platform_desktop",
		(
			SELECT count(DISTINCT visitor_id)
			FROM %s
			WHERE %s
			AND desktop = 0
			AND mobile = 1
		) AS "platform_mobile",
		(
			SELECT count(DISTINCT visitor_id)
			FROM %s
			WHERE %s
			AND desktop = 0
			AND mobile = 0
		) AS "platform_unknown",
		"platform_desktop" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_desktop,
		"platform_mobile" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_mobile,
		"platform_unknown" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_unknown`,
		table, filterQuery, table, filterQuery, table, filterQuery)
	stats := new(PlatformStats)

	if err := analyzer.store.Get(stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Languages returns the visitor count grouped by language.
func (analyzer *Analyzer) Languages(filter *Filter) ([]LanguageStats, error) {
	var stats []LanguageStats

	if err := analyzer.selectByAttribute(&stats, filter, "language"); err != nil {
		return nil, err
	}

	return stats, nil
}

// Countries returns the visitor count grouped by country.
func (analyzer *Analyzer) Countries(filter *Filter) ([]CountryStats, error) {
	var stats []CountryStats

	if err := analyzer.selectByAttribute(&stats, filter, "country_code"); err != nil {
		return nil, err
	}

	return stats, nil
}

// Cities returns the visitor count grouped by city.
func (analyzer *Analyzer) Cities(filter *Filter) ([]CityStats, error) {
	var stats []CityStats

	if err := analyzer.selectByAttribute(&stats, filter, "city"); err != nil {
		return nil, err
	}

	return stats, nil
}

// Browser returns the visitor count grouped by browser.
func (analyzer *Analyzer) Browser(filter *Filter) ([]BrowserStats, error) {
	var stats []BrowserStats

	if err := analyzer.selectByAttribute(&stats, filter, "browser"); err != nil {
		return nil, err
	}

	return stats, nil
}

// OS returns the visitor count grouped by operating system.
func (analyzer *Analyzer) OS(filter *Filter) ([]OSStats, error) {
	var stats []OSStats

	if err := analyzer.selectByAttribute(&stats, filter, "os"); err != nil {
		return nil, err
	}

	return stats, nil
}

// ScreenClass returns the visitor count grouped by screen class.
func (analyzer *Analyzer) ScreenClass(filter *Filter) ([]ScreenClassStats, error) {
	var stats []ScreenClassStats

	if err := analyzer.selectByAttribute(&stats, filter, "screen_class"); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMSource returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMSource(filter *Filter) ([]UTMSourceStats, error) {
	var stats []UTMSourceStats

	if err := analyzer.selectByAttribute(&stats, filter, "utm_source"); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMMedium returns the visitor count grouped by utm medium.
func (analyzer *Analyzer) UTMMedium(filter *Filter) ([]UTMMediumStats, error) {
	var stats []UTMMediumStats

	if err := analyzer.selectByAttribute(&stats, filter, "utm_medium"); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMCampaign returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMCampaign(filter *Filter) ([]UTMCampaignStats, error) {
	var stats []UTMCampaignStats

	if err := analyzer.selectByAttribute(&stats, filter, "utm_campaign"); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMContent returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMContent(filter *Filter) ([]UTMContentStats, error) {
	var stats []UTMContentStats

	if err := analyzer.selectByAttribute(&stats, filter, "utm_content"); err != nil {
		return nil, err
	}

	return stats, nil
}

// UTMTerm returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMTerm(filter *Filter) ([]UTMTermStats, error) {
	var stats []UTMTermStats

	if err := analyzer.selectByAttribute(&stats, filter, "utm_term"); err != nil {
		return nil, err
	}

	return stats, nil
}

// OSVersion returns the visitor count grouped by operating systems and version.
func (analyzer *Analyzer) OSVersion(filter *Filter) ([]OSVersionStats, error) {
	filter = analyzer.getFilter(filter)
	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	query := fmt.Sprintf(`SELECT os,
		os_version,
		count(DISTINCT visitor_id) visitors,
		visitors / greatest((
			SELECT count(DISTINCT visitor_id)
			FROM sessions
			WHERE %s
		), 1) relative_visitors
		FROM %s
		WHERE %s
		GROUP BY os, os_version
		ORDER BY visitors DESC, os, os_version
		%s`, innerFilterQuery, filter.table(), outerFilterQuery, filter.withLimit())
	var stats []OSVersionStats

	if err := analyzer.store.Select(&stats, query, innerFilterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// BrowserVersion returns the visitor count grouped by browser and version.
func (analyzer *Analyzer) BrowserVersion(filter *Filter) ([]BrowserVersionStats, error) {
	filter = analyzer.getFilter(filter)
	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	query := fmt.Sprintf(`SELECT browser,
		browser_version,
		count(DISTINCT visitor_id) visitors,
		visitors / greatest((
			SELECT count(DISTINCT visitor_id)
			FROM sessions
			WHERE %s
		), 1) relative_visitors
		FROM %s
		WHERE %s
		GROUP BY browser, browser_version
		ORDER BY visitors DESC, browser, browser_version
		%s`, innerFilterQuery, filter.table(), outerFilterQuery, filter.withLimit())
	var stats []BrowserVersionStats

	if err := analyzer.store.Select(&stats, query, innerFilterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgSessionDuration returns the average session duration grouped by day.
func (analyzer *Analyzer) AvgSessionDuration(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "events" {
		return []TimeSpentStats{}, nil
	}

	filterArgs, filterQuery := filter.query()
	withFillArgs, withFillQuery := filter.withFill()
	filterArgs = append(filterArgs, withFillArgs...)
	query := fmt.Sprintf(`SELECT toDate(time) day,
		ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0) average_time_spent_seconds
		FROM session
		WHERE %s
		AND duration_seconds != 0
		GROUP BY day
		ORDER BY day
		%s`, filterQuery, withFillQuery)
	var stats []TimeSpentStats

	if err := analyzer.store.Select(&stats, query, filterArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgTimeOnPage returns the average time on page grouped by day.
func (analyzer *Analyzer) AvgTimeOnPage(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.table() == "events" {
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
	query := fmt.Sprintf(`SELECT day,
		toUInt64(avg(time_on_page)) average_time_spent_seconds
		FROM (
			SELECT day,
			%s time_on_page
			FROM (
				SELECT session_id,
				toDate(time, '%s') day,
				duration_seconds,
				argMax(path, time) exit_path
				%s
				FROM hit
				WHERE %s
				GROUP BY visitor_id, session_id, time, duration_seconds %s
				ORDER BY visitor_id, session_id, time
			)
			WHERE time_on_page > 0
			AND session_id = neighbor(session_id, 1, null)
			%s
		)
		GROUP BY day
		ORDER BY day
		%s`, analyzer.timeOnPageQuery(filter), filter.Timezone.String(),
		fieldsQuery, timeQuery, fieldsQuery,
		fieldQuery, withFillQuery)
	timeArgs = append(timeArgs, fieldArgs...)
	timeArgs = append(timeArgs, withFillArgs...)
	var stats []TimeSpentStats

	if err := analyzer.store.Select(&stats, query, timeArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

func (analyzer *Analyzer) totalSessionDuration(filter *Filter) (int, error) {
	filterArgs, filterQuery := filter.query()
	query := fmt.Sprintf(`SELECT sum(duration_seconds)
		FROM (
			SELECT max(duration_seconds) duration_seconds
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id
		)`, filterQuery)
	var averageTimeSpentSeconds int

	if err := analyzer.store.Get(&averageTimeSpentSeconds, query, filterArgs...); err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) totalEventDuration(filter *Filter) (int, error) {
	filterArgs, filterQuery := filter.query()
	query := fmt.Sprintf(`SELECT sum(duration_seconds) FROM events WHERE %s`, filterQuery)
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

	query := fmt.Sprintf(`SELECT sum(time_on_page) average_time_spent_seconds
		FROM (
			SELECT %s time_on_page
			FROM (
				SELECT session_id %s,
				sum(duration_seconds) duration_seconds,
				argMax(path, time) exit_path
				FROM %s
				WHERE %s
				GROUP BY visitor_id, session_id, time %s
				ORDER BY visitor_id, session_id, time
			)
			WHERE time_on_page > 0
			AND session_id = neighbor(session_id, 1, null)
			%s
		)`, analyzer.timeOnPageQuery(filter), fieldsQuery, filter.table(), timeQuery, fieldsQuery, fieldQuery)
	timeArgs = append(timeArgs, fieldArgs...)
	stats := new(struct {
		AverageTimeSpentSeconds int `db:"average_time_spent_seconds" json:"average_time_spent_seconds"`
	})

	if err := analyzer.store.Get(stats, query, timeArgs...); err != nil {
		return 0, err
	}

	return stats.AverageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) timeOnPageQuery(filter *Filter) string {
	timeOnPage := "neighbor(duration_seconds, 1, 0)"

	if filter.MaxTimeOnPageSeconds > 0 {
		timeOnPage = fmt.Sprintf("least(neighbor(duration_seconds, 1, 0), %d)", filter.MaxTimeOnPageSeconds)
	}

	return timeOnPage
}

func (analyzer *Analyzer) selectByAttribute(results interface{}, filter *Filter, attr string) error {
	filter = analyzer.getFilter(filter)
	outerFilterArgs, outerFilterQuery := filter.query()
	innerFilterArgs, innerFilterQuery := filter.queryTime()
	innerFilterArgs = append(innerFilterArgs, outerFilterArgs...)
	query := fmt.Sprintf(`SELECT "%s",
		count(DISTINCT visitor_id) visitors,
		visitors / greatest((
			SELECT count(DISTINCT visitor_id)
			FROM sessions
			WHERE %s
		), 1) relative_visitors
		FROM %s
		WHERE %s
		GROUP BY "%s"
		ORDER BY visitors DESC, "%s" ASC
		%s`, attr, innerFilterQuery, filter.table(), outerFilterQuery, attr, attr, filter.withLimit())
	return analyzer.store.Select(results, query, innerFilterArgs...)
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
