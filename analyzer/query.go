package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v4"
	"strconv"
	"strings"
)

const (
	sessions   = "session t"
	pageViews  = "page_view t"
	events     = "event t"
	dateFormat = "2006-01-02"
)

type table string

type where struct {
	eqContains []string
	notEq      []string
}

type queryBuilder struct {
	filter     *Filter
	fields     []Field
	from       table
	join       *queryBuilder
	joinSecond *queryBuilder
	leftJoin   *queryBuilder
	search     []Search
	groupBy    []Field
	orderBy    []Field
	limit      int
	offset     int

	where []where
	q     strings.Builder
	args  []any
}

func (query *queryBuilder) query() (string, []any) {
	query.args = make([]any, 0)
	combineResults := query.selectFields()

	if !combineResults {
		query.fromTable()
		query.joinQuery()
		query.q.WriteString(query.whereTime())
		query.whereFields()
		query.groupByFields()
		query.having()
		query.orderByFields()
		query.withLimit()
	}

	return query.q.String(), query.args
}

func (query *queryBuilder) getFields() []string {
	fields := make([]string, 0, 25)

	if query.from == sessions {
		query.appendField(&fields, FieldEntryPath.Name, query.filter.EntryPath)
		query.appendField(&fields, FieldExitPath.Name, query.filter.ExitPath)
	} else {
		query.appendField(&fields, FieldPath.Name, query.filter.Path)

		if len(query.filter.Path) == 0 && (len(query.filter.PathPattern) != 0 || len(query.filter.AnyPath) != 0) {
			fields = append(fields, FieldPath.Name)
		}
	}

	if query.from == events {
		query.appendField(&fields, FieldEventName.Name, query.filter.EventName)

		if len(query.filter.EventMeta) > 0 {
			fields = append(fields, "event_meta_keys", "event_meta_values")
		} else {
			query.appendField(&fields, "event_meta_keys", query.filter.EventMetaKey)
		}
	}

	query.appendField(&fields, FieldLanguage.Name, query.filter.Language)
	query.appendField(&fields, FieldCountry.Name, query.filter.Country)
	query.appendField(&fields, FieldCity.Name, query.filter.City)
	query.appendField(&fields, FieldReferrer.Name, query.filter.Referrer)
	query.appendField(&fields, FieldReferrerName.Name, query.filter.ReferrerName)
	query.appendField(&fields, FieldOS.Name, query.filter.OS)
	query.appendField(&fields, FieldOSVersion.Name, query.filter.OSVersion)
	query.appendField(&fields, FieldBrowser.Name, query.filter.Browser)
	query.appendField(&fields, FieldBrowserVersion.Name, query.filter.BrowserVersion)
	query.appendField(&fields, FieldScreenClass.Name, query.filter.ScreenClass)
	query.appendField(&fields, "screen_width", query.filter.ScreenWidth)
	query.appendField(&fields, "screen_height", query.filter.ScreenHeight)
	query.appendField(&fields, FieldUTMSource.Name, query.filter.UTMSource)
	query.appendField(&fields, FieldUTMMedium.Name, query.filter.UTMMedium)
	query.appendField(&fields, FieldUTMCampaign.Name, query.filter.UTMCampaign)
	query.appendField(&fields, FieldUTMContent.Name, query.filter.UTMContent)
	query.appendField(&fields, FieldUTMTerm.Name, query.filter.UTMTerm)

	if query.filter.Platform != "" {
		platform := query.filter.Platform

		if strings.HasPrefix(platform, "!") {
			platform = query.filter.Platform[1:]
		}

		if platform == pirsch.PlatformDesktop {
			fields = append(fields, "desktop")
		} else if platform == pirsch.PlatformMobile {
			fields = append(fields, "mobile")
		} else {
			fields = append(fields, "desktop")
			fields = append(fields, "mobile")
		}
	}

	return fields
}

func (query *queryBuilder) appendField(fields *[]string, field string, value []string) {
	if len(value) != 0 {
		*fields = append(*fields, field)
	}
}

