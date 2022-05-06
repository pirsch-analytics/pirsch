package pirsch

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// PeriodDay groups the results by date.
	PeriodDay = Period(iota)

	// PeriodWeek groups the results by week.
	PeriodWeek

	// PeriodMonth groups the results by month.
	PeriodMonth

	// PeriodYear groups the result by year.
	PeriodYear

	// PlatformDesktop filters for everything on desktops.
	PlatformDesktop = "desktop"

	// PlatformMobile filters for everything on mobile devices.
	PlatformMobile = "mobile"

	// PlatformUnknown filters for everything where the platform is unspecified.
	PlatformUnknown = "unknown"

	// Unknown filters for an unknown (empty) value.
	// This is a synonym for "null".
	Unknown = "null"

	// DirectionASC sorts results in ascending order.
	DirectionASC = "ASC"

	// DirectionDESC sorts results in descending order.
	DirectionDESC = "DESC"
)

// Period is used to group results.
type Period int

// Direction is used to sort results.
type Direction string

// NullClient is a placeholder for no client (0).
var NullClient = int64(0)

// Filter are all fields that can be used to filter the result sets.
// Fields can be inverted by adding a "!" in front of the string.
// To compare to none/unknown/empty, set the value to "null" (case-insensitive).
type Filter struct {
	// ClientID is the optional.
	ClientID int64

	// Timezone sets the timezone used to interpret dates and times.
	// It will be set to UTC by default.
	Timezone *time.Location

	// From is the start date of the selected period.
	From time.Time

	// To is the end date of the selected period.
	To time.Time

	// IncludeTime sets whether the selected period should contain the time (hour, minute, second).
	IncludeTime bool

	// Period sets the period to group results.
	// This is only used by Analyzer.Visitors, Analyzer.AvgSessionDuration, and Analyzer.AvgTimeOnPage.
	// Using it for other queries leads to wrong results and might return an error.
	// This can either be PeriodDay (default), PeriodWeek, or PeriodYear.
	Period Period

	// Path filters for the path.
	// Note that if this and PathPattern are both set, Path will be preferred.
	Path string

	// EntryPath filters for the entry page.
	EntryPath string

	// ExitPath filters for the exit page.
	ExitPath string

	// PathPattern filters for the path using a (ClickHouse supported) regex pattern.
	// Note that if this and Path are both set, Path will be preferred.
	// Examples for useful patterns (all case-insensitive, * is used for every character but slashes, ** is used for all characters including slashes):
	//  (?i)^/path/[^/]+$ // matches /path/*
	//  (?i)^/path/[^/]+/.* // matches /path/*/**
	//  (?i)^/path/[^/]+/slashes$ // matches /path/*/slashes
	//  (?i)^/path/.+/slashes$ // matches /path/**/slashes
	PathPattern string

	// Language filters for the ISO language code.
	Language string

	// Country filters for the ISO country code.
	Country string

	// City filters for the city name.
	City string

	// Referrer filters for the full referrer.
	Referrer string

	// ReferrerName filters for the referrer name.
	ReferrerName string

	// OS filters for the operating system.
	OS string

	// OSVersion filters for the operating system version.
	OSVersion string

	// Browser filters for the browser.
	Browser string

	// BrowserVersion filters for the browser version.
	BrowserVersion string

	// Platform filters for the platform (desktop, mobile, unknown).
	Platform string

	// ScreenClass filters for the screen class.
	ScreenClass string

	// ScreenWidth filters for the screen width.
	ScreenWidth string

	// ScreenHeight filters for the screen width.
	ScreenHeight string

	// UTMSource filters for the utm_source query parameter.
	UTMSource string

	// UTMMedium filters for the utm_medium query parameter.
	UTMMedium string

	// UTMCampaign filters for the utm_campaign query parameter.
	UTMCampaign string

	// UTMContent filters for the utm_content query parameter.
	UTMContent string

	// UTMTerm filters for the utm_term query parameter.
	UTMTerm string

	// EventName filters for an event by its name.
	EventName string

	// EventMetaKey filters for an event meta key.
	// This must be used together with an EventName.
	EventMetaKey string

	// EventMeta filters for event metadata.
	EventMeta map[string]string

	// OrderBy sorts the results.
	// This will overwrite the default order provided by the Analyzer.
	OrderBy []OrderBy

	// Offset limits the number of results. Less or equal to zero means no offset.
	Offset int

	// Limit limits the number of results. Less or equal to zero means no limit.
	Limit int

	// IncludeTitle indicates that the Analyzer.Pages, Analyzer.EntryPages, and Analyzer.ExitPages should contain the page title.
	IncludeTitle bool

	// IncludeTimeOnPage indicates that the Analyzer.Pages and Analyzer.EntryPages should contain the average time on page.
	IncludeTimeOnPage bool

	// MaxTimeOnPageSeconds is an optional maximum for the time spent on page.
	// Visitors who are idle artificially increase the average time spent on a page, this option can be used to limit the effect.
	// Set to 0 to disable this option (default).
	MaxTimeOnPageSeconds int

	eventFilter bool
	minIsBot    uint8
}

