package pirsch

import (
	"errors"
	"fmt"
	"time"
)

const (
	byAttributeQuery = `SELECT "%s", count(DISTINCT fingerprint) visitors, visitors / (
			SELECT sum(s) FROM (
				SELECT count(DISTINCT fingerprint) s
				FROM hit
				WHERE %s
				GROUP BY "%s"
			)
		) relative_visitors
		FROM hit
		WHERE %s
		GROUP BY "%s"
		ORDER BY visitors DESC, "%s" ASC
		%s`
)

var (
	// ErrNoPeriodOrDay is returned in case no period or day was specified to calculate the growth rate.
	ErrNoPeriodOrDay = errors.New("no period or day specified")
)

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

// ActiveVisitors returns the active visitors per path and the total number of active visitors for given duration.
// Use time.Minute*5 for example to get the active visitors for the past 5 minutes.
func (analyzer *Analyzer) ActiveVisitors(filter *Filter, duration time.Duration) ([]Stats, int, error) {
	filter = analyzer.getFilter(filter)
	filter.Start = time.Now().UTC().Add(-duration)
	args, filterQuery := filter.query()
	query := fmt.Sprintf(`SELECT "path", count(DISTINCT fingerprint) visitors
		FROM hit
		WHERE %s
		GROUP BY path
		ORDER BY visitors DESC, path ASC`, filterQuery)
	visitors, err := analyzer.store.Select(query, args...)

	if err != nil {
		return nil, 0, err
	}

	query = fmt.Sprintf(`SELECT count(DISTINCT fingerprint) visitors FROM hit WHERE %s`, filterQuery)
	count, err := analyzer.store.Count(query, args...)

	if err != nil {
		return nil, 0, err
	}

	return visitors, count, nil
}

// Visitors returns the visitor count, session count, bounce rate, views, and average session duration grouped by day.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	args, filterQuery := filter.query()
	withFillArgs, withFillQuery := filter.withFill()
	args = append(args, withFillArgs...)
	query := fmt.Sprintf(`SELECT day,
		sum(visitors) visitors,
		sum(sessions) sessions,
		sum(views) views,
		countIf(bounce = 1) bounces,
		bounces / IF(visitors = 0, 1, visitors) bounce_rate
		FROM (
			SELECT toDate(time) day,
			count(DISTINCT fingerprint) visitors,
			count(DISTINCT(fingerprint, session)) sessions,
			count(*) views,
			length(groupArray(path)) = 1 bounce
			FROM hit
			WHERE %s
			GROUP BY toDate(time), fingerprint
		)
		GROUP BY day
		ORDER BY day ASC %s, visitors DESC`, filterQuery, withFillQuery)
	return analyzer.store.Select(query, args...)
}