func (query *queryBuilder) selectFields() bool {
	combineResults := false

	if len(query.fields) > 0 {
		query.q.WriteString("SELECT ")
		var q strings.Builder

		for i := range query.fields {
			if query.fields[i].filterTime {
				timeQuery := query.whereTime()[len("WHERE "):]
				q.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(query.selectField(query.fields[i]), timeQuery), query.fields[i].Name))
			} else if query.fields[i].timezone {
				if query.fields[i].Name == FieldDay.Name && query.filter.Period != pirsch.PeriodDay {
					switch query.filter.Period {
					case pirsch.PeriodWeek:
						q.WriteString(fmt.Sprintf(`toStartOfWeek(%s, 1) week,`, fmt.Sprintf(query.selectField(query.fields[i]), query.filter.Timezone.String())))
					case pirsch.PeriodMonth:
						q.WriteString(fmt.Sprintf(`toStartOfMonth(%s) month,`, fmt.Sprintf(query.selectField(query.fields[i]), query.filter.Timezone.String())))
					case pirsch.PeriodYear:
						q.WriteString(fmt.Sprintf(`toStartOfYear(%s) year,`, fmt.Sprintf(query.selectField(query.fields[i]), query.filter.Timezone.String())))
					}
				} else {
					q.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(query.selectField(query.fields[i]), query.filter.Timezone.String()), query.fields[i].Name))
				}
			} else if query.from != sessions && (query.fields[i].Name == FieldPlatformDesktop.Name || query.fields[i].Name == FieldPlatformMobile.Name || query.fields[i].Name == FieldPlatformUnknown.Name) {
				q.WriteString(query.selectPlatform(query.fields[i]))
				combineResults = true
			} else if query.fields[i].Name == FieldEventMetaValues.Name {
				query.args = append(query.args, query.filter.EventMetaKey)
				q.WriteString(fmt.Sprintf(`%s %s,`, query.selectField(query.fields[i]), query.fields[i].Name))
			} else {
				q.WriteString(fmt.Sprintf(`%s %s,`, query.selectField(query.fields[i]), query.fields[i].Name))
			}
		}

		str := q.String()
		query.q.WriteString(str[:len(str)-1] + " ")
	}

	return combineResults
}

func (query *queryBuilder) selectField(field Field) string {
	if query.from == sessions {
		return field.querySessions
	}

	return field.queryPageViews
}

func (query *queryBuilder) selectPlatform(field Field) string {
	var join, leftJoin *queryBuilder

	if query.join != nil {
		joinCopy := *query.join
		join = &joinCopy
	}

	if query.leftJoin != nil {
		leftJoinCopy := *query.leftJoin
		leftJoin = &leftJoinCopy
	}

	q := queryBuilder{
		filter:   query.filter,
		fields:   []Field{FieldVisitors},
		from:     query.from,
		join:     join,
		leftJoin: leftJoin,
		where: []where{
			// use notEq so they are connected by AND
			{notEq: strings.Split(field.queryPageViews, ",")},
		},
	}
	subquery, args := q.query()
	query.args = append(query.args, args...)
	return fmt.Sprintf("toInt64OrDefault((%s)) %s,", subquery, field.Name)
}

func (query *queryBuilder) fromTable() {
	query.q.WriteString(fmt.Sprintf("FROM %s ", query.from))
}

func (query *queryBuilder) joinQuery() {
	if query.join != nil {
		q, args := query.join.query()
		query.args = append(query.args, args...)
		query.q.WriteString(fmt.Sprintf("JOIN (%s) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id ", q))
	}

	if query.joinSecond != nil {
		q, args := query.joinSecond.query()
		query.args = append(query.args, args...)
		query.q.WriteString(fmt.Sprintf("JOIN (%s) k ON k.visitor_id = t.visitor_id AND k.session_id = t.session_id ", q))
	}

	if query.leftJoin != nil {
		q, args := query.join.query()
		query.args = append(query.args, args...)
		query.q.WriteString(fmt.Sprintf("LEFT JOIN (%s) l ON l.visitor_id = t.visitor_id AND l.session_id = t.session_id ", q))
	}
}

func (query *queryBuilder) whereTime() string {
	query.args = append(query.args, query.filter.ClientID)
	var q strings.Builder
	q.WriteString("WHERE client_id = ? ")
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

	if query.from == sessions && query.filter.minIsBot > 0 {
		query.args = append(query.args, query.filter.minIsBot)
		q.WriteString(" AND is_bot < ? ")
	}

	return q.String()
}

