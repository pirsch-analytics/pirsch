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
	filter.Start = time.Now().UTC().Add(-duration)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()
	timeArgs = append(timeArgs, fieldArgs...)

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, "path")
	title, orderByTitle := "", ""

	if filter.IncludeTitle {
		filter.addFieldIfRequired(&fields, "title")
		title = ",title"
		orderByTitle = ",title ASC"
	}

	fieldsQuery := strings.Join(fields, ",")
	query := fmt.Sprintf(`SELECT path %s,
		sum(visitors) visitors
		FROM (
			SELECT path %s,
			count(DISTINCT fingerprint) visitors
			FROM (
				SELECT %s,
				fingerprint,
				argMax(path, time) exit_path
				FROM hit
				WHERE %s
				GROUP BY fingerprint, session_id, %s
			)
			%s
			GROUP BY path %s
		)
		GROUP BY path %s
		ORDER BY visitors DESC, path ASC %s
		%s`, title, title, fieldsQuery, timeQuery, fieldsQuery, fieldQuery, title, title, orderByTitle, filter.withLimit())
	var stats []ActiveVisitorStats

	if err := analyzer.store.Select(&stats, query, timeArgs...); err != nil {
		return nil, 0, err
	}

	query = fmt.Sprintf(`SELECT count(DISTINCT fingerprint) visitors
		FROM (
			SELECT %s,
			fingerprint,
			argMax(path, time) exit_path
			FROM hit
			WHERE %s
			GROUP BY fingerprint, session_id, %s
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
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	withFillArgs, withFillQuery := filter.withFill()
	timeArgs = append(timeArgs, fieldArgs...)
	timeArgs = append(timeArgs, withFillArgs...)
	timezone := filter.Timezone.String()
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	query := fmt.Sprintf(`SELECT day,
		count(DISTINCT fingerprint) visitors,
		count(DISTINCT(fingerprint, session_id)) sessions,
		sum(views) views,
		countIf(is_bounce) bounces,
		bounces / IF(sessions = 0, 1, sessions) bounce_rate
		FROM (
			SELECT toDate(time, '%s') day %s,
			fingerprint,
			session_id,
			argMax(page_views, time) views,
			argMax(is_bounce, time) is_bounce,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id, day %s
		)
		%s
		GROUP BY day
		ORDER BY day ASC %s, visitors DESC`, timezone, fieldsQuery, filter.table(), timeQuery, fieldsQuery, fieldQuery, withFillQuery)
	var stats []VisitorStats

	if err := analyzer.store.Select(&stats, query, timeArgs...); err != nil {
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

	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	timeArgs = append(timeArgs, fieldArgs...)
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	query := fmt.Sprintf(`SELECT count(DISTINCT fingerprint) visitors,
		count(DISTINCT(fingerprint, session_id)) sessions,
		sum(views) views,
		countIf(is_bounce) / IF(sessions = 0, 1, sessions) bounce_rate
		FROM (
			SELECT fingerprint %s,
			session_id,
			argMax(page_views, time) views,
			argMax(is_bounce, time) is_bounce,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id %s
		)
		%s`, fieldsQuery, filter.table(), timeQuery, fieldsQuery, fieldQuery)
	current := new(growthStats)

	if err := analyzer.store.Get(current, query, timeArgs...); err != nil {
		return nil, err
	}

	var currentTimeSpent int
	var err error

	if filter.Path == "" {
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

	timeArgs, _ = filter.queryTime()
	fieldArgs, _, _ = filter.queryFields()
	timeArgs = append(timeArgs, fieldArgs...)
	previous := new(growthStats)

	if err := analyzer.store.Get(previous, query, timeArgs...); err != nil {
		return nil, err
	}

	var previousTimeSpent int

	if filter.Path == "" {
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
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	timeArgs = append(timeArgs, fieldArgs...)
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	query := fmt.Sprintf(`SELECT toHour(time, '%s') hour,
		sum(visitors) visitors
		FROM (
			SELECT time %s,
			count(DISTINCT fingerprint) visitors,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id, time %s
		)
		%s
		GROUP BY hour
		ORDER BY hour WITH FILL FROM 0 TO 24`, filter.Timezone.String(), fieldsQuery, filter.table(), timeQuery, fieldsQuery, fieldQuery)
	var stats []VisitorHourStats

	if err := analyzer.store.Select(&stats, query, timeArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Pages returns the visitor count, session count, bounce rate, views, and average time on page grouped by path and (optional) page title.
func (analyzer *Analyzer) Pages(filter *Filter) ([]PageStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, _ := filter.queryFields()
	var fields []string

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
		filter.Path = ""
		_, _, fields = filter.queryFields()
	}

	args := make([]interface{}, 0, len(timeArgs)*3+len(fieldArgs))
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	title, titleGroupBy, titleOrderBy := "", "", ""

	if filter.IncludeTitle {
		title = "title,"
		titleGroupBy = ",title"
		titleOrderBy = "title ASC,"
		filter.addFieldIfRequired(&fields, "title")
	}

	top, topQuery := "", ""

	if filter.IncludeAvgTimeOnPage {
		top = ",avg(average_time_spent_seconds) average_time_spent_seconds"
		var topArgs []interface{}
		topQuery, topArgs = analyzer.avgTimeOnPages(filter, false)
		args = append(args, topArgs...)
	}

	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT arrayJoin(paths) path,
		%s
		count(DISTINCT fingerprint) visitors,
		count(DISTINCT(fingerprint, session_id)) sessions,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM %s
			WHERE %s
		), 1) relative_visitors,
		count(1) views,
		views / greatest((
			SELECT sum(views) FROM (
				SELECT argMax(page_views, time) views
				FROM %s
				WHERE %s
				GROUP BY fingerprint, session_id
			)
		), 1) relative_views,
		countIf(is_bounce) bounces,
		bounces / IF(sessions = 0, 1, sessions) bounce_rate
		%s
		FROM (
			SELECT groupArray(path) paths %s,
			argMax(is_bounce, time) is_bounce,
			argMax(path, time) exit_path,
			fingerprint,
			session_id
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id %s
		) AS entries
		%s
		%s
		GROUP BY path %s
		ORDER BY visitors DESC, %s path ASC
		%s`, title,
		table, timeQuery,
		table, timeQuery,
		top, fieldsQuery, table, timeQuery, fieldsQuery, topQuery, fieldQuery,
		titleGroupBy, titleOrderBy, filter.withLimit())
	var stats []PageStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// EntryPages returns the visitor count and time on page grouped by path and (optional) page title for the first page visited.
