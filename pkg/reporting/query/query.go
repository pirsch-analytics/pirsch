package query

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/report"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
)

var tablePriority = map[string]int{
	pkg.TableSessions:  1,
	pkg.TablePageViews: 2,
	pkg.TableEvents:    2,
}

type classifiedFilter struct {
	table  string
	filter request.Filter
}

// Query queries results for a report.Report.
type Query struct {
	db             *db.ClickHouse
	primaryTable   string
	primaryFilter  []classifiedFilter
	subqueryFilter []classifiedFilter
}

// NewQuery returns a new Query for given database connection.
func NewQuery(db *db.ClickHouse) *Query {
	return &Query{
		db:             db,
		primaryFilter:  make([]classifiedFilter, 0),
		subqueryFilter: make([]classifiedFilter, 0),
	}
}

// Run runs given request.Request and returns the report.Report.
func (q *Query) Run(req request.Request) report.Report {
	if errs := req.Validate(); errs != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: errs,
			},
		}
	}

	q.resolvePrimaryTable(req)

	for _, filter := range req.Filter {
		if err := q.classifyFilter(filter); err != nil {
			return report.Report{
				Meta: report.Meta{
					Errors: []error{err},
				},
			}
		}
	}

	query, args := q.buildQuery(req)
	rows, err := q.db.Query(req.Ctx, query, args...)

	if err != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: []error{
					errors.New("error executing query"),
					err,
				},
			},
		}
	}

	results, err := q.scanRows(rows, req)

	if err != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: []error{
					errors.New("error scanning rows"),
					err,
				},
			},
		}
	}

	return report.Report{
		Request: req,
		Results: results,
	}
}

func (q *Query) resolvePrimaryTable(req request.Request) {
	// dimensions drive the primary table
	if len(req.Dimensions) > 0 {
		q.primaryTable = q.resolveBestTable(q.dimensionTables(req.Dimensions))
		return
	}

	// no dimensions, use metrics instead
	if len(req.Metrics) > 0 {
		q.primaryTable = q.resolveBestTable(q.metricTables(req.Metrics))
		return
	}

	// fall back to sessions for simple queries
	q.primaryTable = pkg.TableSessions
}

func (q *Query) resolveBestTable(tableSets [][]string) string {
	candidates := []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
	valid := make([]string, 0, len(candidates))

	for _, candidate := range candidates {
		satisfiesAll := true

		for _, tables := range tableSets {
			if !slices.Contains(tables, candidate) {
				satisfiesAll = false
				break
			}
		}

		if satisfiesAll {
			valid = append(valid, candidate)
		}
	}

	if len(valid) == 0 {
		// no single table works, use highest priority preferred table
		preferred := make([]string, 0, len(tableSets))

		for _, tables := range tableSets {
			if len(tables) > 0 {
				preferred = append(preferred, tables[0])
			}
		}

		return q.highestPriorityTable(preferred)
	}

	if len(valid) == 1 {
		return valid[0]
	}

	// multiple tables satisfy all dimensions, prefer the one most dimensions
	// list as their first choice (preferred table)
	score := make(map[string]int, len(valid))

	for _, tables := range tableSets {
		if len(tables) > 0 && slices.Contains(valid, tables[0]) {
			score[tables[0]]++
		}
	}

	best := valid[0]

	for _, t := range valid[1:] {
		if score[t] > score[best] {
			best = t
		}
	}

	return best
}

func (q *Query) highestPriorityTable(tables []string) string {
	table := pkg.TableSessions
	priority := 0

	for _, t := range tables {
		if p := tablePriority[t]; p > priority {
			table = t
			priority = p
		}
	}

	return table
}

func (q *Query) dimensionTables(dimensions []dimensions.Dimension) [][]string {
	tables := make([][]string, 0, len(dimensions))

	for _, d := range dimensions {
		tables = append(tables, d.Table())
	}

	return tables
}

func (q *Query) metricTables(metrics []metrics.Metric) [][]string {
	tables := make([][]string, 0, len(metrics))

	for _, m := range metrics {
		tables = append(tables, m.Table())
	}

	return tables
}

func (q *Query) classifyFilter(filter request.Filter) error {
	table, err := q.filterTable(filter)

	if err != nil {
		return err
	}

	if table == q.primaryTable || table == "" {
		q.primaryFilter = append(q.primaryFilter, classifiedFilter{
			table:  q.primaryTable,
			filter: filter,
		})
	} else {
		q.subqueryFilter = append(q.subqueryFilter, classifiedFilter{
			table:  table,
			filter: filter,
		})
	}

	return nil
}

