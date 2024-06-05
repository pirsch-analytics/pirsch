package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"strings"
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
	query.WriteString(fmt.Sprintf(`SELECT "day", toUInt64(greatest(round(avg(duration)), 0)) average_time_spent_seconds
		FROM (
			SELECT toDate(time, '%s') "day", sum(duration_seconds*sign)/sum(sign) duration
			FROM "session" s `, filter.Timezone.String()))

	if len(filter.Path) > 0 || len(filter.PathPattern) > 0 || len(filter.Tag) > 0 || len(filter.Tags) > 0 {
		tagField := ""

		if len(filter.Tags) > 0 {
			tagField = fmt.Sprintf(", %s, %s", FieldTagKeysRaw.Name, FieldTagValuesRaw.Name)
		} else if len(filter.Tag) > 0 {
			tagField = fmt.Sprintf(", %s", FieldTagKeysRaw.Name)
		}

		query.WriteString(fmt.Sprintf(`INNER JOIN (
			SELECT visitor_id,
			session_id,
			path %s
			FROM page_view
			WHERE %s
		) v
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, tagField, q.whereTime()[len("WHERE "):]))
	}

	query.WriteString(q.whereTime())
	q.from = pageViews
	q.whereFields()
	where := q.q.String()

	if where != "" {
		query.WriteString(where)
	}

	query.WriteString(fmt.Sprintf(`AND duration_seconds != 0
			GROUP BY "day", visitor_id, session_id
			HAVING sum(sign) > 0
		)
		GROUP BY "day"
		ORDER BY "day"
		%s `, q.withFill()))
	t.groupByPeriod(filter.Period, &query)
	stats, err := t.store.SelectTimeSpentStats(filter.Ctx, filter.Period, query.String(), q.args...)

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
		from:   pageViews,
		search: filter.Search,
	}
	var query strings.Builder
	t.selectAvgTimeSpentPeriod(filter.Period, &query)
	fields := q.getFields()
	filterFields := strings.Join(fields, ",")

	if filterFields != "" {
		filterFields = "," + filterFields
	}

	fields = append(fields, FieldEventDurationSeconds.Name)
	query.WriteString(fmt.Sprintf(`SELECT "day", toUInt64(greatest(ifNotFinite(round(avg(time_on_page)), 0), 0)) average_time_spent_seconds
		FROM (
			SELECT toDate(time, '%s') "day",
				nth_value(%s, 2) OVER (PARTITION BY v.visitor_id, v.session_id ORDER BY v."time" ASC Rows BETWEEN CURRENT ROW AND 1 FOLLOWING) AS time_on_page
				%s
			FROM page_view v `, filter.Timezone.String(), t.analyzer.timeOnPageQuery(filter), filterFields))

	if len(filter.EntryPath) > 0 || len(filter.ExitPath) > 0 {
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

	query.WriteString(fmt.Sprintf(`WHERE %s)
		WHERE time_on_page > 0 `, q.whereTime()[len("WHERE "):]))
	q.whereFields()
	where := q.q.String()
	query.WriteString(fmt.Sprintf(`%s
		GROUP BY "day"
		ORDER BY "day"
		%s`, where, q.withFill()))
	t.groupByPeriod(filter.Period, &query)
	stats, err := t.store.SelectTimeSpentStats(filter.Ctx, filter.Period, query.String(), q.args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (t *Time) selectAvgTimeSpentPeriod(period pkg.Period, query *strings.Builder) {
	if period != pkg.PeriodDay {
		switch period {
		case pkg.PeriodWeek:
			query.WriteString(`SELECT toUInt64(greatest(round(avg(average_time_spent_seconds)), 0)) average_time_spent_seconds, toStartOfWeek("day", 1) week FROM (`)
		case pkg.PeriodMonth:
			query.WriteString(`SELECT toUInt64(greatest(round(avg(average_time_spent_seconds)), 0)) average_time_spent_seconds, toStartOfMonth("day") month FROM (`)
		case pkg.PeriodYear:
			query.WriteString(`SELECT toUInt64(greatest(round(avg(average_time_spent_seconds)), 0)) average_time_spent_seconds, toStartOfYear("day") year FROM (`)
		default:
			panic("unknown case for filter period")
		}
	}
}

func (t *Time) groupByPeriod(period pkg.Period, query *strings.Builder) {
	if period != pkg.PeriodDay {
		switch period {
		case pkg.PeriodWeek:
			query.WriteString(`) GROUP BY week ORDER BY week ASC`)
		case pkg.PeriodMonth:
			query.WriteString(`) GROUP BY month ORDER BY month ASC`)
		case pkg.PeriodYear:
			query.WriteString(`) GROUP BY year ORDER BY year ASC`)
		default:
			panic("unknown case for filter period")
		}
	}
}