// OrderBy sorts a results by a field and direction.
type OrderBy struct {
	Field     Field
	Direction Direction
}

// NewFilter creates a new filter for given client ID.
func NewFilter(clientID int64) *Filter {
	return &Filter{
		ClientID: clientID,
	}
}

func (filter *Filter) validate() {
	if filter.Timezone == nil {
		filter.Timezone = time.UTC
	}

	if !filter.From.IsZero() {
		if filter.IncludeTime {
			filter.From = filter.From.In(time.UTC)
		} else {
			filter.From = filter.toDate(filter.From)
		}
	}

	if !filter.To.IsZero() {
		if filter.IncludeTime {
			filter.To = filter.To.In(time.UTC)
		} else {
			filter.To = filter.toDate(filter.To)
		}
	}

	if !filter.To.IsZero() && filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}

	// use tomorrow instead of limiting to "today", so that all timezones are included
	tomorrow := Today().Add(time.Hour * 24)

	if !filter.To.IsZero() && filter.To.After(tomorrow) {
		filter.To = tomorrow
	}

	if filter.Path != "" && filter.PathPattern != "" {
		filter.PathPattern = ""
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if filter.Limit < 0 {
		filter.Limit = 0
	}
}

func (filter *Filter) buildQuery(fields, groupBy, orderBy []Field) ([]any, string) {
	table := filter.table()
	args := make([]any, 0)
	var query strings.Builder

	if filter.Period != PeriodDay && filter.fieldsContain(fields, FieldDay.name) {
		query.WriteString(`SELECT `)

		for i := range fields {
			if fields[i] != FieldDay {
				query.WriteString(fmt.Sprintf(`%s %s,`, fields[i].queryPeriod, fields[i].name))
			}
		}

		switch filter.Period {
		case PeriodWeek:
			query.WriteString(`toStartOfWeek(day) week FROM (`)
		case PeriodMonth:
			query.WriteString(`toStartOfMonth(day) month FROM (`)
		case PeriodYear:
			query.WriteString(`toStartOfYear(day) year FROM (`)
		}
	}

	if table == "event" || filter.Path != "" || filter.PathPattern != "" || filter.fieldsContain(fields, FieldPath.name) {
		if table == "session" {
			table = "page_view"
		}

		query.WriteString(fmt.Sprintf(`SELECT %s FROM %s v `, filter.joinPageViewFields(&args, fields), table))

		if filter.minIsBot > 0 ||
			filter.EntryPath != "" ||
			filter.ExitPath != "" ||
			filter.fieldsContain(fields, FieldBounces.name) ||
			filter.fieldsContain(fields, FieldViews.name) ||
			filter.fieldsContain(fields, FieldEntryPath.name) ||
			filter.fieldsContain(fields, FieldExitPath.name) {
			filterArgs, filterQuery := filter.joinSessions(table, fields)
			args = append(args, filterArgs...)
			query.WriteString(filterQuery)
		}

		filter.EntryPath, filter.ExitPath = "", ""
		filterArgs, filterQuery := filter.query(false)
		args = append(args, filterArgs...)
		query.WriteString(fmt.Sprintf(`WHERE %s `, filterQuery))

		if len(groupBy) > 0 {
			query.WriteString(fmt.Sprintf(`GROUP BY %s `, filter.joinGroupBy(groupBy)))
		}

		if len(orderBy) > 0 {
			query.WriteString(fmt.Sprintf(`ORDER BY %s `, filter.joinOrderBy(&args, orderBy)))
		}
	} else {
		filterArgs, filterQuery := filter.query(true)
		query.WriteString(fmt.Sprintf(`SELECT %s
			FROM session
			WHERE %s `, filter.joinSessionFields(&args, fields), filterQuery))
		args = append(args, filterArgs...)

		if len(groupBy) > 0 {
			query.WriteString(fmt.Sprintf(`GROUP BY %s `, filter.joinGroupBy(groupBy)))
		}

		query.WriteString(`HAVING sum(sign) > 0 `)

		if len(orderBy) > 0 {
			query.WriteString(fmt.Sprintf(`ORDER BY %s `, filter.joinOrderBy(&args, orderBy)))
		}
	}

	if filter.Period != PeriodDay && filter.fieldsContain(fields, FieldDay.name) {
		switch filter.Period {
		case PeriodWeek:
			query.WriteString(`) GROUP BY week ORDER BY week ASC`)
		case PeriodMonth:
			query.WriteString(`) GROUP BY month ORDER BY month ASC`)
		case PeriodYear:
			query.WriteString(`) GROUP BY year ORDER BY year ASC`)
		}
	}

	query.WriteString(filter.withLimit())
	return args, query.String()
}

