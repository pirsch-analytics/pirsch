package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"strconv"
	"strings"
)

const (
	sessions   = table(`"session" t`)
	pageViews  = table(`"page_view" t`)
	events     = table(`"event" t`)
	dateFormat = "2006-01-02"
)

type table string

type where struct {
	eqContains []string
	notEq      []string
}

type queryBuilder struct {
	filter             *Filter
	fields             []Field
	fieldsImported     []Field
	from               table
	fromImported       string
	parent             *queryBuilder
	join               *queryBuilder
	joinSecond         *queryBuilder
	joinThird          *queryBuilder
	leftJoin           *queryBuilder
	joinStep           int
	search             []Search
	groupBy            []Field
	orderBy            []Field
	limit              int
	offset             int
	includeEventFilter bool
	sample             uint
	final              bool

	where []where
	q     strings.Builder
	args  []any
}

func (query *queryBuilder) query() (string, []any) {
	query.args = make([]any, 0)
	includeImported := query.includeImported()
	fromImported := query.fromImported

	if includeImported {
		query.selectFields()
		query.q.WriteString("FROM (")
		query.fromImported = ""
	}

	combineResults := query.selectFields()

	if !combineResults {
		query.fromTable()
		query.joinQuery()
		query.q.WriteString(query.whereTime())
		query.whereFields()
		query.groupByFields(false)
		query.having()

		if includeImported {
			query.orderByFields()
			query.unionImported()
			query.q.WriteString(") t ")
			query.joinImported(fromImported)
			query.groupByFields(true)
		}

		query.orderByFields()
		query.withLimit()
	}

	return query.q.String(), query.args
}

func (query *queryBuilder) getFields() []string {
	fields := make([]string, 0, 40)

	if query.from == sessions {
		query.appendField(&fields, FieldEntryPath.Name, query.filter.EntryPath)
		query.appendField(&fields, FieldExitPath.Name, query.filter.ExitPath)
	} else {
		query.appendField(&fields, FieldPath.Name, query.filter.Path)

		if len(query.filter.Path) == 0 && (len(query.filter.PathPattern) > 0 || len(query.filter.AnyPath) > 0) {
			fields = append(fields, FieldPath.Name)
		}
	}

	if query.from == pageViews {
		if len(query.filter.Tags) > 0 {
			fields = append(fields, FieldTagKeysRaw.Name, FieldTagValuesRaw.Name)
		} else if len(query.filter.Tag) > 0 {
			fields = append(fields, FieldTagKeysRaw.Name)
		}
	}

	if query.from == events {
		query.appendField(&fields, FieldEventName.Name, query.filter.EventName)

		if len(query.filter.EventMeta) > 0 || len(query.filter.Tags) > 0 {
			fields = append(fields, FieldEventMetaKeysRaw.Name, FieldEventMetaValuesRaw.Name)
		} else {
			query.appendField(&fields, FieldEventMetaKeysRaw.Name, query.filter.EventMetaKey)
		}
	}

	query.appendField(&fields, FieldHostname.Name, query.filter.Hostname)
	query.appendField(&fields, FieldLanguage.Name, query.filter.Language)
	query.appendField(&fields, FieldCountry.Name, query.filter.Country)
	query.appendField(&fields, FieldRegion.Name, query.filter.Region)
	query.appendField(&fields, FieldCity.Name, query.filter.City)
	query.appendField(&fields, FieldReferrer.Name, query.filter.Referrer)
	query.appendField(&fields, FieldReferrerName.Name, query.filter.ReferrerName)
	query.appendField(&fields, FieldChannel.Name, query.filter.Channel)
	query.appendField(&fields, FieldOS.Name, query.filter.OS)
	query.appendField(&fields, FieldOSVersion.Name, query.filter.OSVersion)
	query.appendField(&fields, FieldBrowser.Name, query.filter.Browser)
	query.appendField(&fields, FieldBrowserVersion.Name, query.filter.BrowserVersion)
	query.appendField(&fields, FieldScreenClass.Name, query.filter.ScreenClass)
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

		if platform == pkg.PlatformDesktop {
			fields = append(fields, "desktop")
		} else if platform == pkg.PlatformMobile {
			fields = append(fields, "mobile")
		} else {
			fields = append(fields, "desktop")
			fields = append(fields, "mobile")
		}
	}

	return fields
}