func (query *queryBuilder) whereFields() {
	if query.from == sessions {
		query.whereField(FieldEntryPath.Name, query.filter.EntryPath)
		query.whereField(FieldExitPath.Name, query.filter.ExitPath)
	} else {
		query.whereField(FieldPath.Name, query.filter.Path)
		query.whereFieldPathPattern()
		query.whereFieldPathIn()
	}

	if query.from == events {
		query.whereField(FieldEventName.Name, query.filter.EventName)
		query.whereField("event_meta_keys", query.filter.EventMetaKey)
		query.whereFieldMeta()
	}

	query.whereField(FieldLanguage.Name, query.filter.Language)
	query.whereField(FieldCountry.Name, query.filter.Country)
	query.whereField(FieldCity.Name, query.filter.City)
	query.whereField(FieldReferrer.Name, query.filter.Referrer)
	query.whereField(FieldReferrerName.Name, query.filter.ReferrerName)
	query.whereField(FieldOS.Name, query.filter.OS)
	query.whereField(FieldOSVersion.Name, query.filter.OSVersion)
	query.whereField(FieldBrowser.Name, query.filter.Browser)
	query.whereField(FieldBrowserVersion.Name, query.filter.BrowserVersion)
	query.whereField(FieldScreenClass.Name, query.filter.ScreenClass)
	query.whereFieldUInt16("screen_width", query.filter.ScreenWidth)
	query.whereFieldUInt16("screen_height", query.filter.ScreenHeight)
	query.whereField(FieldUTMSource.Name, query.filter.UTMSource)
	query.whereField(FieldUTMMedium.Name, query.filter.UTMMedium)
	query.whereField(FieldUTMCampaign.Name, query.filter.UTMCampaign)
	query.whereField(FieldUTMContent.Name, query.filter.UTMContent)
	query.whereField(FieldUTMTerm.Name, query.filter.UTMTerm)
	query.whereFieldPlatform()

	for i := range query.search {
		query.whereFieldSearch(query.search[i].Field.Name, query.search[i].Input)
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
				parts = append(parts, strings.Join(fields.notEq, " AND ")+" ")
			} else if len(fields.notEq) == 1 {
				parts = append(parts, fields.notEq[0])
			}
		}

		query.q.WriteString(strings.Join(parts, "AND "))
	}
}

func (query *queryBuilder) whereField(field string, value []string) {
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

func (query *queryBuilder) whereFieldSearch(field, value string) {
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

func (query *queryBuilder) whereFieldUInt16(field string, value []string) {
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

func (query *queryBuilder) whereFieldMeta() {
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

func (query *queryBuilder) whereFieldPlatform() {
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

func (query *queryBuilder) whereFieldPathPattern() {
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

func (query *queryBuilder) whereFieldPathIn() {
	if len(query.filter.AnyPath) != 0 {
		for _, path := range query.filter.AnyPath {
			query.args = append(query.args, path)
		}

		q := strings.Repeat("?,", len(query.filter.AnyPath))
		query.where = append(query.where, where{
			eqContains: []string{fmt.Sprintf("path IN (%s) ", q[:len(q)-1])},
		})
	}
}

func (query *queryBuilder) nullValue(value string) string {
	if strings.ToLower(value) == "null" {
		return ""
	}

	return value
}

func (query *queryBuilder) groupByFields() {
	if len(query.groupBy) > 0 {
		query.q.WriteString("GROUP BY ")
		var q strings.Builder

		for i := range query.groupBy {
			if query.groupBy[i].Name == FieldDay.Name && query.filter.Period != pirsch.PeriodDay {
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

func (query *queryBuilder) having() {
	if query.from == sessions {
		query.q.WriteString("HAVING sum(sign) > 0 ")
	}
}

func (query *queryBuilder) orderByFields() {
	if len(query.filter.Sort) > 0 {
		query.orderBy = make([]Field, 0, len(query.filter.Sort))

		for i := range query.filter.Sort {
			for j := range query.orderBy {
				if query.filter.Sort[i].Field == query.orderBy[j] {
					query.orderBy[i].queryDirection = query.filter.Sort[i].Direction
					break
				}
			}
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

				if query.orderBy[i].Name == FieldDay.Name && query.filter.Period != pirsch.PeriodDay {
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

func (query *queryBuilder) withFill() string {
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

func (query *queryBuilder) withLimit() {
	if query.limit > 0 && query.offset > 0 {
		query.q.WriteString(fmt.Sprintf("LIMIT %d OFFSET %d ", query.limit, query.offset))
	} else if query.limit > 0 {
		query.q.WriteString(fmt.Sprintf("LIMIT %d ", query.limit))
	}
}
