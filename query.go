package pirsch

import (
	"fmt"
	"strings"
	"time"
)

var (
	fieldPath = field{
		querySessions:  "path",
		queryPageViews: "path",
		queryDirection: "ASC",
		name:           "path",
	}
	fieldUniqVisitors = field{
		querySessions:  "uniq(visitor_id)",
		queryPageViews: "uniq(visitor_id)",
		queryDirection: "DESC",
		name:           "visitors",
	}
	fieldUniqSessions = field{
		querySessions:  "uniq(visitor_id, session_id)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		name:           "sessions",
	}
	fieldViews = field{
		querySessions:  "sum(page_views*sign)",
		queryPageViews: "count(1)",
		queryDirection: "DESC",
		name:           "views",
	}
	fieldBounces = field{
		querySessions:  "sum(is_bounce*sign)",
		queryPageViews: "sum(is_bounce)",
		queryDirection: "DESC",
		name:           "bounces",
	}
	fieldBounceRate = field{
		querySessions:  "bounces / IF(sessions = 0, 1, sessions)",
		queryPageViews: "bounces / IF(sessions = 0, 1, sessions)",
		queryDirection: "DESC",
		name:           "bounce_rate",
	}
	fieldDay = field{
		querySessions:  "toDate(time, '%s')",
		queryPageViews: "toDate(time, '%s')",
		queryDirection: "ASC",
		withFill:       true,
		timezone:       true,
		name:           "day",
	}
)

type field struct {
	querySessions  string
	queryPageViews string
	queryDirection string
	withFill       bool
	timezone       bool
	name           string
}

func buildQuery(filter *Filter, fields, groupBy, orderBy []field) ([]interface{}, string) {
	table := filter.table()
	args := make([]interface{}, 0)
	var query strings.Builder

	if table == "event" || filter.Path != "" || filter.PathPattern != "" || fieldsContain(fields, "path") {
		if table == "session" {
			table = "page_view"
		}

		query.WriteString(fmt.Sprintf(`SELECT %s FROM %s v `, joinPageViewFields(fields, filter.Timezone), table))

		if filter.EntryPath != "" ||
			filter.ExitPath != "" ||
			fieldsContain(fields, fieldBounces.querySessions) ||
			fieldsContain(fields, fieldViews.querySessions) {
			path, pathPattern, eventName := filter.Path, filter.PathPattern, filter.EventName
			filter.Path, filter.PathPattern, filter.EventName = "", "", ""
			filterArgs, filterQuery := filter.query()
			filter.Path, filter.PathPattern, filter.EventName = path, pathPattern, eventName
			args = append(args, filterArgs...)
			query.WriteString(fmt.Sprintf(`INNER JOIN (
				SELECT visitor_id,
				session_id,
				sum(is_bounce*sign) is_bounce,
				sum(page_views*sign) page_views
				FROM session
				WHERE %s
				GROUP BY visitor_id, session_id
				HAVING sum(sign) > 0
			) s
			ON s.visitor_id = v.visitor_id AND s.session_id = v.session_id `, filterQuery))

			if filter.Path != "" || filter.PathPattern != "" || filter.EventName != "" {
				filterArgs, filterQuery = filter.queryPageOrEvent()
				args = append(args, filterArgs...)
				query.WriteString(fmt.Sprintf(`WHERE %s `, filterQuery))
			}
		} else {
			filterArgs, filterQuery := filter.query()
			args = append(args, filterArgs...)
			query.WriteString(fmt.Sprintf(`WHERE %s `, filterQuery))
		}

		if len(groupBy) > 0 {
			query.WriteString(fmt.Sprintf(`GROUP BY %s `, joinGroupBy(groupBy)))
		}

		if len(orderBy) > 0 {
			query.WriteString(fmt.Sprintf(`ORDER BY %s `, joinOrderBy(&args, filter, orderBy)))
		}
	} else {
		filterArgs, filterQuery := filter.query()
		args = append(args, filterArgs...)
		query.WriteString(fmt.Sprintf(`SELECT %s
			FROM session
			WHERE %s `, joinSessionFields(fields, filter.Timezone), filterQuery))

		if len(groupBy) > 0 {
			query.WriteString(fmt.Sprintf(`GROUP BY %s `, joinGroupBy(groupBy)))
		}

		query.WriteString(`HAVING sum(sign) > 0 `)

		if len(orderBy) > 0 {
			query.WriteString(fmt.Sprintf(`ORDER BY %s `, joinOrderBy(&args, filter, orderBy)))
		}
	}

	return args, query.String()
}

func joinPageViewFields(fields []field, tz *time.Location) string {
	if len(fields) == 0 {
		return ""
	}

	var out strings.Builder

	for i := range fields {
		if fields[i].timezone {
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].queryPageViews, tz.String()), fields[i].name))
		} else {
			out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].queryPageViews, fields[i].name))
		}
	}

	str := out.String()
	return str[:len(str)-1]
}

func joinSessionFields(fields []field, tz *time.Location) string {
	if len(fields) == 0 {
		return ""
	}

	var out strings.Builder

	for i := range fields {
		if fields[i].timezone {
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].querySessions, tz.String()), fields[i].name))
		} else {
			out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].querySessions, fields[i].name))
		}
	}

	str := out.String()
	return str[:len(str)-1]
}

func joinGroupBy(fields []field) string {
	if len(fields) == 0 {
		return ""
	}

	var out strings.Builder

	for i := range fields {
		out.WriteString(fields[i].name + ",")
	}

	str := out.String()
	return str[:len(str)-1]
}

func joinOrderBy(args *[]interface{}, filter *Filter, fields []field) string {
	if len(fields) == 0 {
		return ""
	}

	var out strings.Builder

	for i := range fields {
		if fields[i].withFill {
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

func fieldsContain(haystack []field, needle string) bool {
	for i := range haystack {
		if haystack[i].querySessions == needle {
			return true
		}
	}

	return false
}