// Growth returns the growth rate for visitor count, session count, bounces, views, and average session duration or average time on page (if path is set).
// The growth rate is relative to the previous time range or day.
// The period or day for the filter must be set, else an error is returned.
func (analyzer *Analyzer) Growth(filter *Filter) (*Growth, error) {
	filter = analyzer.getFilter(filter)

	if filter.Day.IsZero() && (filter.From.IsZero() || filter.To.IsZero()) {
		return nil, ErrNoPeriodOrDay
	}

	args, filterQuery := filter.query()
	query := fmt.Sprintf(`SELECT sum(visitors) visitors,
		sum(sessions) sessions,
		sum(views) views,
		countIf(bounce = 1) bounces
		FROM (
			SELECT count(DISTINCT fingerprint) visitors,
			count(DISTINCT(fingerprint, session)) sessions,
			count(*) views,
			length(groupArray(path)) = 1 bounce
			FROM hit
			WHERE %s
			GROUP BY toDate(time), fingerprint
		)`, filterQuery)
	current, err := analyzer.store.Get(query, args...)

	if err != nil {
		return nil, err
	}

	var currentTimeSpent int

	if filter.Path == "" {
		currentTimeSpent, err = analyzer.TotalSessionDuration(filter)
	} else {
		currentTimeSpent, err = analyzer.TotalTimeOnPage(filter)
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

	args, _ = filter.query()
	previous, err := analyzer.store.Get(query, args...)

	if err != nil {
		return nil, err
	}

	var previousTimeSpent int

	if filter.Path == "" {
		previousTimeSpent, err = analyzer.TotalSessionDuration(filter)
	} else {
		previousTimeSpent, err = analyzer.TotalTimeOnPage(filter)
	}

	if err != nil {
		return nil, err
	}

	return &Growth{
		VisitorsGrowth:  analyzer.calculateGrowth(current.Visitors, previous.Visitors),
		ViewsGrowth:     analyzer.calculateGrowth(current.Views, previous.Views),
		SessionsGrowth:  analyzer.calculateGrowth(current.Sessions, previous.Sessions),
		BouncesGrowth:   analyzer.calculateGrowth(current.Bounces, previous.Bounces),
		TimeSpentGrowth: analyzer.calculateGrowth(currentTimeSpent, previousTimeSpent),
	}, nil
}

// VisitorHours returns the visitor count grouped by time of day.
func (analyzer *Analyzer) VisitorHours(filter *Filter) ([]Stats, error) {
	args, filterQuery := analyzer.getFilter(filter).query()
	query := fmt.Sprintf(`SELECT toHour(time) hour, count(DISTINCT fingerprint) visitors
		FROM hit
		WHERE %s
		GROUP BY hour
		ORDER BY hour WITH FILL FROM 0 TO 24`, filterQuery)
	return analyzer.store.Select(query, args...)
}

// Pages returns the visitor count, session count, bounce rate, views, and average time on page grouped by path.
func (analyzer *Analyzer) Pages(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	filterArgs, filterQuery := filter.query()
	query := fmt.Sprintf(`SELECT path,
		count(DISTINCT fingerprint) visitors,
		visitors / (
			SELECT sum(s) FROM (
				SELECT count(DISTINCT fingerprint) s
				FROM hit
				WHERE %s
				GROUP BY path, fingerprint
			)
		) relative_visitors,
		sum(sessions) sessions,
		sum(views) views,
		views / (
			SELECT sum(s) FROM (
				SELECT count(*) s
				FROM hit
				WHERE %s
				GROUP BY path, fingerprint
			)
		) relative_views,
		countIf(bounce = 1) bounces,
		bounces / IF(visitors = 0, 1, visitors) bounce_rate
		FROM (
			SELECT path,
			fingerprint,
			count(DISTINCT(fingerprint, session)) sessions,
			count(*) views,
			length(groupArray(path)) = 1 bounce
			FROM hit
			WHERE %s
			GROUP BY path, fingerprint
		)
		GROUP BY path
		ORDER BY visitors DESC, path ASC
		%s`, filterQuery, filterQuery, filterQuery, filter.withLimit())
	args := make([]interface{}, 0, len(filterArgs)*3)
	args = append(args, filterArgs...)
	args = append(args, filterArgs...)
	args = append(args, filterArgs...)
	return analyzer.store.Select(query, args...)
}

// Referrer returns the visitor count and bounce rate grouped by referrer.
func (analyzer *Analyzer) Referrer(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	args, filterQuery := filter.query()
	query := fmt.Sprintf(`SELECT referrer,
		referrer_name,
		referrer_icon,
		count(DISTINCT fingerprint) visitors,
		visitors / (
			SELECT sum(s) FROM (
				SELECT count(DISTINCT fingerprint) s
				FROM hit
				WHERE %s
				GROUP BY fingerprint, referrer, referrer_name, referrer_icon
			)
		) relative_visitors,
		countIf(bounce = 1) bounces,
		bounces / IF(visitors = 0, 1, visitors) bounce_rate
		FROM (
			SELECT fingerprint,
			referrer,
			referrer_name,
			referrer_icon,
			length(groupArray(path)) = 1 bounce
			FROM hit
			WHERE %s
			GROUP BY fingerprint, referrer, referrer_name, referrer_icon
		)
		GROUP BY referrer, referrer_name, referrer_icon
		ORDER BY visitors DESC
		%s`, filterQuery, filterQuery, filter.withLimit())
	args = append(args, args...)
	return analyzer.store.Select(query, args...)
}

// Languages returns the visitor count grouped by language.
func (analyzer *Analyzer) Languages(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "language")
}

// Countries returns the visitor count grouped by country.
func (analyzer *Analyzer) Countries(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "country_code")
}

// Browser returns the visitor count grouped by browser.
func (analyzer *Analyzer) Browser(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "browser")
}

// OS returns the visitor count grouped by operating system.
func (analyzer *Analyzer) OS(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "os")
}

// Platform returns the visitor count grouped by platform.
func (analyzer *Analyzer) Platform(filter *Filter) (*Stats, error) {
	filterArgs, filterQuery := analyzer.getFilter(filter).query()
	query := fmt.Sprintf(`SELECT (
			SELECT count(DISTINCT fingerprint)
			FROM "hit"
			WHERE %s
			AND desktop = 1
			AND mobile = 0
		) AS "platform_desktop",
		(
			SELECT count(DISTINCT fingerprint)
			FROM "hit"
			WHERE %s
			AND desktop = 0
			AND mobile = 1
		) AS "platform_mobile",
		(
			SELECT count(DISTINCT fingerprint)
			FROM "hit"
			WHERE %s
			AND desktop = 0
			AND mobile = 0
		) AS "platform_unknown",
		"platform_desktop" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_desktop,
		"platform_mobile" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_mobile,
		"platform_unknown" / IF("platform_desktop" + "platform_mobile" + "platform_unknown" = 0, 1, "platform_desktop" + "platform_mobile" + "platform_unknown") AS relative_platform_unknown`, filterQuery, filterQuery, filterQuery)
	args := make([]interface{}, 0, len(filterArgs)*3)
	args = append(args, filterArgs...)
	args = append(args, filterArgs...)
	args = append(args, filterArgs...)
	return analyzer.store.Get(query, args...)
}