func (query *queryBuilder) appendField(fields *[]string, field string, value []string) {
	if len(value) > 0 {
		*fields = append(*fields, field)
	}
}

func (query *queryBuilder) selectFields() bool {
	includeImported := query.includeImported()
	combineResults := false

	if len(query.fields) > 0 {
		query.q.WriteString("SELECT ")
		var q strings.Builder

		for i := range query.fields {
			if query.fields[i].filterTime {
				sampleFactor, sampleQuery := "", ""

				if query.sample > 0 {
					sampleFactor = "*any(_sample_factor)"
					sampleQuery = fmt.Sprintf(" SAMPLE %d", query.sample)
				}

				timeQuery := query.whereTime()[len("WHERE "):]

				if includeImported {
					dateQuery := query.whereTimeImported()[len("WHERE "):]
					q.WriteString(fmt.Sprintf("%s %s,", fmt.Sprintf(query.selectField(query.fields[i]), sampleFactor, sampleQuery, timeQuery, query.fromImported, dateQuery), query.fields[i].Name))
				} else {
					q.WriteString(fmt.Sprintf("%s %s,", fmt.Sprintf(query.selectField(query.fields[i]), sampleFactor, sampleQuery, timeQuery), query.fields[i].Name))
				}
			} else if query.fields[i].timezone {
				withTz := ""

				if includeImported {
					withTz = query.selectField(query.fields[i])
				} else {
					if query.fields[i] == FieldWeekday {
						withTz = fmt.Sprintf(query.selectField(query.fields[i]), query.filter.WeekdayMode, query.filter.Timezone.String())
					} else {
						withTz = fmt.Sprintf(query.selectField(query.fields[i]), query.filter.Timezone.String())
					}
				}

				if query.filter.Period != pkg.PeriodDay && query.fields[i] == FieldDay {
					if includeImported {
						switch query.filter.Period {
						case pkg.PeriodWeek:
							q.WriteString("week,")
						case pkg.PeriodMonth:
							q.WriteString("month,")
						case pkg.PeriodYear:
							q.WriteString("year,")
						default:
							panic("unknown case for filter period")
						}
					} else {
						switch query.filter.Period {
						case pkg.PeriodWeek:
							q.WriteString(fmt.Sprintf("toStartOfWeek(%s, %d) week,", withTz, query.filter.WeekdayMode))
						case pkg.PeriodMonth:
							q.WriteString(fmt.Sprintf("toStartOfMonth(%s) month,", withTz))
						case pkg.PeriodYear:
							q.WriteString(fmt.Sprintf("toStartOfYear(%s) year,", withTz))
						default:
							panic("unknown case for filter period")
						}
					}
				} else {
					q.WriteString(fmt.Sprintf("%s %s,", withTz, query.fields[i].Name))
				}
			} else if query.from != sessions && (query.fields[i] == FieldPlatformDesktop || query.fields[i] == FieldPlatformMobile || query.fields[i] == FieldPlatformUnknown) {
				q.WriteString(query.selectPlatform(query.fields[i]))
				combineResults = true
			} else if query.fields[i] == FieldEventMetaValues {
				if len(query.filter.EventMetaKey) > 0 {
					query.args = append(query.args, query.filter.EventMetaKey[0])
					q.WriteString(fmt.Sprintf("%s %s,", query.selectField(query.fields[i]), query.fields[i].Name))
				}
			} else if query.fields[i] == FieldTagValue {
				if len(query.filter.Tag) > 0 {
					query.args = append(query.args, query.filter.Tag[0])
					q.WriteString(fmt.Sprintf("%s %s,", query.selectField(query.fields[i]), query.fields[i].Name))
				}
			} else if !includeImported && (query.fields[i] == FieldEventMetaCustomMetricAvg || query.fields[i] == FieldEventMetaCustomMetricTotal) {
				query.args = append(query.args, query.filter.CustomMetricKey)
				fieldCopy := query.fields[i]

				if fieldCopy.sampleType == sampleTypeAuto {
					if query.filter.CustomMetricType == pkg.CustomMetricTypeInteger {
						fieldCopy.sampleType = sampleTypeInt
					} else {
						fieldCopy.sampleType = sampleTypeFloat
					}
				}

				q.WriteString(fmt.Sprintf("%s %s,", fmt.Sprintf(query.selectField(fieldCopy), query.filter.CustomMetricType), fieldCopy.Name))
			} else if query.parent != nil && (query.fields[i] == FieldEntryTitle || query.fields[i] == FieldExitTitle) {
				q.WriteString(query.selectField(query.fields[i]) + ",")
			} else {
				q.WriteString(fmt.Sprintf("%s %s,", query.selectField(query.fields[i]), query.fields[i].Name))
			}
		}

		str := q.String()
		query.q.WriteString(str[:len(str)-1] + " ")
	}

	return combineResults
}

