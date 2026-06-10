package query

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

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
	joinTable      string
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
	q.resolveJoinTable(req)

	for _, filter := range req.Filter {
		if err := q.classifyFilter(filter); err != nil {
			return report.Report{
				Meta: report.Meta{
					Errors: []error{err},
				},
			}
		}
	}

	if q.joinTable != "" {
		return q.runWithJoin(req)
	}

	return q.run(req)
}

func (q *Query) runWithJoin(req request.Request) report.Report {
	requestMetrics := req.Metrics
	primaryMetrics, secondaryMetrics := q.splitMetrics(requestMetrics)
	req.Metrics = primaryMetrics
	primaryQuery, primaryArgs := q.buildQuery(req)
	req.Metrics = secondaryMetrics
	req.OrderBy = nil
	req.Pagination = nil
	q.primaryTable = q.joinTable
	secondaryQuery, secondaryArgs := q.buildQuery(req)
	var wg sync.WaitGroup
	var m sync.Mutex
	var results, secondaryResults []report.Result
	var err error
	wg.Go(func() {
		primaryRows, e := q.db.Query(req.Ctx, primaryQuery, primaryArgs...)

		if e != nil {
			m.Lock()
			defer m.Unlock()
			err = e
			return
		}

		results, e = q.scanRows(primaryRows, req.Dimensions, primaryMetrics)

		if e != nil {
			m.Lock()
			defer m.Unlock()
			err = e
		}
	})
	wg.Go(func() {
		secondaryRows, e := q.db.Query(req.Ctx, secondaryQuery, secondaryArgs...)

		if e != nil {
			m.Lock()
			defer m.Unlock()
			err = e
			return
		}

		secondaryResults, e = q.scanRows(secondaryRows, req.Dimensions, secondaryMetrics)

		if e != nil {
			m.Lock()
			defer m.Unlock()
			err = e
		}
	})
	wg.Wait()

	if err != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: []error{
					errors.New("error executing queries"),
					err,
				},
			},
		}
	}

	q.mergeResults(results, secondaryResults, requestMetrics, primaryMetrics)
	return report.Report{
		Request: req,
		Results: results,
	}
}