func (analyzer *Analyzer) EntryPages(filter *Filter) ([]EntryStats, error) {
	filter = analyzer.getFilter(filter)
	filter.Path = ""
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, "entry_path")
	title, titleInner, titleOuter, titleOrderBy, titleJoin := "", "", "", "", ""

	if filter.IncludeTitle {
		title = "title,"
		titleInner = ",title"
		titleOuter = "entries.title title,"
		titleOrderBy = "title ASC,"
		titleJoin = "AND visitors.title = entries.title"
		filter.addFieldIfRequired(&fields, "title")
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	args = append(args, timeArgs...)
	top, topQuery := "", ""

	if filter.IncludeAvgTimeOnPage {
		top = ",avg(average_time_spent_seconds) average_time_spent_seconds"
		var topArgs []interface{}
		topQuery, topArgs = analyzer.avgTimeOnPages(filter, true)
		args = append(args, topArgs...)
	}

	fieldsQuery := strings.Join(fields, ",")
	query := fmt.Sprintf(`SELECT *
		FROM (
			SELECT entries.path path,
			%s
			sum(visitors) visitors,
			sum(entries) entries,
			entries/IF(visitors = 0, 1, visitors) entry_rate
			%s
			FROM (
				SELECT entry_path path,
				%s
				sum(entries) entries
				FROM (
					SELECT %s,
					argMax(path, time) exit_path,
					count(DISTINCT fingerprint) entries
					FROM %s
					WHERE %s
					GROUP BY fingerprint, session_id, %s
				)
				%s
				GROUP BY path %s
			) AS entries
			INNER JOIN (
				SELECT path,
				%s
				count(DISTINCT fingerprint) visitors
				FROM %s
				WHERE %s
				GROUP BY path %s
			) AS visitors
			ON visitors.path = entries.path %s
			%s
			GROUP BY path %s
		)
		WHERE entries > 0
		ORDER BY entries DESC, %s path ASC
		%s`, titleOuter, top, title,
		fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery, titleInner,
		title, table, timeQuery, titleInner, titleJoin,
		topQuery, titleInner, titleOrderBy, filter.withLimit())
	var stats []EntryStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// ExitPages returns the visitor count and time on page grouped by path and (optional) page title for the last page visited.
func (analyzer *Analyzer) ExitPages(filter *Filter) ([]ExitStats, error) {
	filter = analyzer.getFilter(filter)
	filter.Path = ""
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	title, titleInner, titleInnerArgMax, titleOrderBy := "", "", "", ""

	if filter.IncludeTitle {
		title = "title,"
		titleInner = ",title"
		titleInnerArgMax = ",argMax(title, time) title"
		titleOrderBy = "title ASC,"
	}

	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs)*2)
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT exit_path path,
		%s
		sum(visitors) visitors,
		sum(exits) exits,
		exits/IF(visitors = 0, 1, visitors) exit_rate
		FROM (
			SELECT path exit_path %s %s,
			count(DISTINCT fingerprint) visitors
			FROM %s
			WHERE %s
			GROUP BY exit_path %s %s
		) AS visitors
		INNER JOIN (
			SELECT exit_path %s %s,
			count(1) exits
			FROM (
				SELECT argMax(path, time) exit_path %s %s
				FROM %s
				WHERE %s
				GROUP BY fingerprint, session_id %s
			)
			%s
			GROUP BY exit_path %s %s
		) AS exits
		USING exit_path %s
		%s
		GROUP by  %s exit_path
		ORDER BY exits DESC, %s exit_path ASC
		%s`, title,
		fieldsQuery, titleInner, table, timeQuery, fieldsQuery, titleInner,
		fieldsQuery, titleInner, fieldsQuery, titleInnerArgMax, table, timeQuery, fieldsQuery, fieldQuery, fieldsQuery, titleInner,
		fieldsQuery, fieldQuery, title, titleOrderBy, filter.withLimit())
	var stats []ExitStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// PageConversions returns the visitor count, views, and conversion rate for conversion goals.
// This function is supposed to be used with the Filter.PathPattern, to list page conversions.
func (analyzer *Analyzer) PageConversions(filter *Filter) (*PageConversionsStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT sum(visitors) visitors,
		sum(views) views,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM hit
			WHERE %s
		), 1) cr
		FROM (
			SELECT count(DISTINCT fingerprint) visitors %s,
			argMax(page_views, time) views,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id %s
		)
		%s
		ORDER BY visitors DESC`, timeQuery, fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery)
	stats := new(PageConversionsStats)

	if err := analyzer.store.Get(stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Events returns the visitor count, views, and conversion rate for custom events.
func (analyzer *Analyzer) Events(filter *Filter) ([]EventStats, error) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, "event_name")
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT event_name,
		sum(visitors) visitors,
		sum(views) views,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM hit
			WHERE %s
		), 1) cr,
		toUInt64(avg(avg_duration)) average_duration_seconds,
		groupUniqArrayArray(meta_keys) meta_keys
		FROM (
			SELECT event_name %s,
			groupUniqArrayArray(event_meta_keys) meta_keys,
			count(DISTINCT fingerprint) visitors,
			argMax(page_views, time) views,
			avg(event_duration_seconds) avg_duration,
			argMax(path, time) exit_path
			FROM event
			WHERE %s
			GROUP BY fingerprint, session_id, event_name %s
		)
		%s
		GROUP BY event_name
		ORDER BY visitors DESC, event_name
		%s`, timeQuery, fieldsQuery, timeQuery, fieldsQuery, fieldQuery, filter.withLimit())
	var stats []EventStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
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

	filterArgs, filterQuery, _ := filter.query()
	filter.EventName = ""
	crFilterArgs, crFilterQuery, _ := filter.query()
	query := fmt.Sprintf(`SELECT event_name,
		sum(visitors) visitors,
		sum(views) views,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM hit
			WHERE %s
		), 1) cr,
		toUInt64(avg(avg_duration)) average_duration_seconds,
		meta_value
		FROM (
			SELECT event_name,
			count(DISTINCT fingerprint) visitors,
			argMax(page_views, time) views,
			avg(event_duration_seconds) avg_duration,
			event_meta_values[indexOf(event_meta_keys, ?)] meta_value
			FROM event
			WHERE %s
			AND has(event_meta_keys, ?)
			GROUP BY fingerprint, session_id, event_name, meta_value
		)
		GROUP BY event_name, meta_value
		ORDER BY visitors DESC, meta_value
		%s`, crFilterQuery, filterQuery, filter.withLimit())
	args := make([]interface{}, 0, len(filterArgs)*2)
	args = append(args, crFilterArgs...)
	args = append(args, filter.EventMetaKey)
	args = append(args, filterArgs...)
	args = append(args, filter.EventMetaKey)
	var stats []EventStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// Referrer returns the visitor count and bounce rate grouped by referrer.
func (analyzer *Analyzer) Referrer(filter *Filter) ([]ReferrerStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	selectReferrer, groupReferrer := "", ""

	if filter.Referrer != "" || filter.ReferrerName != "" {
		selectReferrer = ",referrer"
		groupReferrer = ",referrer"
	} else {
		selectReferrer = ",any(referrer) referrer"
	}

	filter.addFieldIfRequired(&fields, "referrer_name")
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT referrer_name,
		any(referrer_icon) referrer_icon %s,
		count(DISTINCT fingerprint) visitors,
		count(DISTINCT(fingerprint, session_id)) sessions,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM %s
			WHERE %s
		), 1) relative_visitors,
		countIf(is_bounce) bounces,
		bounces / IF(sessions = 0, 1, sessions) bounce_rate
		FROM (
			SELECT fingerprint %s %s,
			any(referrer_icon) referrer_icon,
			session_id,
			argMax(is_bounce, time) is_bounce,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id %s %s
		)
		%s
		GROUP BY referrer_name %s
		ORDER BY visitors DESC, referrer_name ASC
		%s`, selectReferrer, table, timeQuery,
		selectReferrer, fieldsQuery, table, timeQuery, groupReferrer, fieldsQuery, fieldQuery,
		groupReferrer, filter.withLimit())
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
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "AND " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, "desktop")
	filter.addFieldIfRequired(&fields, "mobile")
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*3+len(fieldArgs)*3)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT (
			SELECT count(DISTINCT fingerprint)
			FROM (
				SELECT fingerprint %s,
				argMax(path, time) exit_path
				FROM %s
				WHERE %s
				GROUP BY fingerprint, session_id %s
			)
			WHERE desktop = 1
			AND mobile = 0
			%s
		) AS "platform_desktop",
		(
			SELECT count(DISTINCT fingerprint)
			FROM (
				SELECT fingerprint %s,
				argMax(path, time) exit_path
				FROM %s
				WHERE %s
				GROUP BY fingerprint, session_id %s
			)
			WHERE desktop = 0
			AND mobile = 1
			%s
		) AS "platform_mobile",
		(
			SELECT count(DISTINCT fingerprint)
			FROM (
				SELECT fingerprint %s,
				argMax(path, time) exit_path
				FROM %s
				WHERE %s
				GROUP BY fingerprint, session_id %s
			)
			WHERE desktop = 0
			AND mobile = 0
			%s
		) AS "platform_unknown",
		"platform_desktop" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_desktop,
		"platform_mobile" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_mobile,
		"platform_unknown" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_unknown`,
		fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery,
		fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery,
		fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery)
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
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, "os")
	filter.addFieldIfRequired(&fields, "os_version")
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT os,
		os_version,
		sum(visitors) visitors,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM hit
			WHERE %s
		), 1) relative_visitors
		FROM (
			SELECT count(DISTINCT fingerprint) visitors %s,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id %s
		)
		%s
		GROUP BY os, os_version
		ORDER BY visitors DESC, os, os_version
		%s`, timeQuery, fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery, filter.withLimit())
	var stats []OSVersionStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// BrowserVersion returns the visitor count grouped by browser and version.