func (query *queryBuilder) selectField(field Field) string {
	includeImported := query.includeImported()
	queryField, sampleFactor := "", ""

	if includeImported {
		queryField = field.queryImported
	} else if query.from == sessions {
		queryField = field.querySessions
	} else if query.from == events && field.queryEvents != "" {
		queryField = field.queryEvents
	} else {
		queryField = field.queryPageViews
	}

	if !includeImported {
		sampleFactor = "*any(_sample_factor)"
	}

	if query.sample > 0 && field.sampleType != 0 {
		if field.sampleType == sampleTypeInt {
			return fmt.Sprintf("toUInt64(greatest(%s%s, 0))", queryField, sampleFactor)
		}

		return fmt.Sprintf("%s%s", queryField, sampleFactor)
	}

	return queryField
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
		sample: query.sample,
	}
	subquery, args := q.query()
	query.args = append(query.args, args...)
	return fmt.Sprintf("toInt64OrDefault((%s)) %s,", subquery, field.Name)
}

func (query *queryBuilder) fromTable() {
	if query.sample > 0 {
		query.q.WriteString(fmt.Sprintf("FROM %s SAMPLE %d ", query.from, query.sample))
	} else {
		query.q.WriteString(fmt.Sprintf("FROM %s ", query.from))
	}

	if query.from == sessions && query.final {
		query.q.WriteString("FINAL ")
	}
}

func (query *queryBuilder) joinQuery() {
	if query.join != nil {
		q, args := query.join.query()
		query.args = append(query.args, args...)
		query.q.WriteString(fmt.Sprintf("JOIN (%s) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id ", q))

		if query.filter.fieldsContain(query.join.groupBy, FieldHour) {
			query.q.WriteString("AND j.hour = hour ")
		} else if query.filter.fieldsContain(query.join.groupBy, FieldMinute) {
			query.q.WriteString("AND j.minute = minute ")
		}
	}

	if query.joinSecond != nil {
		q, args := query.joinSecond.query()
		query.args = append(query.args, args...)
		query.q.WriteString(fmt.Sprintf("JOIN (%s) k ON k.visitor_id = t.visitor_id AND k.session_id = t.session_id ", q))

		if query.filter.fieldsContain(query.joinSecond.groupBy, FieldHour) {
			query.q.WriteString("AND k.hour = hour ")
		} else if query.filter.fieldsContain(query.joinSecond.groupBy, FieldMinute) {
			query.q.WriteString("AND k.minute = minute ")
		}
	}

	if query.leftJoin != nil {
		q, args := query.leftJoin.query()
		query.args = append(query.args, args...)
		query.q.WriteString(fmt.Sprintf("LEFT JOIN (%s) l ON l.visitor_id = t.visitor_id AND l.session_id = t.session_id ", q))
	}

	if query.joinThird != nil {
		q, args := query.joinThird.query()
		query.args = append(query.args, args...)

		if query.filter.fieldsContain(query.groupBy, FieldHour) {
			query.q.WriteString(fmt.Sprintf("JOIN (%s) uvd ON hour = uvd.hour ", q))
		} else if query.filter.fieldsContain(query.groupBy, FieldMinute) {
			query.q.WriteString(fmt.Sprintf("JOIN (%s) uvd ON minute = uvd.minute ", q))
		} else if query.filter.Period == pkg.PeriodDay {
			query.q.WriteString(fmt.Sprintf("JOIN (%s) uvd ON day = uvd.day ", q))
		} else if query.filter.Period == pkg.PeriodWeek {
			query.q.WriteString(fmt.Sprintf("JOIN (%s) uvd ON week = uvd.week ", q))
		} else if query.filter.Period == pkg.PeriodMonth {
			query.q.WriteString(fmt.Sprintf("JOIN (%s) uvd ON month = uvd.month ", q))
		} else {
			query.q.WriteString(fmt.Sprintf("JOIN (%s) uvd ON year = uvd.year ", q))
		}
	}

	if query.joinStep > 1 {
		query.q.WriteString(fmt.Sprintf("JOIN step%d s ON t.visitor_id = s.visitor_id AND t.session_id = s.session_id ", query.joinStep-1))
	}
}

