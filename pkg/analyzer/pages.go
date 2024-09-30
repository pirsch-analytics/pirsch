package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"sort"
	"strings"
)

// Pages aggregates statistics regarding pages.
type Pages struct {
	analyzer *Analyzer
	store    db.Store
}

// Hostname returns the visitor count, session count, bounce rate, and views grouped by hostname.
func (pages *Pages) Hostname(filter *Filter) ([]model.HostnameStats, error) {
	filter = pages.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldHostname,
		FieldVisitors,
		FieldViews,
		FieldSessions,
		FieldBounces,
		FieldRelativeVisitors,
		FieldRelativeViews,
		FieldBounceRate,
	}, []Field{
		FieldHostname,
	}, []Field{
		FieldVisitors,
		FieldHostname,
	}, nil, "")
	stats, err := pages.store.SelectHostnameStats(filter.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// ByPath returns the visitor count, session count, bounce rate, views, and average time on page grouped by hostname, path, and (optional) page title.
func (pages *Pages) ByPath(filter *Filter) ([]model.PageStats, error) {
	return pages.byPath(filter, false)
}

// ByEventPath returns the visitor count, session count, bounce rate, views, and average time on page grouped by hostname, event path, and (optional) title.
func (pages *Pages) ByEventPath(filter *Filter) ([]model.PageStats, error) {
	if len(filter.EventName) == 0 {
		return []model.PageStats{}, nil
	}

	return pages.byPath(filter, true)
}

func (pages *Pages) byPath(filter *Filter, eventPath bool) ([]model.PageStats, error) {
	filter = pages.analyzer.getFilter(filter)
	pathField := FieldPath

	if eventPath {
		pathField = FieldEventPath
	}

	fields := []Field{
		FieldHostname,
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
		FieldHostname,
		pathField,
	}
	orderBy := []Field{
		FieldVisitors,
		FieldHostname,
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

	q, args := filter.buildQuery(fields, groupBy, orderBy, []Field{
		FieldPath,
		FieldVisitors,
		FieldViews,
		FieldSessions,
		FieldBounces,
	}, "imported_page")
	stats, err := pages.store.SelectPageStats(filter.Ctx, filter.IncludeTitle, false, q, args...)

	if err != nil {
		return nil, err
	}

	if filter.IncludeTimeOnPage {
		n := len(stats)

		for start := 0; start < n; start += 1000 {
			end := start + 1000

			if end > n {
				end = n
			}

			pathList := getPathList(stats[start:end])
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
	}

	return stats, nil
}

// Entry returns the visitor count and time on page grouped by hostname, path, and (optional) page title for the first page visited.
func (pages *Pages) Entry(filter *Filter) ([]model.EntryStats, error) {
	filter = pages.analyzer.getFilter(filter)
	var sortVisitors pkg.Direction

	if len(filter.Sort) > 0 && filter.Sort[0].Field == FieldVisitors {
		sortVisitors = filter.Sort[0].Direction
		filter.Sort = filter.Sort[:0]
	}

	fields := []Field{
		FieldHostname,
		FieldEntryPath,
		FieldEntries,
		FieldEntryRate,
	}
	groupBy := []Field{
		FieldHostname,
		FieldEntryPath,
	}
	orderBy := []Field{
		FieldEntries,
		FieldHostname,
		FieldEntryPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldEntryTitle)
		groupBy = append(groupBy, FieldEntryTitle)
		orderBy = append(orderBy, FieldEntryTitle)
	}

	q, args := filter.buildQuery(fields, groupBy, orderBy, []Field{
		FieldEntryPath,
		FieldVisitors,
	}, "imported_entry_page")
	stats, err := pages.store.SelectEntryStats(filter.Ctx, filter.IncludeTitle, q, args...)

	if err != nil {
		return nil, err
	}

	n := len(stats)

	for start := 0; start < n; start += 1000 {
		end := start + 1000

		if end > n {
			end = n
		}

		pathList := getPathList(stats[start:end])
		total, err := pages.totalVisitorsSessions(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range total {
				if stats[i].Path == total[j].Path {
					stats[i].Visitors = total[j].Visitors
					stats[i].Sessions = total[j].Sessions
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
	}

	if sortVisitors != "" {
		if sortVisitors == pkg.DirectionASC {
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

// Exit returns the visitor count and time on page grouped by hostname, path, and (optional) page title for the last page visited.
func (pages *Pages) Exit(filter *Filter) ([]model.ExitStats, error) {
	filter = pages.analyzer.getFilter(filter)
	var sortVisitors pkg.Direction

	if len(filter.Sort) > 0 && filter.Sort[0].Field == FieldVisitors {
		sortVisitors = filter.Sort[0].Direction
		filter.Sort = filter.Sort[:0]
	}

	fields := []Field{
		FieldHostname,
		FieldExitPath,
		FieldExits,
		FieldExitRate,
	}
	groupBy := []Field{
		FieldHostname,
		FieldExitPath,
	}
	orderBy := []Field{
		FieldExits,
		FieldHostname,
		FieldExitPath,
	}

	if filter.IncludeTitle {
		fields = append(fields, FieldExitTitle)
		groupBy = append(groupBy, FieldExitTitle)
		orderBy = append(orderBy, FieldExitTitle)
	}

	q, args := filter.buildQuery(fields, groupBy, orderBy, []Field{
		FieldExitPath,
		FieldVisitors,
	}, "imported_exit_page")
	stats, err := pages.store.SelectExitStats(filter.Ctx, filter.IncludeTitle, q, args...)

	if err != nil {
		return nil, err
	}

	n := len(stats)

	for start := 0; start < n; start += 1000 {
		end := start + 1000

		if end > n {
			end = n
		}

		pathList := getPathList(stats[start:end])
		total, err := pages.totalVisitorsSessions(filter, pathList)

		if err != nil {
			return nil, err
		}

		for i := range stats {
			for j := range total {
				if stats[i].Path == total[j].Path {
					stats[i].Visitors = total[j].Visitors
					stats[i].Sessions = total[j].Sessions
					break
				}
			}
		}
	}

	if sortVisitors != "" {
		if sortVisitors == pkg.DirectionASC {
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

// Conversions returns the visitor count, views, conversion rate, and custom metric for conversion goals.
func (pages *Pages) Conversions(filter *Filter) (*model.ConversionsStats, error) {
	filter = pages.analyzer.getFilter(filter)
	fields := []Field{
		FieldVisitors,
		FieldViews,
		FieldCR,
	}
	includeCustomMetric := false

	if len(filter.EventName) > 0 && filter.CustomMetricType != "" && filter.CustomMetricKey != "" {
		fields = append(fields, FieldEventMetaCustomMetricAvg, FieldEventMetaCustomMetricTotal)
		includeCustomMetric = true
	}

	q, args := filter.buildQuery(fields, nil, []Field{FieldVisitors}, nil, "")
	stats, err := pages.store.GetConversionsStats(filter.Ctx, q, includeCustomMetric, args...)

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
	filter.Path = nil
	filter.EntryPath = nil
	filter.ExitPath = nil
	filter.AnyPath = paths
	filter.PathPattern = nil
	filter.Tags = nil
	filter.Search = nil
	filter.IncludeTitle = false
	filter.Sort = nil
	filter.Offset = 0
	filter.Limit = 0
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
	}, nil, "")
	stats, err := pages.store.SelectTotalVisitorSessionStats(filter.Ctx, q, args...)

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
	filter.Search = nil
	q := queryBuilder{
		filter: filter,
		from:   pageViews,
		search: filter.Search,
	}
	fields := q.getFields()
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
	query.WriteString(fmt.Sprintf(`SELECT path, round(avg(time_on_page)) average_time_spent_seconds
		FROM (
			SELECT nth_value(%s, 2) OVER (PARTITION BY v.visitor_id, v.session_id ORDER BY v."time" ASC Rows BETWEEN CURRENT ROW AND 1 FOLLOWING) AS time_on_page,
				%s
			FROM page_view v `, pages.analyzer.timeOnPageQuery(filter), strings.Join(fields, ",")))

	if len(filter.EntryPath) > 0 || len(filter.ExitPath) > 0 {
		sessionsQuery := queryBuilder{
			filter: filter,
			from:   sessions,
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

	whereTime := q.whereTime()
	q.whereFields()
	pathInQuery := queryBuilder{
		filter: &Filter{
			AnyPath: paths,
		},
	}
	pathInQuery.whereFieldPathIn()
	pathIn := pathInQuery.where[len(pathInQuery.where)-1].eqContains[0]
	q.args = append(q.args, pathInQuery.args...)
	query.WriteString(fmt.Sprintf(`%s)
		WHERE time_on_page > 0 %s
		AND %s
		GROUP BY path`, whereTime, q.q.String(), pathIn))
	stats, err := pages.store.SelectAvgTimeSpentStats(filter.Ctx, query.String(), q.args...)

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
