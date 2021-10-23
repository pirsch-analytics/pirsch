package pirsch

import (
	"fmt"
	"strings"
)

var (
	fieldPath = field{
		querySessions:  "path",
		queryPageViews: "path",
		queryDirection: "ASC",
		name:           "path",
	}
	fieldVisitors = field{
		querySessions:  "uniq(visitor_id)",
		queryPageViews: "uniq(visitor_id)",
		queryDirection: "DESC",
		name:           "visitors",
	}
	fieldRelativeVisitors = field{
		querySessions:  "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryPageViews: "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "relative_visitors",
	}
	fieldSessions = field{
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
	fieldRelativeViews = field{
		querySessions:  "views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1)",
		queryPageViews: "views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "relative_views",
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
	fieldHour = field{
		querySessions:  "toHour(time, '%s')",
		queryPageViews: "toHour(time, '%s')",
		queryDirection: "ASC",
		queryWithFill:  "WITH FILL FROM 0 TO 24",
		timezone:       true,
		name:           "hour",
	}
)

type field struct {
	querySessions  string
	queryPageViews string
	queryDirection string
	queryWithFill  string
	withFill       bool
	timezone       bool
	filterTime     bool
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

		query.WriteString(fmt.Sprintf(`SELECT %s FROM %s v `, joinPageViewFields(&args, filter, fields), table))

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
			WHERE %s `, joinSessionFields(&args, filter, fields), filterQuery))

		if len(groupBy) > 0 {
			query.WriteString(fmt.Sprintf(`GROUP BY %s `, joinGroupBy(groupBy)))
		}

		query.WriteString(`HAVING sum(sign) > 0 `)

		if len(orderBy) > 0 {
			query.WriteString(fmt.Sprintf(`ORDER BY %s `, joinOrderBy(&args, filter, orderBy)))
		}
	}

	query.WriteString(filter.withLimit())
	return args, query.String()
}

func joinPageViewFields(args *[]interface{}, filter *Filter, fields []field) string {
	if len(fields) == 0 {
		return ""
	}

	var out strings.Builder

	for i := range fields {
		if fields[i].filterTime {
			timeArgs, timeQuery := filter.queryTime()
			*args = append(*args, timeArgs...)
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].queryPageViews, timeQuery), fields[i].name))
		} else if fields[i].timezone {
			out.WriteString(fmt.Sprintf(`%s %s,`, fmt.Sprintf(fields[i].queryPageViews, filter.Timezone.String()), fields[i].name))
		} else {
			out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].queryPageViews, fields[i].name))
		}
	}

	str := out.String()
	return str[:len(str)-1]
}

func joinSessionFields(args *[]interface{}, filter *Filter, fields []field) string {
	if len(fields) == 0 {
		return ""
	}

	var out strings.Builder

	for i := range fields {
		if fields[i].filterTime {
			timeArgs, timeQuery := filter.queryTime()
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

func fieldsContain(haystack []field, needle string) bool {
	for i := range haystack {
		if haystack[i].querySessions == needle {
			return true
		}
	}

	return false
}
