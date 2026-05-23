package query

import (
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

// Query queries results for a report.Report.
type Query struct {
	db             *db.ClickHouse
	primaryTable   string
	joinTable      string
	primaryFilter  []request.Filter
	subqueryFilter []request.Filter

	// TODO
	/*
		q     strings.Builder
		args  []any
	*/
}

// NewQuery returns a new Query for given database connection.
func NewQuery(db *db.ClickHouse) *Query {
	return &Query{
		db:             db,
		primaryFilter:  make([]request.Filter, 0),
		subqueryFilter: make([]request.Filter, 0),
	}
}

// Run runs given request.Request and returns the report.Report.
func (q *Query) Run(request request.Request) report.Report {
	if errs := request.Validate(); errs != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: errs,
			},
		}
	}

	// TODO build query and generate report
	q.resolvePrimaryTable(request)

	/*for _, m := range request.Metrics {
		if m.Table() != q.primaryTable {
			q.joinTable = m.Table()
			break
		}
	}*/

	for _, filter := range request.Filter {
		q.classifyFilters(filter)
	}

	switch {
	case q.joinTable != "":
		//q.runWithJoin(req, route)
	case len(q.subqueryFilter) > 0:
		//q.runWithSubquery(req, route)
	default:
		//q.runSimple(req, route)
	}

	return report.Report{
		Request: request,
	}
}

func (q *Query) resolvePrimaryTable(req request.Request) {
	// dimensions drive the primary table
	if len(req.Dimensions) > 0 {
		q.primaryTable = q.highestPriorityTable(q.dimensionTables(req.Dimensions))
		return
	}

	// no dimensions, use metrics instead
	if len(req.Metrics) > 0 {
		q.primaryTable = q.highestPriorityTable(q.metricTables(req.Metrics))
		return
	}

	// fall back to sessions for simple queries
	q.primaryTable = pkg.TableSessions
}

func (q *Query) dimensionTables(dimensions []dimensions.Dimension) []string {
	tables := make([]string, 0, len(dimensions))

	for _, d := range dimensions {
		tables = append(tables, d.Table())
	}

	return tables
}

func (q *Query) metricTables(metrics []metrics.Metric) []string {
	tables := make([]string, 0, len(metrics))

	for _, m := range metrics {
		tables = append(tables, m.Table())
	}

	return tables
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

func (q *Query) classifyFilters(filter request.Filter) {
	// logical operators recurse into their children
	if len(filter.Filter) > 0 {
		for _, child := range filter.Filter {
			q.classifyFilters(child)
		}

		return
	}

	// leaf filter, classify by table
	if filter.Dimension == nil {
		return
	}

	if filter.Dimension.Table() == q.primaryTable {
		q.primaryFilter = append(q.primaryFilter, filter)
	} else {
		q.subqueryFilter = append(q.subqueryFilter, filter)
	}
}
