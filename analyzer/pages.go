package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"strings"
)

// Pages aggregates statistics regarding pages.
type Pages struct {
	analyzer *Analyzer
	store    db.Store
}

// ByPath returns the visitor count, session count, bounce rate, views, and average time on page grouped by path and (optional) page title.
func (pages *Pages) ByPath(filter *Filter) ([]model.PageStats, error) {
	filter = pages.analyzer.getFilter(filter)
	fields := []Field{
		FieldPath,
		FieldVisitors,
		FieldSessions,
		FieldRelativeVisitors,
		FieldViews,
		FieldRelativeViews,
		FieldBounces,
		FieldBounceRate,
	}
	groupBy := []Field{
		FieldPath,
	}
	orderBy := []Field{
		FieldVisitors,
		FieldPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldTitle)
		groupBy = append(groupBy, FieldTitle)
		orderBy = append(orderBy, FieldTitle)
	}

	if filter.table() == "event" {
		fields = append(fields, FieldEventTimeSpent)
	}

	args, query := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := pages.store.SelectPageStats(filter.IncludeTitle, filter.table() == "event", query, args...)

	if err != nil {
		return nil, err
	}

	if filter.IncludeTimeOnPage && filter.table() == "session" {
		paths := make(map[string]struct{})

		for i := range stats {
			paths[stats[i].Path] = struct{}{}
		}

		pathList := make([]string, 0, len(paths))

		for path := range paths {
			pathList = append(pathList, path)
		}

		top, err := pages.avgTimeOnPage(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range top {
				if stats[i].Path == top[j].Path {
					stats[i].AverageTimeSpentSeconds = top[j].AverageTimeSpentSeconds
					break
				}
			}
		}
	}

	return stats, nil
}

// Entry returns the visitor count and time on page grouped by path and (optional) page title for the first page visited.
func (pages *Pages) Entry(filter *Filter) ([]model.EntryStats, error) {
	filter = pages.analyzer.getFilter(filter)

	fields := []Field{
		FieldEntryPath,
		FieldEntries,
	}
	groupBy := []Field{
		FieldEntryPath,
	}
	orderBy := []Field{
		FieldEntries,
		FieldEntryPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldEntryTitle)
		groupBy = append(groupBy, FieldEntryTitle)
		orderBy = append(orderBy, FieldEntryTitle)
	}

	args, query := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := pages.store.SelectEntryStats(filter.IncludeTitle, query, args...)

	if err != nil {
		return nil, err
	}

	paths := make(map[string]struct{})

	for i := range stats {
		paths[stats[i].Path] = struct{}{}
	}

	pathList := make([]string, 0, len(paths))

	for path := range paths {
		pathList = append(pathList, path)
	}

	if filter.table() != "event" {
		total, err := pages.totalVisitorsSessions(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range total {
				if stats[i].Path == total[j].Path {
					stats[i].Visitors = total[j].Visitors
					stats[i].Sessions = total[j].Sessions
					stats[i].EntryRate = float64(stats[i].Entries) / float64(total[j].Sessions)
					break
				}
			}
		}
	}

	if filter.IncludeTimeOnPage && filter.table() != "event" {
		top, err := pages.avgTimeOnPage(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range top {
				if stats[i].Path == top[j].Path {
					stats[i].AverageTimeSpentSeconds = top[j].AverageTimeSpentSeconds
					break
				}
			}
		}
	}

	return stats, nil
}

// Exit returns the visitor count and time on page grouped by path and (optional) page title for the last page visited.
func (pages *Pages) Exit(filter *Filter) ([]model.ExitStats, error) {
	filter = pages.analyzer.getFilter(filter)

	fields := []Field{
		FieldExitPath,
		FieldExits,
	}
	groupBy := []Field{
		FieldExitPath,
	}
	orderBy := []Field{
		FieldExits,
		FieldExitPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldExitTitle)
		groupBy = append(groupBy, FieldExitTitle)
		orderBy = append(orderBy, FieldExitTitle)
	}

	args, query := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := pages.store.SelectExitStats(filter.IncludeTitle, query, args...)

	if err != nil {
		return nil, err
	}

	if filter.table() != "event" {
		paths := make(map[string]struct{})

		for i := range stats {
			paths[stats[i].Path] = struct{}{}
		}

		pathList := make([]string, 0, len(paths))

		for path := range paths {
			pathList = append(pathList, path)
		}

		total, err := pages.totalVisitorsSessions(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range total {
				if stats[i].Path == total[j].Path {
					stats[i].Visitors = total[j].Visitors
					stats[i].Sessions = total[j].Sessions
					stats[i].ExitRate = float64(stats[i].Exits) / float64(total[j].Sessions)
					break
				}
			}
		}
	}

	return stats, nil
}