func (query *queryBuilder) unionImported() {
	// ensure we return at least one column for the main table so we can join the imported statistics
	if query.joinImportedSum(query.fieldsImported[0]) {
		query.q.WriteString("UNION ALL (SELECT ")

		for i, field := range query.fields {
			query.q.WriteString("0 ")
			query.q.WriteString(field.Name)

			if i < len(query.fields)-1 {
				query.q.WriteString(",")
			}
		}

		query.q.WriteString(")")
	}
}

func (query *queryBuilder) joinImported(from string) {
	fields := make([]string, 0, len(query.fieldsImported))

	for _, field := range query.fieldsImported {
		if field.subqueryImported != "" {
			if query.filter.Period != pkg.PeriodDay && field == FieldDay {
				switch query.filter.Period {
				case pkg.PeriodWeek:
					fields = append(fields, fmt.Sprintf("toStartOfWeek(date, %d) week", query.filter.WeekdayMode))
				case pkg.PeriodMonth:
					fields = append(fields, "toStartOfMonth(date) month")
				case pkg.PeriodYear:
					fields = append(fields, "toStartOfYear(date) year")
				default:
					panic("unknown case for filter period")
				}
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", field.subqueryImported, field.Name))
			}
		} else {
			fields = append(fields, field.Name)
		}
	}

	dateQuery := query.whereTimeImported()
	joinField := query.fieldsImported[0]
	query.where = make([]where, 0)
	query.whereFieldImported(FieldHostname.Name, query.filter.Hostname, joinField.Name)
	query.whereFieldImported(FieldEntryPath.Name, query.filter.EntryPath, joinField.Name)
	query.whereFieldImported(FieldExitPath.Name, query.filter.ExitPath, joinField.Name)
	query.whereFieldImported(FieldPath.Name, query.filter.Path, joinField.Name)
	query.whereFieldImported(FieldLanguage.Name, query.filter.Language, joinField.Name)
	query.whereFieldImported(FieldCountry.Name, query.filter.Country, joinField.Name)
	query.whereFieldImported(FieldRegion.Name, query.filter.Region, joinField.Name)
	query.whereFieldImported(FieldCity.Name, query.filter.City, joinField.Name)
	query.whereFieldImported(FieldReferrer.Name, query.filter.Referrer, joinField.Name)
	query.whereFieldImported(FieldReferrerName.Name, query.filter.Referrer, joinField.Name)
	query.whereFieldImported(FieldReferrer.Name, query.filter.ReferrerName, joinField.Name)
	query.whereFieldImported(FieldReferrerName.Name, query.filter.ReferrerName, joinField.Name)
	query.whereFieldImported(FieldOS.Name, query.filter.OS, joinField.Name)
	query.whereFieldImported(FieldBrowser.Name, query.filter.Browser, joinField.Name)
	query.whereFieldImported(FieldUTMSource.Name, query.filter.UTMSource, joinField.Name)
	query.whereFieldImported(FieldUTMMedium.Name, query.filter.UTMMedium, joinField.Name)
	query.whereFieldImported(FieldUTMCampaign.Name, query.filter.UTMCampaign, joinField.Name)

	if joinField == FieldPlatformDesktop ||
		joinField == FieldPlatformMobile ||
		joinField == FieldPlatformUnknown {
		query.whereFieldPlatformImported()
	}

	if joinField == FieldPath {
		query.whereFieldPathPattern()
	}

	for _, s := range query.search {
		if s.Field == joinField ||
			(s.Field == FieldReferrer && joinField == FieldReferrerName) ||
			(s.Field == FieldReferrerName && joinField == FieldReferrer) {
			query.whereFieldSearch(joinField.Name, s.Input)
			break
		}
	}

	query.q.WriteString(fmt.Sprintf(`FULL JOIN (SELECT %s FROM "%s" %s `, strings.Join(fields, ","), from, dateQuery))
	query.whereWrite()

	if joinField != FieldPlatformDesktop &&
		joinField != FieldPlatformMobile &&
		joinField != FieldPlatformUnknown &&
		!query.joinImportedSum(joinField) {
		groupBy := query.groupBy
		query.groupBy = []Field{joinField}
		query.groupByFields(false)
		query.groupBy = groupBy
	}

	if query.filter.Period != pkg.PeriodDay && joinField == FieldDay {
		switch query.filter.Period {
		case pkg.PeriodWeek:
			query.q.WriteString(") imp ON t.week = imp.week ")
		case pkg.PeriodMonth:
			query.q.WriteString(") imp ON t.month = imp.month ")
		case pkg.PeriodYear:
			query.q.WriteString(") imp ON t.year = imp.year ")
		default:
			panic("unknown case for filter period")
		}
	} else {
		if joinField == FieldReferrerName {
			query.q.WriteString(fmt.Sprintf(") imp ON t.%s = imp.%s ", FieldReferrer.Name, joinField.Name))
		} else {
			query.q.WriteString(fmt.Sprintf(") imp ON t.%s = imp.%s ", joinField.Name, joinField.Name))
		}
	}
}

