package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"strings"
	"time"
)

// Time aggregates statistics regarding the time on page and session duration.
type Time struct {
	analyzer *Analyzer
	store    db.Store
}

// AvgSessionDuration returns the average session duration grouped by day, week, month, or year.
func (t *Time) AvgSessionDuration(filter *Filter) ([]model.TimeSpentStats, error) {
	filter = t.analyzer.getFilter(filter)
	table := filter.table([]Field{})

	if table == events {
		return []model.TimeSpentStats{}, nil
	}

	q := queryBuilder{
		filter: filter,
		from:   table,
		search: filter.Search,
	}
	var query strings.Builder
	t.selectAvgTimeSpentPeriod(filter.Period, &query)
	query.WriteString(fmt.Sprintf(`SELECT "day", toUInt64(ifNotFinite(round(duration/n), 0)) average_time_spent_seconds
		FROM (
			SELECT toDate(time, '%s') "day",
			sum(duration_seconds*sign) duration,
			uniq(visitor_id, session_id) n
			FROM "session" s `, time.UTC.String()))

	if len(filter.Path) != 0 || len(filter.PathPattern) != 0 {
		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			path
			FROM page_view
			WHERE %s
		) v
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, q.whereTime()[len("WHERE "):]))
	}

	query.WriteString(q.whereTime())
	q.whereFields()
	where := q.q.String()

	if where != "" {
		query.WriteString(where)
	}

	query.WriteString(fmt.Sprintf(`AND duration_seconds != 0
			GROUP BY "day"
			HAVING sum(sign) > 0
			ORDER BY "day"
			%s
		)`, q.withFill()))
	t.groupByPeriod(filter.Period, &query)
	stats, err := t.store.SelectTimeSpentStats(filter.Period, query.String(), q.args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// AvgTimeOnPage returns the average time on page grouped by day, week, month, or year.
func (t *Time) AvgTimeOnPage(filter *Filter) ([]model.TimeSpentStats, error) {
	filter = t.analyzer.getFilter(filter)
	table := filter.table([]Field{})

	if table == events {
		return []model.TimeSpentStats{}, nil
	}

	q := queryBuilder{
		filter: filter,
		from:   table,
		search: filter.Search,
	}
	var query strings.Builder
	t.selectAvgTimeSpentPeriod(filter.Period, &query)
	fields := q.getFields()
	fields = append(fields, "duration_seconds")
	query.WriteString(fmt.Sprintf(`SELECT "day",
		toUInt64(ifNotFinite(round(avg(time_on_page)), 0)) average_time_spent_seconds
		FROM (
			SELECT "day",
			%s time_on_page
			FROM (
				SELECT session_id,
				toDate(time, '%s') "day",
				%s
				FROM page_view v `, t.analyzer.timeOnPageQuery(filter), time.UTC.String(), strings.Join(fields, ",")))

	if len(filter.EntryPath) != 0 || len(filter.ExitPath) != 0 {
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
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, q.whereTime()[len("WHERE "):]))
	}

	query.WriteString(fmt.Sprintf(`WHERE %s ORDER BY visitor_id, session_id, time)
		WHERE session_id = neighbor(session_id, 1, null) AND time_on_page > 0 `, q.whereTime()[len("WHERE "):]))
	q.whereFields()
	where := q.q.String()
	query.WriteString(fmt.Sprintf(`%s) GROUP BY "day" ORDER BY "day" %s`, where, q.withFill()))
	t.groupByPeriod(filter.Period, &query)
	stats, err := t.store.SelectTimeSpentStats(filter.Period, query.String(), q.args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (t *Time) selectAvgTimeSpentPeriod(period pirsch.Period, query *strings.Builder) {
	if period != pirsch.PeriodDay {
		switch period {
		case pirsch.PeriodWeek:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfWeek("day", 1) week FROM (`)
		case pirsch.PeriodMonth:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfMonth("day") month FROM (`)
		case pirsch.PeriodYear:
			query.WriteString(`SELECT toUInt64(round(avg(average_time_spent_seconds))) average_time_spent_seconds, toStartOfYear("day") year FROM (`)
		}
	}
}

func (t *Time) groupByPeriod(period pirsch.Period, query *strings.Builder) {
	if period != pirsch.PeriodDay {
		switch period {
		case pirsch.PeriodWeek:
			query.WriteString(`) GROUP BY week ORDER BY week ASC`)
		case pirsch.PeriodMonth:
			query.WriteString(`) GROUP BY month ORDER BY month ASC`)
		case pirsch.PeriodYear:
			query.WriteString(`) GROUP BY year ORDER BY year ASC`)
		}
	}
}
