package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v5"
	"github.com/pirsch-analytics/pirsch/v5/db"
	"github.com/pirsch-analytics/pirsch/v5/model"
	"sort"
	"strings"
)

// Pages aggregates statistics regarding pages.
type Pages struct {
	analyzer *Analyzer
	store    db.Store
}

// ByPath returns the visitor count, session count, bounce rate, views, and average time on page grouped by path and (optional) page title.
func (pages *Pages) ByPath(filter *Filter) ([]model.PageStats, error) {
	return pages.byPath(filter, false)
}

// ByEventPath returns the visitor count, session count, bounce rate, views, and average time on page grouped by event path and (optional) title.
func (pages *Pages) ByEventPath(filter *Filter) ([]model.PageStats, error) {
	return pages.byPath(filter, true)
}

func (pages *Pages) byPath(filter *Filter, eventPath bool) ([]model.PageStats, error) {
	filter = pages.analyzer.getFilter(filter)
	pathField := FieldPath

	if eventPath {
		pathField = FieldEventPath
	}

	fields := []Field{
		pathField,
		FieldVisitors,
		FieldSessions,
		FieldRelativeVisitors,
		FieldViews,
		FieldRelativeViews,
		FieldBounces,
		FieldBounceRate,
	}
	groupBy := []Field{
		pathField,
	}
	orderBy := []Field{
		FieldVisitors,
		pathField,
	}

	if filter.IncludeTitle {
		if eventPath {
			fields = append(fields, FieldEventTitle)
			groupBy = append(groupBy, FieldEventTitle)
			orderBy = append(orderBy, FieldEventTitle)
		} else {
			fields = append(fields, FieldTitle)
			groupBy = append(groupBy, FieldTitle)
			orderBy = append(orderBy, FieldTitle)
		}
	}

	q, args := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := pages.store.SelectPageStats(filter.IncludeTitle, false, q, args...)

	if err != nil {
		return nil, err
	}

	if filter.IncludeTimeOnPage {
		pathList := getPathList(stats)
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
	var sortVisitors pirsch.Direction

	if len(filter.Sort) > 0 && filter.Sort[0].Field == FieldVisitors {
		sortVisitors = filter.Sort[0].Direction
		filter.Sort = filter.Sort[:0]
	}

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

	q, args := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := pages.store.SelectEntryStats(filter.IncludeTitle, q, args...)

	if err != nil {
		return nil, err
	}

	pathList := getPathList(stats)
	totalSessions, err := pages.totalSessions(filter)

	if err != nil {
		return nil, err
	}

	totalSessionsFloat64 := float64(totalSessions)
	total, err := pages.totalVisitorsSessions(filter, pathList)

	if err != nil {
		return nil, err
	}

	for i := range stats {
		for j := range total {
			if stats[i].Path == total[j].Path {
				stats[i].Visitors = total[j].Visitors
				stats[i].Sessions = total[j].Sessions
				stats[i].EntryRate = float64(stats[i].Entries) / totalSessionsFloat64
				break
			}
		}
	}

	if filter.IncludeTimeOnPage {
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

	if sortVisitors != "" {
		if sortVisitors == pirsch.DirectionASC {
			sort.Slice(stats, func(i, j int) bool {
				return stats[i].Visitors < stats[j].Visitors
			})
		} else {
			sort.Slice(stats, func(i, j int) bool {
				return stats[i].Visitors > stats[j].Visitors
			})
		}
	}

	return stats, nil
}

// Exit returns the visitor count and time on page grouped by path and (optional) page title for the last page visited.
func (pages *Pages) Exit(filter *Filter) ([]model.ExitStats, error) {
	filter = pages.analyzer.getFilter(filter)
	var sortVisitors pirsch.Direction

	if len(filter.Sort) > 0 && filter.Sort[0].Field == FieldVisitors {
		sortVisitors = filter.Sort[0].Direction
		filter.Sort = filter.Sort[:0]
	}

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

	q, args := filter.buildQuery(fields, groupBy, orderBy)
	stats, err := pages.store.SelectExitStats(filter.IncludeTitle, q, args...)

	if err != nil {
		return nil, err
	}

	pathList := getPathList(stats)
	totalSessions, err := pages.totalSessions(filter)

	if err != nil {
		return nil, err
	}

	totalSessionsFloat64 := float64(totalSessions)
	total, err := pages.totalVisitorsSessions(filter, pathList)

	if err != nil {
		return nil, err
	}

	for i := range stats {
		for j := range total {
			if stats[i].Path == total[j].Path {
				stats[i].Visitors = total[j].Visitors
				stats[i].Sessions = total[j].Sessions
				stats[i].ExitRate = float64(stats[i].Exits) / totalSessionsFloat64
				break
			}
		}
	}

	if sortVisitors != "" {
		if sortVisitors == pirsch.DirectionASC {
			sort.Slice(stats, func(i, j int) bool {
				return stats[i].Visitors < stats[j].Visitors
			})
		} else {
			sort.Slice(stats, func(i, j int) bool {
				return stats[i].Visitors > stats[j].Visitors
			})
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

	q, args := filter.buildQuery([]Field{
		FieldVisitors,
		FieldViews,
		FieldCR,
	}, nil, []Field{
		FieldVisitors,
	})
	stats, err := pages.store.GetPageConversionsStats(q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (pages *Pages) totalSessions(filter *Filter) (int, error) {
	filter = pages.analyzer.getFilter(filter)
	filterQuery, filterArgs := filter.buildTimeQuery()
	query := fmt.Sprintf("SELECT uniq(visitor_id, session_id) FROM session %s HAVING sum(sign) > 0", filterQuery)
	stats, err := pages.store.SelectTotalSessions(query, filterArgs...)

	if err != nil {
		return 0, err
	}

	return stats, nil
}

func (pages *Pages) totalVisitorsSessions(filter *Filter, paths []string) ([]model.TotalVisitorSessionStats, error) {
	if len(paths) == 0 {
		return []model.TotalVisitorSessionStats{}, nil
	}

	filter = pages.analyzer.getFilter(filter)
	filter.Path = nil
	filter.EntryPath = nil
	filter.ExitPath = nil
	filter.AnyPath = paths
	filter.IncludeTitle = false
	filter.Sort = nil
	q, args := filter.buildQuery([]Field{
		FieldPath,
		FieldVisitors,
		FieldSessions,
		FieldViews,
	}, []Field{
		FieldPath,
	}, []Field{
		FieldVisitors,
		FieldSessions,
	})
	stats, err := pages.store.SelectTotalVisitorSessionStats(q, args...)

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
	filter.Sort = nil
	q := queryBuilder{
		filter: filter,
		from:   pageViews,
		search: filter.Search,
	}
	fields := q.getFields()
	fields = append(fields, "duration_seconds")
	hasPath := false

	for _, field := range fields {
		if field == FieldPath.Name {
			hasPath = true
			break
		}
	}

	if !hasPath {
		fields = append(fields, FieldPath.Name)
	}

	var query strings.Builder
	query.WriteString(fmt.Sprintf(`SELECT path,
		ifNull(toUInt64(avg(nullIf(time_on_page, 0))), 0) average_time_spent_seconds
		FROM (
			SELECT path,
			%s time_on_page
			FROM (
				SELECT v.session_id sid,
				%s
				FROM page_view v `, pages.analyzer.timeOnPageQuery(filter), strings.Join(fields, ",")))

	if pages.analyzer.minIsBot > 0 || len(filter.EntryPath) != 0 || len(filter.ExitPath) != 0 {
		sessionsQuery := queryBuilder{
			filter: filter,
			from:   sessions,
			search: filter.Search,
			fields: []Field{
				FieldVisitorID,
				FieldSessionID,
			},
			groupBy: []Field{
				FieldVisitorID,
				FieldSessionID,
			},
		}
		str, args := sessionsQuery.query()
		q.args = append(q.args, args...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (%s) s ON v.visitor_id = s.visitor_id AND v.session_id = s.session_id `, str))
	}

	if len(filter.EventName) > 0 {
		eventsQuery := queryBuilder{
			filter: filter,
			from:   events,
			search: filter.Search,
			fields: []Field{
				FieldVisitorID,
				FieldSessionID,
			},
			groupBy: []Field{
				FieldVisitorID,
				FieldSessionID,
			},
		}
		str, args := eventsQuery.query()
		q.args = append(q.args, args...)
		query.WriteString(fmt.Sprintf(`INNER JOIN (%s) ev ON v.visitor_id = ev.visitor_id AND v.session_id = ev.session_id `, str))
	}

	whereTime := q.whereTime()[len("WHERE "):]
	q.whereFields()
	query.WriteString(fmt.Sprintf(`WHERE %s ORDER BY v.visitor_id, v.session_id, time)
		WHERE time_on_page > 0 AND sid = neighbor(sid, 1, null) %s) GROUP BY path`, whereTime, q.q.String()))
	stats, err := pages.store.SelectAvgTimeSpentStats(query.String(), q.args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func getPathList[T interface{ GetPath() string }](stats []T) []string {
	paths := make(map[string]struct{})

	for i := range stats {
		paths[stats[i].GetPath()] = struct{}{}
	}

	pathList := make([]string, 0, len(paths))

	for path := range paths {
		pathList = append(pathList, path)
	}

	return pathList
}