func (query *queryBuilder) joinImportedSum(field Field) bool {
	return field == FieldVisitors ||
		field == FieldSessions ||
		field == FieldViews ||
		field == FieldBounces
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

	return q.String()
}

func (query *queryBuilder) whereTimeImported() string {
	query.args = append(query.args, query.filter.ClientID)
	var q strings.Builder
	q.WriteString("WHERE client_id = ? ")
	tz := query.filter.Timezone.String()

	if !query.filter.importedFrom.IsZero() && !query.filter.importedTo.IsZero() && query.filter.importedFrom.Equal(query.filter.importedTo) {
		query.args = append(query.args, query.filter.importedFrom.Format(dateFormat))
		q.WriteString(fmt.Sprintf("AND toDate(date, '%s') = toDate(?) ", tz))
	} else {
		if !query.filter.importedFrom.IsZero() {
			if query.filter.IncludeTime {
				query.args = append(query.args, query.filter.importedFrom)
				q.WriteString(fmt.Sprintf("AND toDateTime(date, '%s') >= toDateTime(?, '%s') ", tz, tz))
			} else {
				query.args = append(query.args, query.filter.importedFrom.Format(dateFormat))
				q.WriteString(fmt.Sprintf("AND toDate(date, '%s') >= toDate(?) ", tz))
			}
		}

		if !query.filter.importedTo.IsZero() {
			if query.filter.IncludeTime {
				query.args = append(query.args, query.filter.importedTo)
				q.WriteString(fmt.Sprintf("AND toDateTime(date, '%s') <= toDateTime(?, '%s') ", tz, tz))
			} else {
				query.args = append(query.args, query.filter.importedTo.Format(dateFormat))
				q.WriteString(fmt.Sprintf("AND toDate(date, '%s') <= toDate(?) ", tz))
			}
		}
	}

	return q.String()
}