func (q *Query) filterTable(filter request.Filter) (string, error) {
	// leaf node, dimension must be set
	if len(filter.Filter) == 0 {
		if filter.Dimension == nil {
			return "", nil
		}

		tables := filter.Dimension.Table()

		if slices.Contains(tables, q.primaryTable) {
			return q.primaryTable, nil
		}

		return tables[0], nil
	}

	// logical group, collect tables from all children
	childTables := make(map[string]bool)

	for _, child := range filter.Filter {
		table, err := q.filterTable(child)

		if err != nil {
			return "", err
		}

		if table != "" {
			childTables[table] = true
		}
	}

	if len(childTables) == 0 {
		return "", nil
	}

	// multiple tables in one logical group is only safe for AND at the top level
	if len(childTables) > 1 {
		// check if children are compatible with a single shared table
		sharedTable := q.findSharedTable(filter.Filter)

		if sharedTable != "" {
			return sharedTable, nil
		}

		if filter.Operator == request.OperatorOr || filter.Operator == request.OperatorNot {
			tables := make([]string, 0, len(childTables))

			for t := range childTables {
				tables = append(tables, t)
			}

			return "", fmt.Errorf("cannot mix dimensions from different tables (%s) within an OR/NOT filter group", strings.Join(tables, ", "))
		}

		return "", nil
	}

	// all leaves on the same table
	for t := range childTables {
		return t, nil
	}

	return "", nil
}

func (q *Query) findSharedTable(filter []request.Filter) string {
	// collect all table sets from leaf dimensions in the group
	tableSets := q.collectLeafTableSets(filter)

	if len(tableSets) == 0 {
		return ""
	}

	// try primary table first
	if q.allCompatible(tableSets, q.primaryTable) {
		return q.primaryTable
	}

	// try other tables in priority order
	for _, candidate := range []string{pkg.TablePageViews, pkg.TableEvents, pkg.TableSessions} {
		if candidate != q.primaryTable && q.allCompatible(tableSets, candidate) {
			return candidate
		}
	}

	return ""
}

func (q *Query) collectLeafTableSets(filter []request.Filter) [][]string {
	result := make([][]string, 0)

	for _, f := range filter {
		if len(f.Filter) == 0 {
			if f.Dimension != nil {
				result = append(result, f.Dimension.Table())
			}
		} else {
			result = append(result, q.collectLeafTableSets(f.Filter)...)
		}
	}

	return result
}

func (q *Query) allCompatible(tableSets [][]string, candidate string) bool {
	for _, tables := range tableSets {
		if !slices.Contains(tables, candidate) {
			return false
		}
	}

	return true
}

func (q *Query) buildQuery(req request.Request) (string, []any) {
	var query strings.Builder
	args := make([]any, 0)
	query.WriteString(q.buildQuerySelect(req.Metrics, req.Dimensions))
	query.WriteString(q.buildQuereFrom(q.primaryTable))
	whereQuery, whereArgs := q.buildQueryWhere(req)
	query.WriteString(whereQuery)
	args = append(args, whereArgs...)
	query.WriteString(q.buildQueryGroupBy(req.Dimensions))
	query.WriteString(q.buildOrderBy(req.OrderBy))
	query.WriteString(q.buildPagination(req.Pagination))
	return query.String(), args
}

func (q *Query) buildQuerySelect(metrics []metrics.Metric, dimensions []dimensions.Dimension) string {
	fields := make([]string, 0, len(metrics)+len(dimensions))

	for _, metric := range metrics {
		fields = append(fields, fmt.Sprintf("%s %s", metric.Expression(q.primaryTable), metric.Column()))
	}

	for _, dimension := range dimensions {
		if dimension.Expression() != "" {
			fields = append(fields, fmt.Sprintf("%s %s", dimension.Expression(), dimension.Column()))
		} else {
			fields = append(fields, dimension.Column())
		}
	}

	return fmt.Sprintf("SELECT %s ", strings.Join(fields, ","))
}

func (q *Query) buildQuereFrom(table string) string {
	return fmt.Sprintf("FROM %s ", table)
}

func (q *Query) buildQueryWhere(req request.Request) (string, []any) {
	var query strings.Builder
	args := make([]any, 0)
	whereQuery, whereArgs := q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period)
	query.WriteString(whereQuery)
	args = append(args, whereArgs...)

	if len(q.primaryFilter) > 0 {
		for _, filter := range q.primaryFilter {
			query.WriteString("AND (")
			where, a := q.buildQueryFilter(filter.filter)
			query.WriteString(where)
			args = append(args, a...)
			query.WriteString(") ")
		}
	}

	if len(q.subqueryFilter) > 0 {
		query.WriteString("AND (visitor_id, session_id) IN (SELECT visitor_id, session_id ")
		query.WriteString(q.buildQuereFrom(q.subqueryFilter[0].table))
		whereQuery, whereArgs = q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period)
		query.WriteString(whereQuery)
		args = append(args, whereArgs...)

		for _, filter := range q.subqueryFilter {
			query.WriteString("AND (")
			where, a := q.buildQueryFilter(filter.filter)
			query.WriteString(where)
			args = append(args, a...)
			query.WriteString(") ")
		}

		query.WriteString(") ")
	}

	return query.String(), args
}

func (q *Query) buildQueryWhereSiteAndPeriod(siteID uint64, period request.Period) (string, []any) {
	// TODO time zone, compare dates, time, ...
	return `WHERE site_id = ? AND toDate("time") BETWEEN toDate(?) AND toDate(?) `, []any{siteID, period.From, period.To}
}