func (analyzer *Analyzer) BrowserVersion(filter *Filter) ([]BrowserVersionStats, error) {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, "browser")
	filter.addFieldIfRequired(&fields, "browser_version")
	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]interface{}, 0, len(timeArgs)*2+len(fieldArgs))
	args = append(args, timeArgs...)
	args = append(args, timeArgs...)
	args = append(args, fieldArgs...)
	query := fmt.Sprintf(`SELECT browser,
		browser_version,
		sum(visitors) visitors,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM hit
			WHERE %s
		), 1) relative_visitors
		FROM (
			SELECT count(DISTINCT fingerprint) visitors %s,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id %s
		)
		%s
		GROUP BY browser, browser_version
		ORDER BY visitors DESC, browser, browser_version
		%s`, timeQuery, fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery, filter.withLimit())
	var stats []BrowserVersionStats

	if err := analyzer.store.Select(&stats, query, args...); err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgSessionDuration returns the average session duration grouped by day.
func (analyzer *Analyzer) AvgSessionDuration(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	withFillArgs, withFillQuery := filter.withFill()
	timeArgs = append(timeArgs, fieldArgs...)
	timeArgs = append(timeArgs, withFillArgs...)
	query := fmt.Sprintf(`SELECT day,
		toUInt64(avg(duration)) average_time_spent_seconds
		FROM (
			SELECT toDate(time, '%s') day %s,
			sum(duration_seconds) duration,
			argMax(path, time) exit_path
			FROM hit
			WHERE %s
			GROUP BY fingerprint, session_id, day %s
		)
		WHERE duration != 0 %s
		GROUP BY day
		ORDER BY day %s`, filter.Timezone.String(), fieldsQuery, timeQuery, fieldsQuery, fieldQuery, withFillQuery)
	var stats []TimeSpentStats

	if err := analyzer.store.Select(&stats, query, timeArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgTimeOnPage returns the average time on page grouped by day.
func (analyzer *Analyzer) AvgTimeOnPage(filter *Filter) ([]TimeSpentStats, error) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	withFillArgs, withFillQuery := filter.withFill()
	query := fmt.Sprintf(`SELECT day,
		toUInt64(avg(time_on_page)) average_time_spent_seconds
		FROM (
			SELECT toDate(time, '%s') day,
			%s time_on_page
			FROM (
				SELECT session_id,
				time,
				duration_seconds
				%s
				FROM hit
				WHERE %s
				ORDER BY fingerprint, session_id, time
			)
			WHERE time_on_page > 0
			AND session_id = neighbor(session_id, 1, null)
			%s
		)
		GROUP BY day
		ORDER BY day %s`, filter.Timezone.String(), analyzer.timeOnPageQuery(filter), fieldsQuery, timeQuery, fieldQuery, withFillQuery)
	timeArgs = append(timeArgs, fieldArgs...)
	timeArgs = append(timeArgs, withFillArgs...)
	var stats []TimeSpentStats

	if err := analyzer.store.Select(&stats, query, timeArgs...); err != nil {
		return nil, err
	}

	return stats, nil
}

func (analyzer *Analyzer) totalSessionDuration(filter *Filter) (int, error) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	timeArgs = append(timeArgs, fieldArgs...)
	query := fmt.Sprintf(`SELECT sum(duration_seconds)
		FROM (
			SELECT sum(duration_seconds) duration_seconds %s,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id %s
		)
		%s`, fieldsQuery, filter.table(), timeQuery, fieldsQuery, fieldQuery)
	var averageTimeSpentSeconds int

	if err := analyzer.store.Get(&averageTimeSpentSeconds, query, timeArgs...); err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil
}

