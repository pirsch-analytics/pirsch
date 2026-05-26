package query

import (
	"fmt"
	"slices"
	"strings"

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
	subqueryTable  string
	primaryFilter  []classifiedFilter
	subqueryFilter []classifiedFilter
	query          strings.Builder
	args           []any
}

// NewQuery returns a new Query for given database connection.
func NewQuery(db *db.ClickHouse) *Query {
	return &Query{
		db:             db,
		primaryFilter:  make([]classifiedFilter, 0),
		subqueryFilter: make([]classifiedFilter, 0),
		args:           make([]any, 0),
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

	q.buildQuery(req)

	// TODO run query

	return report.Report{
		Request: req,
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
	// leaf node
	if len(filter.Filter) == 0 {
		if filter.Dimension == nil {
			return "", nil
		}

		return filter.Dimension.Table()[0], nil
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

func (q *Query) buildQuery(req request.Request) {
	q.buildQuerySelect(req.Metrics, req.Dimensions)
	q.buildQuereFrom(q.primaryTable)
	q.buildQueryWhere(req)
	q.buildQueryGroupBy(req.Dimensions)
}

func (q *Query) buildQuerySelect(metrics []metrics.Metric, dimensions []dimensions.Dimension) {
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

	q.query.WriteString("SELECT ")
	q.query.WriteString(strings.Join(fields, ","))
	q.query.WriteString(" ")
}

func (q *Query) buildQuereFrom(table string) {
	q.query.WriteString(fmt.Sprintf("FROM %s ", table))
}

// TODO
func (q *Query) buildQueryWhere(req request.Request) {
	q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period)
	//q.query.WriteString(q.buildQueryFilter(q.primaryFilter, request.OperatorAnd, nil, nil))

	if len(q.subqueryFilter) > 0 {
		q.query.WriteString("AND (visitor_id, session_id) IN (SELECT visitor_id, session_id ")
		q.buildQuereFrom(q.subqueryTable)
		q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period)
		//q.query.WriteString(q.buildQueryFilter(q.subqueryFilter, request.OperatorAnd, nil, nil))
		q.query.WriteString(") ")
	}
}

func (q *Query) buildQueryWhereSiteAndPeriod(siteID uint64, period request.Period) {
	// TODO time zone, compare dates, time, ...
	q.query.WriteString(`WHERE site_id = ? AND toDate("time") BETWEEN toDate(?) AND toDate(?) `)
	q.args = append(q.args, siteID, period.From, period.To)
}

// TODO args
func (q *Query) buildQueryFilter(filter []request.Filter, operator request.Operator, dimension dimensions.Dimension, values []any) string {
	// filter on a field
	if len(filter) == 0 {
		if len(values) == 0 {
			return ""
		}

		switch operator {
		case request.OperatorIsNot:
			if len(values) > 1 {
				return fmt.Sprintf("%s NOT IN (?)", dimension.Column())
			}

			return fmt.Sprintf("%s != ?", dimension.Column())
		case request.OperatorContains:
			if len(values) > 1 {
				return fmt.Sprintf("arrayExists(v -> ilike(%s, v), ?)", dimension.Column())
			}

			return fmt.Sprintf("%s ILIKE ?", dimension.Column())
		case request.OperatorContainsNot:
			if len(values) > 1 {
				return fmt.Sprintf("arrayExists(v -> ilike(%s, v), ?) = 0", dimension.Column())
			}

			return fmt.Sprintf("%s NOT ILIKE ?", dimension.Column())
		case request.OperatorMatches:
			if len(values) > 1 {
				return fmt.Sprintf("multiMatchAny(%s, ?)", dimension.Column())
			}

			return fmt.Sprintf("match(%s, ?)", dimension.Column())
		case request.OperatorMatchesNot:
			if len(values) > 1 {
				return fmt.Sprintf("multiMatchAny(%s, ?) = 0", dimension.Column())
			}

			return fmt.Sprintf("match(%s, ?) = 0", dimension.Column())
		default:
			if len(values) > 1 {
				return fmt.Sprintf("%s IN (?)", dimension.Column())
			}

			return fmt.Sprintf("%s = ?", dimension.Column())
		}
	}

	// filter group
	groups := make([]string, 0, len(filter))

	for _, f := range filter {
		groups = append(groups, q.buildQueryFilter(f.Filter, f.Operator, f.Dimension, f.Values))
	}

	if operator == request.OperatorOr {
		return fmt.Sprintf("AND (%s) ", strings.Join(groups, " OR "))
	}

	return fmt.Sprintf("AND (%s) ", strings.Join(groups, " AND "))
}

func (q *Query) buildQueryGroupBy(dimensions []dimensions.Dimension) {
	if len(dimensions) > 0 {
		fields := make([]string, 0, len(dimensions))

		for _, dimension := range dimensions {
			fields = append(fields, dimension.Column())
		}

		q.query.WriteString("GROUP BY ")
		q.query.WriteString(strings.Join(fields, ","))
	}
}