func (query *queryBuilder) whereFields() {
	if query.from == sessions {
		query.whereField(FieldEntryPath.Name, query.filter.EntryPath)
		query.whereField(FieldExitPath.Name, query.filter.ExitPath)
	} else {
		query.whereField(FieldPath.Name, query.filter.Path)
		query.whereField(FieldTagKeysRaw.Name, query.filter.Tag)
		query.whereFieldPathPattern()
		query.whereFieldPathIn()
		query.whereFieldTag()
	}

	if query.from == events || query.includeEventFilter {
		query.whereField(FieldPath.Name, query.filter.Path)
		query.whereField(FieldEventName.Name, query.filter.EventName)
		query.whereField(FieldEventMetaKeysRaw.Name, query.filter.EventMetaKey)
		query.whereFieldMeta()
	}

	query.whereField(FieldHostname.Name, query.filter.Hostname)
	query.whereField(FieldLanguage.Name, query.filter.Language)
	query.whereField(FieldCountry.Name, query.filter.Country)
	query.whereField(FieldRegion.Name, query.filter.Region)
	query.whereField(FieldCity.Name, query.filter.City)
	query.whereField(FieldReferrer.Name, query.filter.Referrer)
	query.whereField(FieldReferrerName.Name, query.filter.ReferrerName)
	query.whereField(FieldChannel.Name, query.filter.Channel)
	query.whereField(FieldOS.Name, query.filter.OS)
	query.whereField(FieldOSVersion.Name, query.filter.OSVersion)
	query.whereField(FieldBrowser.Name, query.filter.Browser)
	query.whereField(FieldBrowserVersion.Name, query.filter.BrowserVersion)
	query.whereField(FieldScreenClass.Name, query.filter.ScreenClass)
	query.whereField(FieldUTMSource.Name, query.filter.UTMSource)
	query.whereField(FieldUTMMedium.Name, query.filter.UTMMedium)
	query.whereField(FieldUTMCampaign.Name, query.filter.UTMCampaign)
	query.whereField(FieldUTMContent.Name, query.filter.UTMContent)
	query.whereField(FieldUTMTerm.Name, query.filter.UTMTerm)
	query.whereFieldPlatform()
	query.whereFieldVisitorSessionID()

	for i := range query.search {
		query.whereFieldSearch(query.search[i].Field.Name, query.search[i].Input)
	}

	query.whereWrite()
}