func (q *Query) run(req request.Request) report.Report {
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

	results, err := q.scanRows(rows, req.Dimensions, req.Metrics)

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

func (q *Query) resolveJoinTable(req request.Request) {
	for _, m := range req.Metrics {
		joinTable := m.JoinTable()

		if joinTable != "" && joinTable != q.primaryTable {
			q.joinTable = joinTable
			return
		}
	}
}

func (q *Query) splitMetrics(requestMetrics []metrics.Metric) ([]metrics.Metric, []metrics.Metric) {
	primary := make([]metrics.Metric, 0, len(requestMetrics))
	secondary := make([]metrics.Metric, 0)

	for _, m := range requestMetrics {
		if m.JoinTable() == q.joinTable {
			secondary = append(secondary, m)
		} else {
			primary = append(primary, m)
		}
	}

	return primary, secondary
}

func (q *Query) mergeResults(primary []report.Result, secondary []report.Result, requestMetrics, primaryMetrics []metrics.Metric) {
	totalCount := len(requestMetrics)
	primaryPos := make(map[string]int, len(primaryMetrics))

	for i, m := range primaryMetrics {
		primaryPos[m.Column()] = i
	}

	secondaryIndices := 0
	secondaryPos := make(map[string]int, totalCount-len(primaryMetrics))

	for _, m := range requestMetrics {
		if m.JoinTable() != "" {
			secondaryPos[m.Column()] = secondaryIndices
			secondaryIndices++
		}
	}

	for i := range primary {
		expanded := make([]any, totalCount)

		for fullIdx, m := range requestMetrics {
			if m.JoinTable() == "" {
				if pos, ok := primaryPos[m.Column()]; ok && pos < len(primary[i].MetricValues) {
					expanded[fullIdx] = primary[i].MetricValues[pos]
				}
			} else {
				expanded[fullIdx] = m.Zero()
			}
		}

		primary[i].MetricValues = expanded
	}

	index := make(map[string]int, len(primary))

	for i, r := range primary {
		index[q.dimensionKey(r.DimensionValues)] = i
	}

	for _, sec := range secondary {
		primaryIndex, ok := index[q.dimensionKey(sec.DimensionValues)]

		if !ok {
			// secondary row has no matching primary row (can happen with pagination), skip
			continue
		}

		for fullIndex, m := range requestMetrics {
			if m.JoinTable() == "" {
				continue
			}

			if pos, ok := secondaryPos[m.Column()]; ok && pos < len(sec.MetricValues) {
				primary[primaryIndex].MetricValues[fullIndex] = sec.MetricValues[pos]
			}
		}
	}
}

func (q *Query) dimensionKey(values []any) string {
	parts := make([]string, len(values))

	for i, v := range values {
		parts[i] = fmt.Sprint(v)
	}

	return strings.Join(parts, "|")
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
	withRequired := q.buildQueryWithRequired(req.Metrics)

	if withRequired {
		withQuery, withArgs := q.buildQueryWith(req)
		query.WriteString(withQuery)
		args = append(args, withArgs...)
	}

	selectQuery, selectArgs := q.buildQuerySelect(req)
	query.WriteString(selectQuery)
	args = append(args, selectArgs...)
	query.WriteString(q.buildQuereFrom(q.primaryTable))

	if withRequired {
		query.WriteString(q.buildQueryWithJoin(req))
	}

	whereQuery, whereArgs := q.buildQueryWhere(req)
	query.WriteString(whereQuery)
	args = append(args, whereArgs...)
	query.WriteString(q.buildQueryGroupBy(req.Dimensions))
	query.WriteString(q.buildOrderBy(req.OrderBy))
	query.WriteString(q.buildPagination(req.Pagination))
	return query.String(), args
}

func (q *Query) buildQueryWithRequired(list []metrics.Metric) bool {
	found := false

	for _, m := range list {
		if _, ok := m.(metrics.AvgTimeOnPage); ok {
			found = true
			break
		}
	}

	return found
}

func (q *Query) buildQueryWith(req request.Request) (string, []any) {
	var query strings.Builder
	args := make([]any, 0)
	whereQuery, whereArgs := q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period)
	query.WriteString(fmt.Sprintf(`WITH top AS (
			SELECT path,
				leadInFrame(duration_seconds) OVER (
					PARTITION BY visitor_id, session_id
					ORDER BY time ASC
					ROWS BETWEEN CURRENT ROW AND 1 FOLLOWING
				) time_on_page
			FROM page_view_v7
			%s
		),
		avgtop AS (
			SELECT path, avgIf(time_on_page, time_on_page > 0) avg_time_on_page
			FROM top
			GROUP BY path
		) `, whereQuery))
	args = append(args, whereArgs...)

	// TODO apply filters
	// TODO include dimensions to group by

	return query.String(), args
}

func (q *Query) buildQueryWithJoin(req request.Request) string {
	// use entry_path to join the time on page, exit_path isn't allowed (because it's always 0)
	if q.primaryTable == pkg.TableSessions {
		return "LEFT JOIN avgtop ON entry_path = avgtop.path "
	}

	return "LEFT JOIN avgtop ON path = avgtop.path "
}

func (q *Query) buildQuerySelect(req request.Request) (string, []any) {
	fields := make([]string, 0, len(req.Metrics)+len(req.Dimensions))
	args := make([]any, 0)

	for _, metric := range req.Metrics {
		expression, requiresSubquery := metric.Expression(q.primaryTable)

		if requiresSubquery {
			subquery, a := q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period)
			expression = fmt.Sprintf(expression, subquery)
			args = append(args, a...)
		}

		fields = append(fields, fmt.Sprintf("%s %s", expression, metric.Column()))
	}

	for _, dimension := range req.Dimensions {
		fields = append(fields, q.buildQuerySelectColumn(dimension))
		args = append(args, dimension.Args()...)
	}

	return fmt.Sprintf("SELECT %s ", strings.Join(fields, ",")), args
}

