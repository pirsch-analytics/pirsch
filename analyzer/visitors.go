package analyzer

import (
	"errors"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"time"
)

var (
	// ErrNoPeriodOrDay is returned in case no period or day was specified to calculate the growth rate.
	ErrNoPeriodOrDay = errors.New("no period or day specified")
)

// Visitors aggregates statistics regarding visitors.
type Visitors struct {
	analyzer *Analyzer
	store    db.Store
}

// Active returns the active visitors per path and (optional) page title and the total number of active visitors for given duration.
// Use time.Minute*5 for example to get the active visitors for the past 5 minutes.
func (visitors *Visitors) Active(filter *Filter, duration time.Duration) ([]model.ActiveVisitorStats, int, error) {
	// TODO
	return []model.ActiveVisitorStats{}, 0, nil

	/*filter = visitors.analyzer.getFilter(filter)
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

	if visitors.analyzer.minIsBot > 0 || len(filter.EntryPath) != 0 || len(filter.ExitPath) != 0 {
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
	stats, err := visitors.store.SelectActiveVisitorStats(title != "", query.String(), args...)

	if err != nil {
		return nil, 0, err
	}

	query.Reset()
	query.WriteString(`SELECT uniq(visitor_id) visitors
		FROM page_view v `)

	if visitors.analyzer.minIsBot > 0 || len(filter.EntryPath) != 0 || len(filter.ExitPath) != 0 {
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
	count, err := visitors.store.Count(query.String(), args...)

	if err != nil {
		return nil, 0, err
	}

	return stats, count, nil*/
}

// Total returns the total visitor count, session count, bounce rate, and views.
func (visitors *Visitors) Total(filter *Filter) (*model.TotalVisitorStats, error) {
	q, args := visitors.analyzer.getFilter(filter).buildQuery([]Field{
		FieldVisitors,
		FieldSessions,
		FieldViews,
		FieldBounces,
		FieldBounceRate,
	}, nil, nil)
	stats, err := visitors.store.GetTotalVisitorStats(q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ByPeriod returns the visitor count, session count, bounce rate, and views grouped by day, week, month, or year.
func (visitors *Visitors) ByPeriod(filter *Filter) ([]model.VisitorStats, error) {
	filter = visitors.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
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
	stats, err := visitors.store.SelectVisitorStats(filter.Period, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Growth returns the growth rate for visitor count, session count, bounces, views, and average session duration or average time on page (if path is set).
// The growth rate is relative to the previous time range or day.
// The period or day for the filter must be set, else an error is returned.
func (visitors *Visitors) Growth(filter *Filter) (*model.Growth, error) {
	filter = visitors.analyzer.getFilter(filter)

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
	q, args := filter.buildQuery(fields, nil, nil)
	current, err := visitors.store.GetGrowthStats(q, args...)

	if err != nil {
		return nil, err
	}

	var currentTimeSpent int

	if len(filter.EventName) != 0 {
		currentTimeSpent, err = visitors.totalEventDuration(filter)
	} else if len(filter.Path) == 0 {
		currentTimeSpent, err = visitors.totalSessionDuration(filter)
	} else {
		currentTimeSpent, err = visitors.totalTimeOnPage(filter)
	}

	if err != nil {
		return nil, err
	}

	// get previous statistics
	if filter.From.Equal(filter.To) {
		if filter.To.Equal(util.Today()) {
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

	q, args = filter.buildQuery(fields, nil, nil)
	previous, err := visitors.store.GetGrowthStats(q, args...)

	if err != nil {
		return nil, err
	}

	var previousTimeSpent int

	if len(filter.EventName) != 0 {
		previousTimeSpent, err = visitors.totalEventDuration(filter)
	} else if len(filter.Path) == 0 {
		previousTimeSpent, err = visitors.totalSessionDuration(filter)
	} else {
		previousTimeSpent, err = visitors.totalTimeOnPage(filter)
	}

	if err != nil {
		return nil, err
	}

	return &model.Growth{
		VisitorsGrowth:  calculateGrowth(current.Visitors, previous.Visitors),
		ViewsGrowth:     calculateGrowth(current.Views, previous.Views),
		SessionsGrowth:  calculateGrowth(current.Sessions, previous.Sessions),
		BouncesGrowth:   calculateGrowth(current.BounceRate, previous.BounceRate),
		TimeSpentGrowth: calculateGrowth(currentTimeSpent, previousTimeSpent),
	}, nil
}

// ByHour returns the visitor count grouped by time of day.
func (visitors *Visitors) ByHour(filter *Filter) ([]model.VisitorHourStats, error) {
	q, args := visitors.analyzer.getFilter(filter).buildQuery([]Field{
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
	stats, err := visitors.store.SelectVisitorHourStats(q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Referrer returns the visitor count and bounce rate grouped by referrer.
func (visitors *Visitors) Referrer(filter *Filter) ([]model.ReferrerStats, error) {
	filter = visitors.analyzer.getFilter(filter)
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

	if len(filter.Referrer) != 0 || len(filter.ReferrerName) != 0 {
		fields = append(fields, FieldReferrer)
		groupBy = append(groupBy, FieldReferrer)
		orderBy = append(orderBy, FieldReferrer)
	} else {
		fields = append(fields, FieldAnyReferrer)
	}

	q, args := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := visitors.store.SelectReferrerStats(q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (visitors *Visitors) totalSessionDuration(filter *Filter) (int, error) {
	// TODO
	return 0, nil

	/*filterArgs, filterQuery := filter.query(true)
	innerFilterArgs, innerFilterQuery := filter.queryTime(false)
	args := make([]any, 0, len(innerFilterArgs)+len(filterArgs))
	var query strings.Builder
	query.WriteString(`SELECT sum(duration_seconds)
		FROM (
			SELECT sum(duration_seconds*sign) duration_seconds
			FROM session s `)

	if len(filter.Path) != 0 || len(filter.PathPattern) != 0 {
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
	averageTimeSpentSeconds, err := visitors.store.Count(query.String(), args...)

	if err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil*/
}

func (visitors *Visitors) totalEventDuration(filter *Filter) (int, error) {
	// TODO
	return 0, nil

	/*filterArgs, filterQuery := filter.query(false)
	var query string

	if visitors.analyzer.minIsBot > 0 {
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

	averageTimeSpentSeconds, err := visitors.store.Count(query, filterArgs...)

	if err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil*/
}

func (visitors *Visitors) totalTimeOnPage(filter *Filter) (int, error) {
	// TODO
	return 0, nil

	/*timeArgs, timeQuery := filter.queryTime(false)
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
				FROM page_view v `, visitors.analyzer.timeOnPageQuery(filter), fieldsQuery))

	if visitors.analyzer.minIsBot > 0 || len(filter.EntryPath) != 0 || len(filter.ExitPath) != 0 {
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
	averageTimeSpentSeconds, err := visitors.store.Count(query.String(), args...)

	if err != nil {
		return 0, err
	}

	return averageTimeSpentSeconds, nil*/
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
