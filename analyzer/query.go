package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v4"
	"strings"
)

const (
	sessions   = "session"
	pageViews  = "page_view"
	events     = "events"
	dateFormat = "2006-01-02"
)

type table string

type query struct {
	filter   *Filter
	fields   []Field
	from     table
	join     *query
	leftJoin *query
	where    []Field
	groupBy  []Field
	orderBy  []Field
	q        strings.Builder
	args     []any
}

func (query *query) query() (string, []any) {
	query.args = make([]any, 0)
	query.selectFields()
	query.fromTable()
	query.joinQuery()
	query.whereTime()
	query.whereFields()
	query.groupByFields()
	query.having()
	query.orderByFields()
	return query.q.String(), query.args
}

func (query *query) selectFields() {
	query.q.WriteString("SELECT ")
	var q strings.Builder

	for i := range query.fields {
		if query.fields[i].filterTime {
			timeArgs, timeQuery := query.filter.queryTime(false)
			query.args = append(query.args, timeArgs...)
			q.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(query.fields[i].queryPageViews, timeQuery), query.fields[i].Name))
		} else if query.fields[i].timezone {
			if query.fields[i].Name == "day" && query.filter.Period != pirsch.PeriodDay {
				switch query.filter.Period {
				case pirsch.PeriodWeek:
					q.WriteString(fmt.Sprintf(`toStartOfWeek(%s, 1) week,`, fmt.Sprintf(query.fields[i].queryPageViews, query.filter.Timezone.String())))
				case pirsch.PeriodMonth:
					q.WriteString(fmt.Sprintf(`toStartOfMonth(%s) month,`, fmt.Sprintf(query.fields[i].queryPageViews, query.filter.Timezone.String())))
				case pirsch.PeriodYear:
					q.WriteString(fmt.Sprintf(`toStartOfYear(%s) year,`, fmt.Sprintf(query.fields[i].queryPageViews, query.filter.Timezone.String())))
				}
			} else {
				q.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(query.fields[i].queryPageViews, query.filter.Timezone.String()), query.fields[i].Name))
			}
		} else if query.fields[i].Name == "meta_value" {
			query.args = append(query.args, query.filter.EventMetaKey)
			q.WriteString(fmt.Sprintf(`%s %s,`, query.fields[i].queryPageViews, query.fields[i].Name))
		} else {
			q.WriteString(fmt.Sprintf(`%s %s,`, query.fields[i].queryPageViews, query.fields[i].Name))
		}
	}

	str := q.String()
	query.q.WriteString(str[:len(str)-1] + " ")
}

func (query *query) fromTable() {
	query.q.WriteString(fmt.Sprintf("FROM %s ", query.from))
}

func (query *query) joinQuery() {
	// TODO
}

func (query *query) whereTime() {
	query.args = append(query.args, query.filter.ClientID)
	query.q.WriteString("WHERE client_id = ? ")
	tz := query.filter.Timezone.String()

	if !query.filter.From.IsZero() && !query.filter.To.IsZero() && query.filter.From.Equal(query.filter.To) {
		query.args = append(query.args, query.filter.From.Format(dateFormat))
		query.q.WriteString(fmt.Sprintf("AND toDate(time, '%s') = toDate(?) ", tz))
	} else {
		if !query.filter.From.IsZero() {
			if query.filter.IncludeTime {
				query.args = append(query.args, query.filter.From)
				query.q.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') >= toDateTime(?, '%s') ", tz, tz))
			} else {
				query.args = append(query.args, query.filter.From.Format(dateFormat))
				query.q.WriteString(fmt.Sprintf("AND toDate(time, '%s') >= toDate(?) ", tz))
			}
		}

		if !query.filter.To.IsZero() {
			if query.filter.IncludeTime {
				query.args = append(query.args, query.filter.To)
				query.q.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') <= toDateTime(?, '%s') ", tz, tz))
			} else {
				query.args = append(query.args, query.filter.To.Format(dateFormat))
				query.q.WriteString(fmt.Sprintf("AND toDate(time, '%s') <= toDate(?) ", tz))
			}
		}
	}
}

func (query *query) whereFields() {
	// TODO
}

func (query *query) groupByFields() {
	if len(query.groupBy) > 0 {
		query.q.WriteString("GROUP BY ")
		var q strings.Builder

		for i := range query.groupBy {
			if query.groupBy[i].Name == "day" && query.filter.Period != pirsch.PeriodDay {
				switch query.filter.Period {
				case pirsch.PeriodWeek:
					q.WriteString("week,")
				case pirsch.PeriodMonth:
					q.WriteString("month,")
				case pirsch.PeriodYear:
					q.WriteString("year,")
				}
			} else {
				q.WriteString(query.groupBy[i].Name + ",")
			}
		}

		str := q.String()
		query.q.WriteString(str[:len(str)-1] + " ")
	}
}

func (query *query) having() {
	if query.from == sessions {
		query.q.WriteString("HAVING sum(sign) > 0 ")
	}
}

func (query *query) orderByFields() {
	if len(query.filter.Sort) > 0 {
		query.orderBy = make([]Field, 0, len(query.filter.Sort))

		for i := range query.filter.Sort {
			query.filter.Sort[i].Field.queryDirection = string(query.filter.Sort[i].Direction)
			query.orderBy = append(query.orderBy, query.filter.Sort[i].Field)
		}
	}

	if len(query.orderBy) > 0 {
		query.q.WriteString("ORDER BY ")
		var q strings.Builder

		for i := range query.orderBy {
			if query.orderBy[i].queryWithFill != "" {
				q.WriteString(fmt.Sprintf(`%s %s %s,`, query.orderBy[i].Name, query.orderBy[i].queryDirection, query.orderBy[i].queryWithFill))
			} else if query.orderBy[i].withFill {
				fillArgs, fillQuery := query.filter.withFill()
				query.args = append(query.args, fillArgs...)
				name := query.orderBy[i].Name

				if query.orderBy[i].Name == "day" && query.filter.Period != pirsch.PeriodDay {
					switch query.filter.Period {
					case pirsch.PeriodWeek:
						name = "week"
					case pirsch.PeriodMonth:
						name = "month"
					case pirsch.PeriodYear:
						name = "year"
					}
				}

				q.WriteString(fmt.Sprintf(`%s %s %s,`, name, query.orderBy[i].queryDirection, fillQuery))
			} else {
				q.WriteString(fmt.Sprintf(`%s %s,`, query.orderBy[i].Name, query.orderBy[i].queryDirection))
			}
		}

		str := q.String()
		query.q.WriteString(str[:len(str)-1])
	}
}