func (q *Query) buildQuerySelectColumn(dimension dimensions.Dimension) string {
	switch d := dimension.(type) {
	case dimensions.EventMeta:
		if d.Path != "" {
			return d.Select(q.buildQueryFilterJSONPath(d.Path))
		}

		return fmt.Sprintf("%s %s", dimension.Expression(), dimension.Column(q.primaryTable))
	default:
		if dimension.Expression() != "" {
			return fmt.Sprintf("%s %s", dimension.Expression(), dimension.Column(q.primaryTable))
		}

		return dimension.Column(q.primaryTable)
	}
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
	tz := "UTC"

	if period.Timezone != nil {
		tz = period.Timezone.String()
	}

	dateFunc := "toDate"

	if period.IncludeTime {
		dateFunc = "toDateTime"
	}

	var query strings.Builder
	args := make([]any, 0)
	query.WriteString(`WHERE site_id = ? `)
	args = append(args, siteID)

	if period.From.Equal(period.To) {
		query.WriteString(fmt.Sprintf(`AND %s("time", '%s') = %s(?, '%s') `, dateFunc, tz, dateFunc, tz))
		args = append(args, period.From)
	} else {
		query.WriteString(fmt.Sprintf(`AND %s("time", '%s') BETWEEN %s(?, '%s') AND %s(?, '%s') `, dateFunc, tz, dateFunc, tz, dateFunc, tz))
		args = append(args, period.From, period.To)
	}

	return query.String(), args
}

func (q *Query) buildQueryFilter(filter request.Filter) (string, []any) {
	// filter for a column
	if len(filter.Filter) == 0 {
		if len(filter.Values) == 0 {
			return "", nil
		}

		return q.buildQueryFilterColumn(filter)
	}

	// filter for a group
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

func (q *Query) buildQueryFilterColumn(filter request.Filter) (string, []any) {
	switch filter.Dimension.(type) {
	case dimensions.TagKey:
		switch filter.Operator {
		case request.OperatorIsNot:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					expression, _ := q.buildQuereWhereColumn(filter)
					group = append(group, fmt.Sprintf(`mapContainsKey(%s, ?) = 0`, expression))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			expression, _ := q.buildQuereWhereColumn(filter)
			return fmt.Sprintf(`mapContainsKey(%s, ?) = 0`, expression), filter.Values
		default:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					expression, _ := q.buildQuereWhereColumn(filter)
					group = append(group, fmt.Sprintf(`mapContainsKey(%s, ?)`, expression))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			expression, _ := q.buildQuereWhereColumn(filter)
			return fmt.Sprintf(`mapContainsKey(%s, ?)`, expression), filter.Values
		}
	case dimensions.EventMetaKey:
		switch filter.Operator {
		case request.OperatorIsNot:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for _, v := range filter.Values {
					expression, _ := q.buildQuereWhereColumn(filter)
					group = append(group, fmt.Sprintf(`isNull(%s%s)`, expression, q.buildQueryFilterJSONPath(v.(string))))
				}

				return strings.Join(group, " OR "), nil
			}

			expression, _ := q.buildQuereWhereColumn(filter)
			return fmt.Sprintf(`isNull(%s%s)`, expression, q.buildQueryFilterJSONPath(filter.Values[0].(string))), nil
		default:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for _, v := range filter.Values {
					expression, _ := q.buildQuereWhereColumn(filter)
					group = append(group, fmt.Sprintf(`isNotNull(%s%s)`, expression, q.buildQueryFilterJSONPath(v.(string))))
				}

				return strings.Join(group, " OR "), nil
			}

			expression, _ := q.buildQuereWhereColumn(filter)
			return fmt.Sprintf(`isNotNull(%s%s)`, expression, q.buildQueryFilterJSONPath(filter.Values[0].(string))), nil
		}
	default:
		switch filter.Operator {
		case request.OperatorIsNot:
			if len(filter.Values) > 1 {
				expression, args := q.buildQuereWhereColumn(filter)
				args = append(args, filter.Values)
				return fmt.Sprintf("%s NOT IN (?)", expression), args
			}

			expression, args := q.buildQuereWhereColumn(filter)
			args = append(args, filter.Values...)
			return fmt.Sprintf("%s != ?", expression), args
		case request.OperatorContains:
			values := q.buildQueryFilterILikeValues(filter.Values)

			if len(values) > 1 {
				expression, args := q.buildQuereWhereColumn(filter)
				args = append(args, values)
				return fmt.Sprintf("arrayExists(v -> ilike(%s, v), ?)", expression), args
			}

			expression, args := q.buildQuereWhereColumn(filter)
			args = append(args, values...)
			return fmt.Sprintf("%s ILIKE ?", expression), args
		case request.OperatorContainsNot:
			values := q.buildQueryFilterILikeValues(filter.Values)

			if len(values) > 1 {
				expression, args := q.buildQuereWhereColumn(filter)
				args = append(args, values)
				return fmt.Sprintf("arrayExists(v -> ilike(%s, v), ?) = 0", expression), args
			}

			expression, args := q.buildQuereWhereColumn(filter)
			args = append(args, values...)
			return fmt.Sprintf("%s NOT ILIKE ?", expression), args
		case request.OperatorMatches:
			if len(filter.Values) > 1 {
				expression, args := q.buildQuereWhereColumn(filter)
				args = append(args, filter.Values)
				return fmt.Sprintf("multiMatchAny(%s, ?)", expression), args
			}

			expression, args := q.buildQuereWhereColumn(filter)
			args = append(args, filter.Values...)
			return fmt.Sprintf("match(%s, ?)", expression), args
		case request.OperatorMatchesNot:
			if len(filter.Values) > 1 {
				expression, args := q.buildQuereWhereColumn(filter)
				args = append(args, filter.Values)
				return fmt.Sprintf("multiMatchAny(%s, ?) = 0", expression), args
			}

			expression, args := q.buildQuereWhereColumn(filter)
			args = append(args, filter.Values...)
			return fmt.Sprintf("match(%s, ?) = 0", expression), args
		default:
			if len(filter.Values) > 1 {
				expression, args := q.buildQuereWhereColumn(filter)
				args = append(args, filter.Values)
				return fmt.Sprintf("%s IN (?)", expression), args
			}

			expression, args := q.buildQuereWhereColumn(filter)
			args = append(args, filter.Values...)
			return fmt.Sprintf("%s = ?", expression), args
		}
	}
}