// Conversions returns the visitor count, views, and conversion rate for conversion goals.
// This function is supposed to be used with the Filter.PathPattern, to list page conversions.
func (pages *Pages) Conversions(filter *Filter) (*model.PageConversionsStats, error) {
	filter = pages.analyzer.getFilter(filter)

	if len(filter.PathPattern) == 0 {
		return nil, nil
	}

	args, query := filter.buildQuery([]Field{
		FieldVisitors,
		FieldViews,
		FieldCR,
	}, nil, []Field{
		FieldVisitors,
	})
	stats, err := pages.store.GetPageConversionsStats(query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (pages *Pages) totalVisitorsSessions(filter *Filter, paths []string) ([]model.TotalVisitorSessionStats, error) {
	if len(paths) == 0 {
		return []model.TotalVisitorSessionStats{}, nil
	}

	filter = pages.analyzer.getFilter(filter)
	filter.Path, filter.PathPattern, filter.EntryPath, filter.ExitPath = []string{}, []string{}, []string{}, []string{}
	filterArgs, filterQuery := filter.query(pages.analyzer.minIsBot > 0)
	pathQuery := strings.Repeat("?,", len(paths))

	for _, path := range paths {
		filterArgs = append(filterArgs, path)
	}

	var query string

	if pages.analyzer.minIsBot > 0 {
		query = fmt.Sprintf(`SELECT path,
			uniq(visitor_id) visitors,
			uniq(visitor_id, session_id) sessions,
			count(1) views
			FROM page_view v
			INNER JOIN (
				SELECT visitor_id,
				session_id
				FROM session
				WHERE %s
				GROUP BY visitor_id, session_id
				HAVING sum(sign) > 0
			) s
			ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id
			WHERE path IN (%s)
			GROUP BY path
			ORDER BY visitors DESC, sessions DESC
			%s`, filterQuery, pathQuery[:len(pathQuery)-1], filter.withLimit())
	} else {
		query = fmt.Sprintf(`SELECT path,
			uniq(visitor_id) visitors,
			uniq(visitor_id, session_id) sessions,
			count(1) views
			FROM page_view
			WHERE %s
			AND path IN (%s)
			GROUP BY path
			ORDER BY visitors DESC, sessions DESC
			%s`, filterQuery, pathQuery[:len(pathQuery)-1], filter.withLimit())
	}

	stats, err := pages.store.SelectTotalVisitorSessionStats(query, filterArgs...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (pages *Pages) avgTimeOnPage(filter *Filter, paths []string) ([]model.AvgTimeSpentStats, error) {
	if len(paths) == 0 {
		return []model.AvgTimeSpentStats{}, nil
	}

	filter = pages.analyzer.getFilter(filter)

	if filter.table() == "event" {
		return []model.AvgTimeSpentStats{}, nil
	}

	filter.Search, filter.Sort, filter.Offset, filter.Limit = nil, nil, 0, 0
	timeArgs, timeQuery := filter.queryTime(false)
	fieldArgs, fieldQuery := filter.queryFields()

	if len(fieldArgs) > 0 {
		fieldQuery = "AND " + fieldQuery
	}

	fieldsQuery := filter.fields()

	if fieldsQuery != "" {
		fieldsQuery = "," + fieldsQuery
	}

	args := make([]any, 0, len(timeArgs)*2+len(fieldArgs))
	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT path,
		ifNull(toUInt64(avg(nullIf(time_on_page, 0))), 0) average_time_spent_seconds
		FROM (
			SELECT path,
			%s time_on_page
			FROM (
				SELECT session_id,
				path,
				duration_seconds
				%s
				FROM page_view v `, pages.analyzer.timeOnPageQuery(filter), fieldsQuery))

	if pages.analyzer.minIsBot > 0 || len(filter.EntryPath) != 0 || len(filter.ExitPath) != 0 {
		innerTimeArgs, innerTimeQuery := filter.queryTime(false)
		args = append(args, innerTimeArgs...)
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
		ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, innerTimeQuery))
	}

	args = append(args, timeArgs...)
	pathQuery := strings.Repeat("?,", len(paths))

	for _, path := range paths {
		args = append(args, path)
	}

	args = append(args, fieldArgs...)
	query.WriteString(fmt.Sprintf(`WHERE %s
				ORDER BY visitor_id, session_id, time
			)
			WHERE time_on_page > 0
			AND session_id = neighbor(session_id, 1, null)
			AND path IN (%s)
			%s
		)
		GROUP BY path`, timeQuery, pathQuery[:len(pathQuery)-1], fieldQuery))
	stats, err := pages.store.SelectAvgTimeSpentStats(query.String(), args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