func (filter *Filter) joinPageViewFields(args *[]any, fields []Field) string {
	var out strings.Builder

	for i := range fields {
		if fields[i].filterTime {
			timeArgs, timeQuery := filter.queryTime(false)
			*args = append(*args, timeArgs...)
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].queryPageViews, timeQuery), fields[i].name))
		} else if fields[i].timezone {
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].queryPageViews, filter.Timezone.String()), fields[i].name))
		} else if fields[i].name == "meta_value" {
			*args = append(*args, filter.EventMetaKey)
			out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].queryPageViews, fields[i].name))
		} else {
			out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].queryPageViews, fields[i].name))
		}
	}

	str := out.String()
	return str[:len(str)-1]
}

func (filter *Filter) joinSessionFields(args *[]any, fields []Field) string {
	var out strings.Builder

	for i := range fields {
		if fields[i].filterTime {
			timeArgs, timeQuery := filter.queryTime(false)
			*args = append(*args, timeArgs...)
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].queryPageViews, timeQuery), fields[i].name))
		} else if fields[i].timezone {
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].querySessions, filter.Timezone.String()), fields[i].name))
		} else {
			out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].querySessions, fields[i].name))
		}
	}

	str := out.String()
	return str[:len(str)-1]
}

func (filter *Filter) joinSessions(table string, fields []Field) ([]any, string) {
	path, pathPattern, eventName, eventMetaKey, eventMeta := filter.Path, filter.PathPattern, filter.EventName, filter.EventMetaKey, filter.EventMeta
	filter.Path, filter.PathPattern, filter.EventName, filter.EventMetaKey, filter.EventMeta = "", "", "", "", nil
	filterArgs, filterQuery := filter.query(true)
	filter.Path, filter.PathPattern, filter.EventName, filter.EventMetaKey, filter.EventMeta = path, pathPattern, eventName, eventMetaKey, eventMeta
	sessionFields := make([]string, 0, 6)
	groupBy := make([]string, 0, 4)

	if filter.fieldsContain(fields, FieldEntryPath.name) {
		sessionFields = append(sessionFields, FieldEntryPath.name)
		groupBy = append(groupBy, FieldEntryPath.name)
	}

	if filter.fieldsContain(fields, FieldExitPath.name) {
		sessionFields = append(sessionFields, FieldExitPath.name)
		groupBy = append(groupBy, FieldExitPath.name)
	}

	if filter.fieldsContainByQuerySession(fields, FieldEntryTitle.querySessions) {
		sessionFields = append(sessionFields, FieldEntryTitle.querySessions)
		groupBy = append(groupBy, FieldEntryTitle.querySessions)
	}

	if filter.fieldsContainByQuerySession(fields, FieldExitTitle.querySessions) {
		sessionFields = append(sessionFields, FieldExitTitle.querySessions)
		groupBy = append(groupBy, FieldExitTitle.querySessions)
	}

	if filter.fieldsContain(fields, FieldBounces.name) {
		sessionFields = append(sessionFields, "sum(is_bounce*sign) is_bounce")
	}

	if filter.fieldsContain(fields, FieldViews.name) {
		sessionFields = append(sessionFields, "sum(page_views*sign) page_views")
	}

	sessionFieldsQuery := strings.Join(sessionFields, ",")

	if sessionFieldsQuery != "" {
		sessionFieldsQuery = "," + sessionFieldsQuery
	}

	query := ""

	if table == "page_view" || table == "event" {
		query = "INNER "
	} else {
		query = "LEFT "
	}

	groupByQuery := strings.Join(groupBy, ",")

	if groupByQuery != "" {
		groupByQuery = "," + groupByQuery
	}

	query += fmt.Sprintf(`JOIN (
			SELECT visitor_id,
			session_id
			%s
			FROM session
			WHERE %s
			GROUP BY visitor_id, session_id %s
			HAVING sum(sign) > 0
		) s
		ON s.visitor_id = v.visitor_id AND s.session_id = v.session_id `, sessionFieldsQuery, filterQuery, groupByQuery)
	return filterArgs, query
}

