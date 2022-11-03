package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v4"
	"strconv"
	"strings"
)

const (
	sessions   = "session"
	pageViews  = "page_view"
	events     = "events"
	dateFormat = "2006-01-02"
)

type table string

type where struct {
	eqContains []string
	notEq      []string
}

type query struct {
	filter   *Filter
	fields   []Field
	from     table
	join     *query
	leftJoin *query
	groupBy  []Field
	orderBy  []Field

	where []where
	q     strings.Builder
	args  []any
}

func (query *query) query() (string, []any) {
	query.args = make([]any, 0)
	query.selectFields()
	query.fromTable()
	query.joinQuery()
	query.q.WriteString(query.whereTime(true))
	query.whereFields()
	query.groupByFields()
	query.having()
	query.orderByFields()
	query.withLimit()
	return query.q.String(), query.args
}

func (query *query) selectFields() {
	query.q.WriteString("SELECT ")
	var q strings.Builder

	for i := range query.fields {
		if query.fields[i].filterTime {
			timeQuery := query.whereTime(false)
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

func (query *query) whereTime(filterBots bool) string {
	query.args = append(query.args, query.filter.ClientID)
	var q strings.Builder

	if filterBots {
		q.WriteString("WHERE ")
	}

	q.WriteString("client_id = ? ")
	tz := query.filter.Timezone.String()

	if !query.filter.From.IsZero() && !query.filter.To.IsZero() && query.filter.From.Equal(query.filter.To) {
		query.args = append(query.args, query.filter.From.Format(dateFormat))
		q.WriteString(fmt.Sprintf("AND toDate(time, '%s') = toDate(?) ", tz))
	} else {
		if !query.filter.From.IsZero() {
			if query.filter.IncludeTime {
				query.args = append(query.args, query.filter.From)
				q.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') >= toDateTime(?, '%s') ", tz, tz))
			} else {
				query.args = append(query.args, query.filter.From.Format(dateFormat))
				q.WriteString(fmt.Sprintf("AND toDate(time, '%s') >= toDate(?) ", tz))
			}
		}

		if !query.filter.To.IsZero() {
			if query.filter.IncludeTime {
				query.args = append(query.args, query.filter.To)
				q.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') <= toDateTime(?, '%s') ", tz, tz))
			} else {
				query.args = append(query.args, query.filter.To.Format(dateFormat))
				q.WriteString(fmt.Sprintf("AND toDate(time, '%s') <= toDate(?) ", tz))
			}
		}
	}

	if filterBots && query.filter.minIsBot > 0 {
		query.args = append(query.args, query.filter.minIsBot)
		q.WriteString(" AND is_bot < ? ")
	}

	return q.String()
}

func (query *query) whereFields() {
	query.whereField("path", query.filter.Path)
	query.whereField("entry_path", query.filter.EntryPath)
	query.whereField("exit_path", query.filter.ExitPath)
	query.whereField("language", query.filter.Language)
	query.whereField("country_code", query.filter.Country)
	query.whereField("city", query.filter.City)
	query.whereField("referrer", query.filter.Referrer)
	query.whereField("referrer_name", query.filter.ReferrerName)
	query.whereField("os", query.filter.OS)
	query.whereField("os_version", query.filter.OSVersion)
	query.whereField("browser", query.filter.Browser)
	query.whereField("browser_version", query.filter.BrowserVersion)
	query.whereField("screen_class", query.filter.ScreenClass)
	query.whereFieldUInt16("screen_width", query.filter.ScreenWidth)
	query.whereFieldUInt16("screen_height", query.filter.ScreenHeight)
	query.whereField("utm_source", query.filter.UTMSource)
	query.whereField("utm_medium", query.filter.UTMMedium)
	query.whereField("utm_campaign", query.filter.UTMCampaign)
	query.whereField("utm_content", query.filter.UTMContent)
	query.whereField("utm_term", query.filter.UTMTerm)
	query.whereField("event_name", query.filter.EventName)
	query.whereField("event_meta_keys", query.filter.EventMetaKey)
	query.whereFieldPlatform()
	query.whereFieldPathPattern()
	query.whereFieldMeta()

	for i := range query.filter.Search {
		query.whereFieldSearch(query.filter.Search[i].Field.Name, query.filter.Search[i].Input)
	}

	if len(query.where) > 0 {
		query.q.WriteString("AND ")
		parts := make([]string, 0, len(query.where))

		for _, fields := range query.where {
			if len(fields.eqContains) > 1 {
				parts = append(parts, fmt.Sprintf("(%s) ", strings.Join(fields.eqContains, "OR ")))
			} else if len(fields.eqContains) == 1 {
				parts = append(parts, fields.eqContains[0])
			}

			if len(fields.notEq) > 1 {
				parts = append(parts, strings.Join(fields.notEq, " AND "))
			} else if len(fields.notEq) == 1 {
				parts = append(parts, fields.notEq[0])
			}
		}

		query.q.WriteString(strings.Join(parts, "AND "))
	}
}

func (query *query) whereField(field string, value []string) {
	if len(value) != 0 {
		var group where
		eqContainsArgs := make([]any, 0, len(value))
		notEqArgs := make([]any, 0, len(value))

		for _, v := range value {
			comparator := "%s = ? "
			not := strings.HasPrefix(v, "!")

			if field == "event_meta_keys" {
				if not {
					v = v[1:]
					comparator = "!has(%s, ?) "
				} else {
					comparator = "has(%s, ?) "
				}
			} else if not {
				v = v[1:]
				comparator = "%s != ? "
			} else if strings.HasPrefix(v, "~") {
				if field == FieldLanguage.Name || field == FieldCountry.Name {
					v = v[1:]
					comparator = "has(splitByChar(',', ?), %s) = 1 "
				} else {
					v = fmt.Sprintf("%%%s%%", v[1:])
					comparator = "ilike(%s, ?) = 1 "
				}
			}

			if not {
				notEqArgs = append(notEqArgs, query.nullValue(v))
				group.notEq = append(group.notEq, fmt.Sprintf(comparator, field))
			} else {
				eqContainsArgs = append(eqContainsArgs, query.nullValue(v))
				group.eqContains = append(group.eqContains, fmt.Sprintf(comparator, field))
			}
		}

		for _, v := range eqContainsArgs {
			query.args = append(query.args, v)
		}

		for _, v := range notEqArgs {
			query.args = append(query.args, v)
		}

		query.where = append(query.where, group)
	}
}

func (query *query) whereFieldSearch(field, value string) {
	if value != "" {
		comparator := ""

		if field == FieldLanguage.Name || field == FieldCountry.Name {
			comparator = "has(splitByChar(',', ?), %s) = 1 "

			if strings.HasPrefix(value, "!") {
				value = value[1:]
				comparator = "has(splitByChar(',', ?), %s) = 0 "
			}

			query.args = append(query.args, value)
		} else {
			comparator = "ilike(%s, ?) = 1 "

			if strings.HasPrefix(value, "!") {
				value = value[1:]
				comparator = "ilike(%s, ?) = 0 "
			}

			query.args = append(query.args, fmt.Sprintf("%%%s%%", value))
		}

		// use eqContains because it doesn't matter for a single field
		query.where = append(query.where, where{eqContains: []string{fmt.Sprintf(comparator, field)}})
	}
}

func (query *query) whereFieldUInt16(field string, value []string) {
	if len(value) != 0 {
		var group where
		eqContainsArgs := make([]any, 0, len(value))
		notEqArgs := make([]any, 0, len(value))

		for _, v := range value {
			comparator := "%s = ? "
			not := strings.HasPrefix(v, "!")

			if not {
				v = v[1:]
				comparator = "%s != ? "
			}

			var valueInt uint16

			if strings.ToLower(v) != "null" {
				i, err := strconv.ParseUint(v, 10, 16)

				if err == nil {
					valueInt = uint16(i)
				}
			}

			if not {
				notEqArgs = append(notEqArgs, valueInt)
				group.notEq = append(group.notEq, fmt.Sprintf(comparator, field))
			} else {
				eqContainsArgs = append(eqContainsArgs, valueInt)
				group.eqContains = append(group.eqContains, fmt.Sprintf(comparator, field))
			}
		}

		for _, v := range eqContainsArgs {
			query.args = append(query.args, v)
		}

		for _, v := range notEqArgs {
			query.args = append(query.args, v)
		}

		query.where = append(query.where, group)
	}
}

func (query *query) whereFieldMeta() {
	if len(query.filter.EventMeta) != 0 {
		var group where

		for k, v := range query.filter.EventMeta {
			comparator := "event_meta_values[indexOf(event_meta_keys, '%s')] = ? "

			if strings.HasPrefix(v, "!") {
				v = v[1:]
				comparator = "event_meta_values[indexOf(event_meta_keys, '%s')] != ? "
			} else if strings.HasPrefix(v, "~") {
				v = fmt.Sprintf("%%%s%%", v[1:])
				comparator = "ilike(event_meta_values[indexOf(event_meta_keys, '%s')], ?) = 1 "
			}

			// use notEq because they will all be joined using AND
			query.args = append(query.args, query.nullValue(v))
			group.notEq = append(group.notEq, fmt.Sprintf(comparator, k))
		}

		query.where = append(query.where, group)
	}
}

func (query *query) whereFieldPlatform() {
	if query.filter.Platform != "" {
		if strings.HasPrefix(query.filter.Platform, "!") {
			platform := query.filter.Platform[1:]

			if platform == pirsch.PlatformDesktop {
				query.where = append(query.where, where{notEq: []string{"desktop != 1 "}})
			} else if platform == pirsch.PlatformMobile {
				query.where = append(query.where, where{notEq: []string{"mobile != 1 "}})
			} else {
				query.where = append(query.where, where{notEq: []string{"(desktop = 1 OR mobile = 1) "}})
			}
		} else {
			if query.filter.Platform == pirsch.PlatformDesktop {
				query.where = append(query.where, where{eqContains: []string{"desktop = 1 "}})
			} else if query.filter.Platform == pirsch.PlatformMobile {
				query.where = append(query.where, where{eqContains: []string{"mobile = 1 "}})
			} else {
				query.where = append(query.where, where{eqContains: []string{"desktop = 0 AND mobile = 0 "}})
			}
		}
	}
}

func (query *query) whereFieldPathPattern() {
	if len(query.filter.PathPattern) != 0 {
		var group where

		for _, pattern := range query.filter.PathPattern {
			if strings.HasPrefix(pattern, "!") {
				query.args = append(query.args, pattern[1:])
				group.notEq = append(group.notEq, `match("path", ?) = 0 `)
			} else {
				query.args = append(query.args, pattern)
				group.eqContains = append(group.eqContains, `match("path", ?) = 1 `)
			}
		}

		query.where = append(query.where, group)
	}
}

func (query *query) nullValue(value string) string {
	if strings.ToLower(value) == "null" {
		return ""
	}

	return value
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
				fillQuery := query.withFill()
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
		query.q.WriteString(str[:len(str)-1] + " ")
	}
}

func (query *query) withFill() string {
	if !query.filter.From.IsZero() && !query.filter.To.IsZero() {
		q := ""

		switch query.filter.Period {
		case pirsch.PeriodDay:
			q = "WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY "
		case pirsch.PeriodWeek:
			q = "WITH FILL FROM toStartOfWeek(toDate(?), 1) TO toDate(?)+1 STEP INTERVAL 1 WEEK "
		case pirsch.PeriodMonth:
			q = "WITH FILL FROM toStartOfMonth(toDate(?)) TO toDate(?)+1 STEP INTERVAL 1 MONTH "
		case pirsch.PeriodYear:
			q = "WITH FILL FROM toStartOfYear(toDate(?)) TO toDate(?)+1 STEP INTERVAL 1 YEAR "
		}

		query.args = append(query.args, query.filter.From.Format(dateFormat), query.filter.To.Format(dateFormat))
		return q
	}

	return ""
}

func (query *query) withLimit() {
	if query.filter.Limit > 0 && query.filter.Offset > 0 {
		query.q.WriteString(fmt.Sprintf("LIMIT %d OFFSET %d ", query.filter.Limit, query.filter.Offset))
	} else if query.filter.Limit > 0 {
		query.q.WriteString(fmt.Sprintf("LIMIT %d ", query.filter.Limit))
	}
}