func (analyzer *Analyzer) totalTimeOnPage(filter *Filter) (int, error) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := strings.Join(fields, ",")

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
				GROUP BY fingerprint, session_id, time %s
				ORDER BY fingerprint, session_id, time
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

func (analyzer *Analyzer) avgTimeOnPages(filter *Filter, useJoinOn bool) (string, []interface{}) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "WHERE " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, "path")
	title, joinTitle := "", ""

	if filter.IncludeTitle {
		title = ",title"
		joinTitle = "AND entries.title = top.title"
		filter.addFieldIfRequired(&fields, "title")
	}

	fieldsQuery := strings.Join(fields, ",")

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	join := ""

	if useJoinOn {
		join = fmt.Sprintf(`ON entries.path = top.path %s`, joinTitle)
	} else {
		join = fmt.Sprintf(`USING path %s`, title)
	}

	query := fmt.Sprintf(`LEFT JOIN (
			SELECT path %s,
			toUInt64(avg(time_on_page)) average_time_spent_seconds
			FROM (
				SELECT *,
				%s time_on_page
				FROM (
					SELECT session_id,
					time,
					sum(duration_seconds) duration_seconds,
					argMax(path, time) exit_path
					%s
					FROM hit
					WHERE %s
					GROUP BY fingerprint, session_id, time %s
					ORDER BY fingerprint, session_id, time
				)
				WHERE time_on_page > 0
				AND session_id = neighbor(session_id, 1, null)
			)
			%s
			GROUP BY path %s
		) AS top
		%s`, title,
		analyzer.timeOnPageQuery(filter), fieldsQuery, timeQuery, fieldsQuery, fieldQuery,
		title, join)
	timeArgs = append(timeArgs, fieldArgs...)
	return query, timeArgs
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