func (filter *Filter) joinGroupBy(fields []Field) string {
	var out strings.Builder

	for i := range fields {
		out.WriteString(fields[i].name + ",")
	}

	str := out.String()
	return str[:len(str)-1]
}

func (filter *Filter) joinOrderBy(args *[]any, fields []Field) string {
	if len(filter.OrderBy) > 0 {
		fields = make([]Field, 0, len(filter.OrderBy))

		for i := range filter.OrderBy {
			filter.OrderBy[i].Field.queryDirection = string(filter.OrderBy[i].Direction)
			fields = append(fields, filter.OrderBy[i].Field)
		}
	}

	var out strings.Builder

	for i := range fields {
		if fields[i].queryWithFill != "" {
			out.WriteString(fmt.Sprintf(`%s %s %s,`, fields[i].name, fields[i].queryDirection, fields[i].queryWithFill))
		} else if fields[i].withFill {
			fillArgs, fillQuery := filter.withFill()
			*args = append(*args, fillArgs...)
			out.WriteString(fmt.Sprintf(`%s %s %s,`, fields[i].name, fields[i].queryDirection, fillQuery))
		} else {
			out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].name, fields[i].queryDirection))
		}
	}

	str := out.String()
	return str[:len(str)-1]
}

func (filter *Filter) table() string {
	if filter.EventName != "" || filter.eventFilter {
		return "event"
	}

	return "session"
}

func (filter *Filter) queryTime(filterBots bool) ([]any, string) {
	args := make([]any, 0, 5)
	args = append(args, filter.ClientID)
	var sqlQuery strings.Builder
	sqlQuery.WriteString("client_id = ? ")
	tz := filter.Timezone.String()

	if !filter.From.IsZero() && !filter.To.IsZero() && filter.From.Equal(filter.To) {
		args = append(args, filter.From)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') = toDate(?) ", tz))
	} else {
		if !filter.From.IsZero() {
			args = append(args, filter.From)

			if filter.IncludeTime {
				sqlQuery.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') >= toDateTime(?, '%s') ", tz, tz))
			} else {
				sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') >= toDate(?) ", tz))
			}
		}

		if !filter.To.IsZero() {
			args = append(args, filter.To)

			if filter.IncludeTime {
				sqlQuery.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') <= toDateTime(?, '%s') ", tz, tz))
			} else {
				sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') <= toDate(?) ", tz))
			}
		}
	}

	if filterBots && filter.minIsBot > 0 {
		args = append(args, filter.minIsBot)
		sqlQuery.WriteString(" AND is_bot < ? ")
	}

	return args, sqlQuery.String()
}

