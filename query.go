package pirsch

import (
	"fmt"
	"strings"
)

var (
	fieldPath = field{
		querySessions:  "path",
		queryPageViews: "path",
	}
	fieldUniqVisitors = field{
		querySessions:  "uniq(visitor_id)",
		queryPageViews: "uniq(visitor_id)",
		name:           "visitors",
	}
	fieldUniqSessions = field{
		querySessions:  "uniq(visitor_id, session_id)",
		queryPageViews: "uniq(visitor_id, session_id)",
		name:           "sessions",
	}
	fieldViews = field{
		querySessions:  "sum(page_views*sign)",
		queryPageViews: "count(1)",
		name:           "views",
	}
	fieldBounces = field{
		querySessions:  "sum(is_bounce*sign)",
		queryPageViews: "sum(is_bounce)",
		name:           "bounces",
	}
	fieldBounceRate = field{
		querySessions:  "bounces / IF(sessions = 0, 1, sessions)",
		queryPageViews: "bounces / IF(sessions = 0, 1, sessions)",
		name:           "bounce_rate",
	}
)

type field struct {
	querySessions  string
	queryPageViews string
	name           string
}

func baseQuery(filter *Filter, fields []field, groupBy []string) ([]interface{}, string) {
	table := filter.table()
	args := make([]interface{}, 0)
	var query strings.Builder

	if table == "event" || filter.Path != "" || filter.PathPattern != "" || fieldsContain(fields, "path") {
		if table == "session" {
			table = "page_view"
		}

		query.WriteString(fmt.Sprintf(`SELECT %s FROM %s v `, joinPageViewFields(fields), table))

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
			query.WriteString(fmt.Sprintf(`GROUP BY %s`, strings.Join(groupBy, ",")))
		}
	} else {
		filterArgs, filterQuery := filter.query()
		args = append(args, filterArgs...)
		query.WriteString(fmt.Sprintf(`SELECT %s
			FROM session
			WHERE %s `, joinSessionFields(fields), filterQuery))

		if len(groupBy) > 0 {
			query.WriteString(fmt.Sprintf(`GROUP BY %s `, strings.Join(groupBy, ",")))
		}

		query.WriteString(`HAVING sum(sign) > 0`)
	}

	return args, query.String()
}

func fieldsContain(haystack []field, needle string) bool {
	for i := range haystack {
		if haystack[i].querySessions == needle {
			return true
		}
	}

	return false
}

func joinPageViewFields(fields []field) string {
	var out strings.Builder

	for i := range fields {
		out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].queryPageViews, fields[i].name))
	}

	str := out.String()

	if len(str) == 0 {
		return ""
	}

	return str[:len(str)-1]
}

func joinSessionFields(fields []field) string {
	var out strings.Builder

	for i := range fields {
		out.WriteString(fmt.Sprintf(`%s %s,`, fields[i].querySessions, fields[i].name))
	}

	str := out.String()

	if len(str) == 0 {
		return ""
	}

	return str[:len(str)-1]
}