func (analyzer *Analyzer) timeOnPageQuery(filter *Filter) string {
	timeOnPage := "neighbor(duration_seconds, 1, 0)"

	if filter.MaxTimeOnPageSeconds > 0 {
		timeOnPage = fmt.Sprintf("least(neighbor(duration_seconds, 1, 0), %d)", filter.MaxTimeOnPageSeconds)
	}

	return timeOnPage
}

func (analyzer *Analyzer) selectByAttribute(results interface{}, filter *Filter, attr string) error {
	filter = analyzer.getFilter(filter)
	table := filter.table()
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery, fields := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	filter.addFieldIfRequired(&fields, attr)
	fieldsQuery := strings.Join(fields, ",")
	timeArgs = append(timeArgs, timeArgs...)
	timeArgs = append(timeArgs, fieldArgs...)
	query := fmt.Sprintf(`SELECT "%s",
		count(DISTINCT fingerprint) visitors,
		visitors / greatest((
			SELECT count(DISTINCT fingerprint)
			FROM %s
			WHERE %s
		), 1) relative_visitors
		FROM (
			SELECT %s,
			fingerprint,
			argMax(path, time) exit_path
			FROM %s
			WHERE %s
			GROUP BY fingerprint, session_id, %s
		)
		%s
		GROUP BY "%s"
		ORDER BY visitors DESC, "%s" ASC
		%s`, attr, table, timeQuery,
		fieldsQuery, table, timeQuery, fieldsQuery, fieldQuery,
		attr, attr, filter.withLimit())
	return analyzer.store.Select(results, query, timeArgs...)
}

func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		filter = NewFilter(NullClient)
	}

	filter.validate()
	return filter
}