func (q *Query) buildQueryFilter(filter request.Filter) (string, []any) {
	// filter field
	if len(filter.Filter) == 0 {
		if len(filter.Values) == 0 {
			return "", nil
		}

		switch filter.Operator {
		case request.OperatorIsNot:
			if len(filter.Values) > 1 {
				return fmt.Sprintf("%s NOT IN (?)", filter.Dimension.Column()), []any{filter.Values}
			}

			return fmt.Sprintf("%s != ?", filter.Dimension.Column()), filter.Values
		case request.OperatorContains:
			if len(filter.Values) > 1 {
				return fmt.Sprintf("arrayExists(v -> ilike(%s, v), ?)", filter.Dimension.Column()), []any{filter.Values}
			}

			return fmt.Sprintf("%s ILIKE ?", filter.Dimension.Column()), filter.Values
		case request.OperatorContainsNot:
			if len(filter.Values) > 1 {
				return fmt.Sprintf("arrayExists(v -> ilike(%s, v), ?) = 0", filter.Dimension.Column()), []any{filter.Values}
			}

			return fmt.Sprintf("%s NOT ILIKE ?", filter.Dimension.Column()), filter.Values
		case request.OperatorMatches:
			if len(filter.Values) > 1 {
				return fmt.Sprintf("multiMatchAny(%s, ?)", filter.Dimension.Column()), []any{filter.Values}
			}

			return fmt.Sprintf("match(%s, ?)", filter.Dimension.Column()), filter.Values
		case request.OperatorMatchesNot:
			if len(filter.Values) > 1 {
				return fmt.Sprintf("multiMatchAny(%s, ?) = 0", filter.Dimension.Column()), []any{filter.Values}
			}

			return fmt.Sprintf("match(%s, ?) = 0", filter.Dimension.Column()), filter.Values
		default:
			if len(filter.Values) > 1 {
				return fmt.Sprintf("%s IN (?)", filter.Dimension.Column()), []any{filter.Values}
			}

			return fmt.Sprintf("%s = ?", filter.Dimension.Column()), filter.Values
		}
	}

	// filter group
	groups := make([]string, 0, len(filter.Filter))
	args := make([]any, 0)

	for _, f := range filter.Filter {
		group, groupArgs := q.buildQueryFilter(f)
		groups = append(groups, group)
		args = append(args, groupArgs...)
	}

	if filter.Operator == request.OperatorOr {
		return strings.Join(groups, " OR "), args
	}

	return strings.Join(groups, " AND "), args
}

func (q *Query) buildQueryGroupBy(dimensions []dimensions.Dimension) string {
	if len(dimensions) > 0 {
		fields := make([]string, 0, len(dimensions))

		for _, dimension := range dimensions {
			fields = append(fields, dimension.Column())
		}

		return fmt.Sprintf("GROUP BY %s ", strings.Join(fields, ","))
	}

	return ""
}

func (q *Query) buildOrderBy(order []request.OrderBy) string {
	if len(order) > 0 {
		fields := make([]string, 0, len(order))

		for _, o := range order {
			direction := o.Direction

			if direction == "" {
				direction = request.DirectionDESC
			}

			if o.Dimension != nil {
				fields = append(fields, fmt.Sprintf("%s %s", o.Dimension.Column(), direction))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", o.Metric.Column(), direction))
			}
		}

		return fmt.Sprintf("ORDER BY %s ", strings.Join(fields, ","))
	}

	return ""
}

func (q *Query) buildPagination(pagination *request.Pagination) string {
	if pagination != nil && pagination.Limit > 0 {
		if pagination.Offset > 0 {
			return fmt.Sprintf("LIMIT %d, %d", pagination.Offset, pagination.Limit)
		}

		return fmt.Sprintf("LIMIT %d", pagination.Limit)
	}

	return ""
}

func (q *Query) scanRows(rows driver.Rows, req request.Request) ([]report.Result, error) {
	metricCount := len(req.Metrics)
	dimensionCount := len(req.Dimensions)
	columnCount := metricCount + dimensionCount
	columns := make([]any, columnCount)
	columnPtrs := make([]any, columnCount)

	for i := range metricCount {
		columns[i] = req.Metrics[i].ScanType()
		columnPtrs[i] = columns[i]
	}

	for i := range dimensionCount {
		columns[metricCount+i] = req.Dimensions[i].ScanType()
		columnPtrs[metricCount+i] = columns[metricCount+i]
	}

	results := make([]report.Result, 0)

	for rows.Next() {
		if err := rows.Scan(columnPtrs...); err != nil {
			return nil, err
		}

		result := report.Result{
			MetricValues:    make([]any, metricCount),
			DimensionValues: make([]any, dimensionCount),
		}

		for i := range metricCount {
			result.MetricValues[i] = reflect.ValueOf(columnPtrs[i]).Elem().Interface()
		}

		for i := range dimensionCount {
			result.DimensionValues[i] = reflect.ValueOf(columnPtrs[metricCount+i]).Elem().Interface()
		}

		results = append(results, result)
	}

	return results, rows.Err()
}