func (query *queryBuilder) whereWrite() {
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
	if len(value) > 0 {
		if query.from == events && field == FieldTagKeysRaw.Name {
			field = FieldEventMetaKeysRaw.Name
		}

		var group where
		eqContainsArgs := make([]any, 0, len(value))
		notEqArgs := make([]any, 0, len(value))

		for _, v := range value {
			comparator := "%s = ? "
			not := strings.HasPrefix(v, "!")

			if field == FieldEventMetaKeysRaw.Name || field == FieldTagKeysRaw.Name {
				if not {
					v = v[1:]
					comparator = "has(%s, ?) = 0 "
				} else {
					comparator = "has(%s, ?) = 1 "
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
			} else if strings.HasPrefix(v, "^") {
				if field == FieldLanguage.Name || field == FieldCountry.Name {
					v = v[1:]
					comparator = "has(splitByChar(',', ?), %s) != 1 "
				} else {
					v = fmt.Sprintf("%%%s%%", v[1:])
					comparator = "ilike(%s, ?) != 1 "
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

func (query *queryBuilder) whereFieldImported(field string, value []string, joinField string) {
	if joinField == field {
		query.whereField(field, value)
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
	if len(value) > 0 {
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

func (query *queryBuilder) whereFieldTag() {
	if len(query.filter.Tags) > 0 {
		values, keys := FieldTagValuesRaw.Name, FieldTagKeysRaw.Name

		if query.from == events {
			values, keys = FieldEventMetaValuesRaw.Name, FieldEventMetaKeysRaw.Name
		}

		var group where

		for k, v := range query.filter.Tags {
			comparator := "%s[indexOf(%s, ?)] = ? "

			if strings.HasPrefix(v, "!") {
				v = v[1:]
				comparator = "%s[indexOf(%s, ?)] != ? "
			} else if strings.HasPrefix(v, "~") {
				v = fmt.Sprintf("%%%s%%", v[1:])
				comparator = "ilike(%s[indexOf(%s, ?)], ?) = 1 "
			} else if strings.HasPrefix(v, "^") {
				v = fmt.Sprintf("%%%s%%", v[1:])
				comparator = "ilike(%s[indexOf(%s, ?)], ?) != 1 "
			}

			// use notEq because they will all be joined using AND
			query.args = append(query.args, k, query.nullValue(v))
			group.notEq = append(group.notEq, fmt.Sprintf(comparator, values, keys))
		}

		query.where = append(query.where, group)
	}
}

func (query *queryBuilder) whereFieldMeta() {
	if len(query.filter.EventMeta) > 0 {
		var group where

		for k, v := range query.filter.EventMeta {
			comparator := "event_meta_values[indexOf(event_meta_keys, ?)] = ? "

			if strings.HasPrefix(v, "!") {
				v = v[1:]
				comparator = "event_meta_values[indexOf(event_meta_keys, ?)] != ? "
			} else if strings.HasPrefix(v, "~") {
				v = fmt.Sprintf("%%%s%%", v[1:])
				comparator = "ilike(event_meta_values[indexOf(event_meta_keys, ?)], ?) = 1 "
			} else if strings.HasPrefix(v, "^") {
				v = fmt.Sprintf("%%%s%%", v[1:])
				comparator = "ilike(event_meta_values[indexOf(event_meta_keys, ?)], ?) != 1 "
			}

			// use notEq because they will all be joined using AND
			query.args = append(query.args, k, query.nullValue(v))
			group.notEq = append(group.notEq, comparator)
		}

		query.where = append(query.where, group)
	}
}

func (query *queryBuilder) whereFieldPlatform() {
	if query.filter.Platform != "" {
		if strings.HasPrefix(query.filter.Platform, "!") {
			platform := query.filter.Platform[1:]

			if platform == pkg.PlatformDesktop {
				query.where = append(query.where, where{notEq: []string{"desktop != 1 "}})
			} else if platform == pkg.PlatformMobile {
				query.where = append(query.where, where{notEq: []string{"mobile != 1 "}})
			} else {
				query.where = append(query.where, where{notEq: []string{"(desktop = 1 OR mobile = 1) "}})
			}
		} else {
			if query.filter.Platform == pkg.PlatformDesktop {
				query.where = append(query.where, where{eqContains: []string{"desktop = 1 "}})
			} else if query.filter.Platform == pkg.PlatformMobile {
				query.where = append(query.where, where{eqContains: []string{"mobile = 1 "}})
			} else {
				query.where = append(query.where, where{eqContains: []string{"desktop = 0 AND mobile = 0 "}})
			}
		}
	}
}

func (query *queryBuilder) whereFieldPlatformImported() {
	if query.filter.Platform != "" {
		if strings.HasPrefix(query.filter.Platform, "!") {
			platform := query.filter.Platform[1:]

			if platform == pkg.PlatformDesktop {
				query.where = append(query.where, where{notEq: []string{
					"lower(category) != 'desktop' ",
					"lower(category) != 'laptop' ",
				}})
			} else if platform == pkg.PlatformMobile {
				query.where = append(query.where, where{notEq: []string{
					"lower(category) != 'mobile' ",
					"lower(category) != 'phone' ",
					"lower(category) != 'tablet' ",
				}})
			} else {
				query.where = append(query.where, where{notEq: []string{"category != '' "}})
			}
		} else {
			if query.filter.Platform == pkg.PlatformDesktop {
				query.where = append(query.where, where{notEq: []string{
					"(lower(category) = 'desktop' OR lower(category) = 'laptop') ",
				}})
			} else if query.filter.Platform == pkg.PlatformMobile {
				query.where = append(query.where, where{notEq: []string{
					"(lower(category) = 'mobile' OR lower(category) = 'phone' OR lower(category) = 'tablet') ",
				}})
			} else {
				query.where = append(query.where, where{notEq: []string{"category = '' "}})
			}
		}
	}
}

func (query *queryBuilder) whereFieldVisitorSessionID() {
	if query.filter.VisitorID != 0 && query.filter.SessionID != 0 {
		query.where = append(query.where, where{
			eqContains: []string{"t.visitor_id = ? "},
		}, where{
			eqContains: []string{"t.session_id = ? "},
		})
		query.args = append(query.args, query.filter.VisitorID, query.filter.SessionID)
	}
}

func (query *queryBuilder) whereFieldPathPattern() {
	if len(query.filter.PathPattern) > 0 {
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
	if len(query.filter.AnyPath) > 0 {
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

func (query *queryBuilder) groupByFields(imported bool) {
	if len(query.groupBy) > 0 {
		query.q.WriteString("GROUP BY ")
		var q strings.Builder

		for i := range query.groupBy {
			if imported && query.groupBy[i] == FieldAnyReferrerImported {
				continue
			} else if query.filter.Period != pkg.PeriodDay && query.groupBy[i] == FieldDay {
				switch query.filter.Period {
				case pkg.PeriodWeek:
					q.WriteString("week,")
				case pkg.PeriodMonth:
					q.WriteString("month,")
				case pkg.PeriodYear:
					q.WriteString("year,")
				default:
					panic("unknown case for filter period")
				}
			} else if query.parent != nil && (query.groupBy[i] == FieldEntryTitle || query.groupBy[i] == FieldExitTitle) {
				q.WriteString(query.selectField(query.groupBy[i]) + ",")
			} else if query.groupBy[i] == FieldVisitorID || query.groupBy[i] == FieldSessionID {
				q.WriteString("t." + query.groupBy[i].Name + ",")
			} else {
				q.WriteString(query.groupBy[i].Name + ",")
			}
		}

		str := q.String()
		query.q.WriteString(str[:len(str)-1] + " ")
	}
}

func (query *queryBuilder) having() {
	if query.from == sessions && query.joinStep == 0 {
		query.q.WriteString("HAVING sum(sign) > 0 ")
	}
}

func (query *queryBuilder) orderByFields() {
	if len(query.filter.Sort) > 0 {
		fields := make([]Field, 0, len(query.filter.Sort))

		for i := range query.filter.Sort {
			query.filter.Sort[i].Field.queryDirection = query.filter.Sort[i].Direction
			fields = append(fields, query.filter.Sort[i].Field)
		}

		query.orderBy = fields
	}

	if len(query.orderBy) > 0 {
		query.q.WriteString("ORDER BY ")
		var q strings.Builder

		for i := range query.orderBy {
			if query.orderBy[i].queryWithFill != "" {
				q.WriteString(fmt.Sprintf("%s %s %s,", query.orderBy[i].Name, query.orderBy[i].queryDirection, query.orderBy[i].queryWithFill))
			} else if query.orderBy[i].withFill {
				fillQuery := query.withFill()
				name := query.orderBy[i].Name

				if query.filter.Period != pkg.PeriodDay && query.orderBy[i] == FieldDay {
					switch query.filter.Period {
					case pkg.PeriodWeek:
						name = "week"
					case pkg.PeriodMonth:
						name = "month"
					case pkg.PeriodYear:
						name = "year"
					default:
						panic("unknown case for filter period")
					}
				}

				q.WriteString(fmt.Sprintf("%s %s %s,", name, query.orderBy[i].queryDirection, fillQuery))
			} else if query.orderBy[i].Name == FieldCity.Name {
				q.WriteString(fmt.Sprintf("normalizeUTF8NFKD(%s) %s,", query.orderBy[i].Name, query.orderBy[i].queryDirection))
			} else {
				q.WriteString(fmt.Sprintf("%s %s,", query.orderBy[i].Name, query.orderBy[i].queryDirection))
			}
		}

		str := q.String()
		query.q.WriteString(str[:len(str)-1] + " ")
	}
}

func (query *queryBuilder) withFill() string {
	from := query.filter.From
	to := query.filter.To

	if !query.filter.importedFrom.IsZero() && query.filter.importedFrom.Before(from) {
		from = query.filter.importedFrom
	}

	if !from.IsZero() && !to.IsZero() {
		q := ""

		switch query.filter.Period {
		case pkg.PeriodDay:
			q = "WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY "
		case pkg.PeriodWeek:
			q = fmt.Sprintf("WITH FILL FROM toStartOfWeek(toDate(?), %d) TO toDate(?)+1 STEP INTERVAL 1 WEEK ", query.filter.WeekdayMode)
		case pkg.PeriodMonth:
			q = "WITH FILL FROM toStartOfMonth(toDate(?)) TO toDate(?)+1 STEP INTERVAL 1 MONTH "
		case pkg.PeriodYear:
			q = "WITH FILL FROM toStartOfYear(toDate(?)) TO toDate(?)+1 STEP INTERVAL 1 YEAR "
		}

		query.args = append(query.args, from.Format(dateFormat), to.Format(dateFormat))
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

func (query *queryBuilder) includeImported() bool {
	return query.filter != nil && !query.filter.ImportedUntil.IsZero() && query.fromImported != ""
}