func (filter *Filter) queryFields() ([]any, string) {
	n := 25 + len(filter.EventMeta) // maximum number of fields + one for bot filter + meta data fields
	args := make([]any, 0, n)
	queryFields := make([]string, 0, n)
	filter.appendQuery(&queryFields, &args, "path", filter.Path)
	filter.appendQuery(&queryFields, &args, "entry_path", filter.EntryPath)
	filter.appendQuery(&queryFields, &args, "exit_path", filter.ExitPath)
	filter.appendQuery(&queryFields, &args, "language", filter.Language)
	filter.appendQuery(&queryFields, &args, "country_code", filter.Country)
	filter.appendQuery(&queryFields, &args, "city", filter.City)
	filter.appendQuery(&queryFields, &args, "referrer", filter.Referrer)
	filter.appendQuery(&queryFields, &args, "referrer_name", filter.ReferrerName)
	filter.appendQuery(&queryFields, &args, "os", filter.OS)
	filter.appendQuery(&queryFields, &args, "os_version", filter.OSVersion)
	filter.appendQuery(&queryFields, &args, "browser", filter.Browser)
	filter.appendQuery(&queryFields, &args, "browser_version", filter.BrowserVersion)
	filter.appendQuery(&queryFields, &args, "screen_class", filter.ScreenClass)
	filter.appendQueryUInt16(&queryFields, &args, "screen_width", filter.ScreenWidth)
	filter.appendQueryUInt16(&queryFields, &args, "screen_height", filter.ScreenHeight)
	filter.appendQuery(&queryFields, &args, "utm_source", filter.UTMSource)
	filter.appendQuery(&queryFields, &args, "utm_medium", filter.UTMMedium)
	filter.appendQuery(&queryFields, &args, "utm_campaign", filter.UTMCampaign)
	filter.appendQuery(&queryFields, &args, "utm_content", filter.UTMContent)
	filter.appendQuery(&queryFields, &args, "utm_term", filter.UTMTerm)
	filter.appendQuery(&queryFields, &args, "event_name", filter.EventName)
	filter.appendQuery(&queryFields, &args, "event_meta_keys", filter.EventMetaKey)
	filter.queryPlatform(&queryFields)
	filter.queryPathPattern(&queryFields, &args)
	filter.appendQueryMeta(&queryFields, &args, filter.EventMeta)
	return args, strings.Join(queryFields, "AND ")
}

func (filter *Filter) queryPlatform(queryFields *[]string) {
	if filter.Platform != "" {
		if strings.HasPrefix(filter.Platform, "!") {
			platform := filter.Platform[1:]

			if platform == PlatformDesktop {
				*queryFields = append(*queryFields, "desktop != 1 ")
			} else if platform == PlatformMobile {
				*queryFields = append(*queryFields, "mobile != 1 ")
			} else {
				*queryFields = append(*queryFields, "(desktop = 1 OR mobile = 1) ")
			}
		} else {
			if filter.Platform == PlatformDesktop {
				*queryFields = append(*queryFields, "desktop = 1 ")
			} else if filter.Platform == PlatformMobile {
				*queryFields = append(*queryFields, "mobile = 1 ")
			} else {
				*queryFields = append(*queryFields, "desktop = 0 AND mobile = 0 ")
			}
		}
	}
}

func (filter Filter) queryPathPattern(queryFields *[]string, args *[]any) {
	if filter.PathPattern != "" {
		if strings.HasPrefix(filter.PathPattern, "!") {
			*args = append(*args, filter.PathPattern[1:])
			*queryFields = append(*queryFields, `match("path", ?) = 0`)
		} else {
			*args = append(*args, filter.PathPattern)
			*queryFields = append(*queryFields, `match("path", ?) = 1`)
		}
	}
}

func (filter *Filter) fields() string {
	fields := make([]string, 0, 26)
	filter.appendField(&fields, "path", filter.Path)
	filter.appendField(&fields, "entry_path", filter.EntryPath)
	filter.appendField(&fields, "exit_path", filter.ExitPath)
	filter.appendField(&fields, "language", filter.Language)
	filter.appendField(&fields, "country_code", filter.Country)
	filter.appendField(&fields, "city", filter.City)
	filter.appendField(&fields, "referrer", filter.Referrer)
	filter.appendField(&fields, "referrer_name", filter.ReferrerName)
	filter.appendField(&fields, "os", filter.OS)
	filter.appendField(&fields, "os_version", filter.OSVersion)
	filter.appendField(&fields, "browser", filter.Browser)
	filter.appendField(&fields, "browser_version", filter.BrowserVersion)
	filter.appendField(&fields, "screen_class", filter.ScreenClass)
	filter.appendField(&fields, "screen_width", filter.ScreenWidth)
	filter.appendField(&fields, "screen_height", filter.ScreenHeight)
	filter.appendField(&fields, "utm_source", filter.UTMSource)
	filter.appendField(&fields, "utm_medium", filter.UTMMedium)
	filter.appendField(&fields, "utm_campaign", filter.UTMCampaign)
	filter.appendField(&fields, "utm_content", filter.UTMContent)
	filter.appendField(&fields, "utm_term", filter.UTMTerm)
	filter.appendField(&fields, "event_name", filter.EventName)

	if filter.Platform != "" {
		platform := filter.Platform

		if strings.HasPrefix(platform, "!") {
			platform = filter.Platform[1:]
		}

		if platform == PlatformDesktop {
			fields = append(fields, "desktop")
		} else if platform == PlatformMobile {
			fields = append(fields, "mobile")
		} else {
			fields = append(fields, "desktop")
			fields = append(fields, "mobile")
		}
	}

	if filter.Path == "" && filter.PathPattern != "" {
		fields = append(fields, "path")
	}

	if len(filter.EventMeta) > 0 {
		fields = append(fields, "event_meta_keys", "event_meta_values")
	} else {
		filter.appendField(&fields, "event_meta_keys", filter.EventMetaKey)
	}

	return strings.Join(fields, ",")
}