// ScreenClass returns the visitor count grouped by screen class.
func (analyzer *Analyzer) ScreenClass(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "screen_class")
}

// UTMSource returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMSource(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "utm_source")
}

// UTMMedium returns the visitor count grouped by utm medium.
func (analyzer *Analyzer) UTMMedium(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "utm_medium")
}

// UTMCampaign returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMCampaign(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "utm_campaign")
}

// UTMContent returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMContent(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "utm_content")
}

// UTMTerm returns the visitor count grouped by utm source.
func (analyzer *Analyzer) UTMTerm(filter *Filter) ([]Stats, error) {
	return analyzer.selectByAttribute(filter, "utm_term")
}

// AvgSessionDuration returns the average session duration grouped by day.
func (analyzer *Analyzer) AvgSessionDuration(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	args, filterQuery := filter.query()
	withFillArgs, withFillQuery := filter.withFill()
	args = append(args, withFillArgs...)
	query := fmt.Sprintf(`SELECT day, toUInt64(avg(duration)) average_time_spent_seconds
			FROM (
				SELECT toDate(time) day, max(time)-min(time) duration
				FROM hit
				WHERE %s
				AND session IS NOT NULL
				GROUP BY day, fingerprint, session
			)
		WHERE duration != 0
		GROUP BY day
		ORDER BY day %s`, filterQuery, withFillQuery)
	return analyzer.store.Select(query, args...)
}

// TotalSessionDuration returns the total session duration in seconds.
func (analyzer *Analyzer) TotalSessionDuration(filter *Filter) (int, error) {
	args, filterQuery := analyzer.getFilter(filter).query()
	query := fmt.Sprintf(`SELECT sum(duration) average_time_spent_seconds
		FROM (
			SELECT toDate(time) day, max(time)-min(time) duration
			FROM hit
			WHERE %s
			AND session IS NOT NULL
			GROUP BY day, fingerprint, session
		)`, filterQuery)
	stats, err := analyzer.store.Get(query, args...)

	if err != nil {
		return 0, err
	}

	return stats.AverageTimeSpentSeconds, nil
}

// AvgTimeOnPage returns the average time on page grouped by path.
func (analyzer *Analyzer) AvgTimeOnPage(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "AND " + fieldQuery
	}

	query := fmt.Sprintf(`SELECT "path", toUInt64(avg(time_on_page)) average_time_spent_seconds
		FROM (
			SELECT "path", neighbor(previous_time_on_page_seconds, 1, 0) time_on_page
			FROM (
				SELECT *
				FROM hit
				WHERE %s
				ORDER BY fingerprint, time
			)
			WHERE time_on_page > 0
			%s
		)
		GROUP BY "path"
		ORDER BY "path"`, timeQuery, fieldQuery)
	timeArgs = append(timeArgs, fieldArgs...)
	return analyzer.store.Select(query, timeArgs...)
}

// TotalTimeOnPage returns the total time on page in seconds.
func (analyzer *Analyzer) TotalTimeOnPage(filter *Filter) (int, error) {
	filter = analyzer.getFilter(filter)
	timeArgs, timeQuery := filter.queryTime()
	fieldArgs, fieldQuery := filter.queryFields()

	if fieldQuery != "" {
		fieldQuery = "WHERE " + fieldQuery
	}

	query := fmt.Sprintf(`SELECT sum(time_on_page) average_time_spent_seconds
		FROM (
			SELECT neighbor(previous_time_on_page_seconds, 1, 0) time_on_page
			FROM (
				SELECT *
				FROM hit
				WHERE %s
				ORDER BY fingerprint, time
			)
			%s
		)`, timeQuery, fieldQuery)
	timeArgs = append(timeArgs, fieldArgs...)
	stats, err := analyzer.store.Get(query, timeArgs...)

	if err != nil {
		return 0, err
	}

	return stats.AverageTimeSpentSeconds, nil
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

func (analyzer *Analyzer) selectByAttribute(filter *Filter, attr string) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	args, filterQuery := filter.query()
	query := fmt.Sprintf(byAttributeQuery, attr, filterQuery, attr, filterQuery, attr, attr, filter.withLimit())
	args = append(args, args...)
	return analyzer.store.Select(query, args...)
}

func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		return NewFilter(NullClient)
	}

	filter.validate()
	return filter
}