func (q *Query) buildQuereWhereColumn(filter request.Filter) (string, []any) {
	switch d := filter.Dimension.(type) {
	case dimensions.TagValue:
		return "tags[?]", []any{d.Key}
	case dimensions.EventMeta:
		return fmt.Sprintf("%s%s", d.Column(""), q.buildQueryFilterJSONPath(d.Path)), nil
	default:
		return d.Column(""), nil
	}
}

func (q *Query) buildQueryFilterJSONPath(value string) string {
	parts := strings.Split(value, ".")
	path := ""

	for _, p := range parts {
		if index, err := strconv.Atoi(p); err == nil {
			path += fmt.Sprintf("[%d]", index+1)
		} else {
			path += fmt.Sprintf(`."%s"`, p)
		}
	}

	return path
}

func (q *Query) buildQueryFilterILikeValues(filterValues []any) []any {
	values := make([]any, 0, len(filterValues))

	for _, v := range filterValues {
		values = append(values, fmt.Sprintf("%%%s%%", v))
	}

	return values
}

func (q *Query) buildQueryGroupBy(dimensions []dimensions.Dimension) string {
	if len(dimensions) > 0 {
		fields := make([]string, 0, len(dimensions))

		for _, dimension := range dimensions {
			column := dimension.Column(q.primaryTable)

			if column != "" {
				fields = append(fields, column)
			}
		}

		if len(fields) > 0 {
			return fmt.Sprintf("GROUP BY %s ", strings.Join(fields, ","))
		}
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
				fields = append(fields, fmt.Sprintf("%s %s", o.Dimension.Column(q.primaryTable), direction))
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

func (q *Query) scanRows(rows driver.Rows, dimensions []dimensions.Dimension, metrics []metrics.Metric) ([]report.Result, error) {
	metricCount := len(metrics)
	dimensionCount := len(dimensions)
	columnCount := metricCount + dimensionCount
	columns := make([]any, columnCount)
	columnPtrs := make([]any, columnCount)

	for i := range metricCount {
		columns[i] = metrics[i].ScanType()
		columnPtrs[i] = columns[i]
	}

	for i := range dimensionCount {
		columns[metricCount+i] = dimensions[i].ScanType()
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