func (filter *Filter) appendField(fields *[]string, field, value string) {
	if value != "" {
		*fields = append(*fields, field)
	}
}

func (filter *Filter) fieldsContain(haystack []Field, needle string) bool {
	for i := range haystack {
		if haystack[i].name == needle {
			return true
		}
	}

	return false
}

func (filter *Filter) fieldsContainByQuerySession(haystack []Field, needle string) bool {
	for i := range haystack {
		if haystack[i].querySessions == needle {
			return true
		}
	}

	return false
}

func (filter *Filter) withFill() ([]any, string) {
	if !filter.From.IsZero() && !filter.To.IsZero() {
		tz := filter.Timezone.String()
		query := fmt.Sprintf("WITH FILL FROM toDate(?, '%s') TO toDate(?, '%s')+1 ", tz, tz)
		return []any{filter.From, filter.To}, query
	}

	return nil, ""
}

func (filter *Filter) withLimit() string {
	if filter.Limit > 0 && filter.Offset > 0 {
		return fmt.Sprintf("LIMIT %d OFFSET %d ", filter.Limit, filter.Offset)
	} else if filter.Limit > 0 {
		return fmt.Sprintf("LIMIT %d ", filter.Limit)
	}

	return ""
}

func (filter *Filter) query(filterBots bool) ([]any, string) {
	args, query := filter.queryTime(false)
	fieldArgs, queryFields := filter.queryFields()
	args = append(args, fieldArgs...)

	if queryFields != "" {
		query += "AND " + queryFields
	}

	if filterBots && filter.minIsBot > 0 {
		args = append(args, filter.minIsBot)
		query += " AND is_bot < ? "
	}

	return args, query
}

func (filter *Filter) appendQuery(queryFields *[]string, args *[]any, field, value string) {
	if value != "" {
		comparator := "%s = ? "
		not := strings.HasPrefix(value, "!")

		if field == "event_meta_keys" {
			if not {
				value = value[1:]
				comparator = "!has(%s, ?) "
			} else {
				comparator = "has(%s, ?) "
			}
		} else if not {
			value = value[1:]
			comparator = "%s != ? "
		}

		*args = append(*args, filter.nullValue(value))
		*queryFields = append(*queryFields, fmt.Sprintf(comparator, field))
	}
}

func (filter *Filter) appendQueryUInt16(queryFields *[]string, args *[]any, field, value string) {
	if value != "" {
		comparator := "%s = ? "

		if strings.HasPrefix(value, "!") {
			value = value[1:]
			comparator = "%s != ? "
		}

		var valueInt uint16

		if strings.ToLower(value) != "null" {
			i, err := strconv.ParseUint(value, 10, 16)

			if err == nil {
				valueInt = uint16(i)
			}
		}

		*args = append(*args, valueInt)
		*queryFields = append(*queryFields, fmt.Sprintf(comparator, field))
	}
}

func (filter *Filter) appendQueryMeta(queryFields *[]string, args *[]any, kv map[string]string) {
	for k, v := range kv {
		comparator := "event_meta_values[indexOf(event_meta_keys, '%s')] = ? "

		if strings.HasPrefix(v, "!") {
			v = v[1:]
			comparator = "event_meta_values[indexOf(event_meta_keys, '%s')] != ? "
		}

		*args = append(*args, filter.nullValue(v))
		*queryFields = append(*queryFields, fmt.Sprintf(comparator, k))
	}
}

func (filter *Filter) nullValue(value string) string {
	if strings.ToLower(value) == "null" {
		return ""
	}

	return value
}

func (filter *Filter) toDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func (filter *Filter) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
